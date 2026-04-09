package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/developer-badhan/Flixor/config"
	"github.com/developer-badhan/Flixor/internal/model"
	"github.com/developer-badhan/Flixor/internal/repository"
	"github.com/developer-badhan/Flixor/pkg/utils"
)

/**
 * AuthService handles ALL authentication logic:
 * - User registration & login
 * - Access + Refresh token issuing
 * - Token rotation
 * - Logout & session revocation
 *
 * Think of this as the "brain" of your auth system.
 */
type AuthService struct {
	userRepo    *repository.UserRepository
	refreshRepo repository.RefreshTokenRepository
	cfg         *config.Config
}

/**
 * Constructor — inject dependencies once
 */
func NewAuthService(
	userRepo *repository.UserRepository,
	refreshRepo repository.RefreshTokenRepository,
	cfg *config.Config,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		refreshRepo: refreshRepo,
		cfg:         cfg,
	}
}

/**
 * REGISTER (Upgraded to use token pair):
 * Registers a new user and returns a token pair.
*/
func (s *AuthService) Register(ctx context.Context, req *model.RegisterRequest) (*model.TokenResponse, error) {

	// Rule 1: Check duplicate email
	existing, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return nil, errors.New("registration failed — please try again")
	}
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	// Rule 2: Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("registration failed — please try again")
	}

	// Rule 3: Create user model
	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
	}

	// Rule 4: Save user
	createdUser, err := s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	// Rule 5: Issue token pair (NEW SYSTEM)
	tokenPair, err := s.IssueTokenPair(
		ctx,
		createdUser.ID,
		createdUser.Email,
		"", // IP (optional)
		"", // UserAgent (optional)
	)
	if err != nil {
		return nil, errors.New("account created but login failed — please log in manually")
	}

	return tokenPair, nil
}

/**
 * LOGIN (Upgraded to use token pair):
 * Logs in a user and returns a token pair.
*/
func (s *AuthService) Login(ctx context.Context, req *model.LoginRequest) (*model.TokenResponse, error) {

	// Step 1: Find user
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Step 2: Verify password
	if err := utils.CheckPassword(req.Password, user.Password); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Step 3: Issue token pair
	tokenPair, err := s.IssueTokenPair(
		ctx,
		user.ID,
		user.Email,
		"", // IP
		"", // UserAgent
	)
	if err != nil {
		return nil, errors.New("login failed — please try again")
	}

	// Step 4: Async last seen update
	go s.updateLastSeen(user.ID.Hex())

	return tokenPair, nil
}

/**
 * TOKEN ISSUING (Access + Refresh):
 * Issues a new pair of access and refresh tokens for a user.
*/
func (s *AuthService) IssueTokenPair(
	ctx context.Context,
	userID primitive.ObjectID,
	email, ip, userAgent string,
) (*model.TokenResponse, error) {

	// Access Token (short-lived)
	accessToken, err := utils.GenerateAccessToken(userID.Hex(), email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Refresh Token (secure random)
	rawRefresh, hash, expiresAt, err := utils.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store hash in DB
	record := &model.RefreshToken{
		UserID:      userID,
		TokenHash:   hash,
		Blacklisted: false,
		UserAgent:   userAgent,
		IP:          ip,
		ExpiresAt:   expiresAt,
	}

	if err := s.refreshRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &model.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		TokenType:    "Bearer",
		ExpiresIn:    utils.AccessTokenTTLSeconds,
	}, nil
}

/**
 * TOKEN ISSUING (Access + Refresh):
 * Issues a new pair of access and refresh tokens for a user.
*/
var (
	ErrTokenNotFound    = errors.New("refresh token not found")
	ErrTokenBlacklisted = errors.New("refresh token has been revoked")
	ErrTokenExpired     = errors.New("refresh token has expired")
	ErrTokenReuse       = errors.New("refresh token reuse detected — all sessions revoked")
)

/**
 * REFRESH TOKENS (Rotation + Security):
 * Refreshes the access and refresh tokens using the provided refresh token.
*/
func (s *AuthService) RefreshTokens(
	ctx context.Context,
	rawRefreshToken, ip, userAgent string,
) (*model.TokenResponse, error) {

	hash := utils.HashToken(rawRefreshToken)

	record, err := s.refreshRepo.FindByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrTokenNotFound
		}
		return nil, fmt.Errorf("db error: %w", err)
	}

	// Reuse attack detection
	if record.Blacklisted {
		_ = s.refreshRepo.BlacklistAllForUser(ctx, record.UserID)
		return nil, ErrTokenReuse
	}

	// Expiry check
	if time.Now().After(record.ExpiresAt) {
		return nil, ErrTokenExpired
	}

	// Blacklist old token
	if err := s.refreshRepo.BlacklistByHash(ctx, hash); err != nil {
		return nil, fmt.Errorf("failed to revoke token: %w", err)
	}

	// Issue new token pair
	return s.IssueTokenPair(ctx, record.UserID, "", ip, userAgent)
}

/**
 * LOGOUT (Single Device):
 * Revokes a single refresh token, logging the user out of one device.
*/
func (s *AuthService) Logout(ctx context.Context, rawRefreshToken string) error {

	hash := utils.HashToken(rawRefreshToken)

	record, err := s.refreshRepo.FindByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrTokenNotFound
		}
		return fmt.Errorf("db error: %w", err)
	}

	if record.Blacklisted {
		return nil // idempotent
	}

	return s.refreshRepo.BlacklistByHash(ctx, hash)
}

/**
 * LOGOUT ALL (All Devices):
 * Revokes all refresh tokens for a user, effectively logging them out of all devices.
*/
func (s *AuthService) LogoutAll(ctx context.Context, userID primitive.ObjectID) error {
	return s.refreshRepo.BlacklistAllForUser(ctx, userID)
}

/**
 * BACKGROUND TASK:
 * Updates the last seen timestamp for a user.
 * This is called asynchronously after a successful login.
*/
func (s *AuthService) updateLastSeen(userID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// TODO: Implement userRepo.UpdateLastSeen(userID)
	_ = ctx
	_ = userID
}