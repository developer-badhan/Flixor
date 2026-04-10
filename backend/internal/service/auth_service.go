package service

import (
	"context"
	"errors"
	"fmt"
	"time"
	"io"
	"mime/multipart"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"

	"github.com/developer-badhan/Flixor/config"
	"github.com/developer-badhan/Flixor/internal/model"
	"github.com/developer-badhan/Flixor/internal/repository"
	"github.com/developer-badhan/Flixor/pkg/utils"
	"github.com/developer-badhan/Flixor/pkg/cloudinary"
	"github.com/developer-badhan/Flixor/pkg/email"
)

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
 * Sentinel errors — handlers use errors.Is() against these,
 * never string-match on error messages.
*/
var (
	ErrAlreadyVerified   = errors.New("account is already verified")
	ErrOTPNotFound       = errors.New("no OTP pending for this account — request a new one")
	ErrOTPExpired        = errors.New("OTP has expired — request a new one")
	ErrOTPInvalid        = errors.New("invalid OTP")
	ErrNoFieldsToUpdate  = errors.New("no valid fields to update")
	ErrInvalidFileType   = errors.New("only jpeg, png, and webp images are accepted")
	ErrFileTooLarge      = errors.New("image must be smaller than 5 MB")
)

// otpTTL is the window a user has to enter their OTP after requesting it.
const otpTTL = 10 * time.Minute

// maxProfilePictureBytes is 5 MB — enforced before any upload to Cloudinary.
const maxProfilePictureBytes = 5 * 1024 * 1024

// allowedImageTypes is the MIME type allowlist for profile pictures.
var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

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
	userRepo    repository.UserRepository
	refreshRepo repository.RefreshTokenRepository
	cfg         *config.Config
}

/**
 * UserService handles all user-profile concerns:
 * - reading profile, updating profile, picture upload, OTP send/verify.
 * - Auth (register/login/tokens) stays in AuthService — this service does not touch tokens.
*/
type UserService struct {
	userRepo   repository.UserRepository
	cloudinary *cloudinary.Client
	mailer     *email.Mailer
}

/**
 * Constructor — inject dependencies once:
 * - UserRepository
 * - RefreshTokenRepository
 * - Config
*/
func NewAuthService(
	userRepo repository.UserRepository,
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
 * Constructor — inject dependencies once:
 * - UserRepository
 * - Cloudinary client
 * - Email sender
*/
func NewUserService(
	userRepo repository.UserRepository,
	cldClient *cloudinary.Client,
	mailer *email.Mailer,
) *UserService {
	return &UserService{
		userRepo:   userRepo,
		cloudinary: cldClient,
		mailer:     mailer,
	}
}

/**
 * REGISTER (Upgraded to use token pair):
 * Registers a new user and returns a token pair.
*/
func (s *AuthService) Register(ctx context.Context, req *model.RegisterRequest, ip, userAgent string) (*model.TokenResponse, error) {

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
		ip,
		userAgent,
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
func (s *AuthService) Login(ctx context.Context, req *model.LoginRequest, ip, userAgent string) (*model.TokenResponse, error) {

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
		ip,
		userAgent,
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
	accessToken, err := utils.GenerateAccessToken(userID.Hex(), email, s.cfg.JWTSecret)
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

	// Retrieve the user's current email so the new access token carries
	// a fully populated claims payload — not an empty email string.
	user, err := s.userRepo.FindByID(ctx, record.UserID.Hex())
	if err != nil {
		return nil, fmt.Errorf("could not retrieve user for token refresh: %w", err)
	}	

	// Issue new token pair
	return s.IssueTokenPair(ctx, record.UserID, user.Email, ip, userAgent)
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
 * GetMe fetches the full user document and converts it to the safe public view.
 * This is what GET /api/v1/user/me returns.
*/
func (s *UserService) GetMe(ctx context.Context, userID string) (*model.PublicUser, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, err
		}
		return nil, errors.New("failed to retrieve profile")
	}
	return user.ToPublic(), nil
}

/**
 * UpdateProfile applies whitelisted field changes to the user document.
 * Email can never be changed here — it is silently ignored even if sent.
 * Password is hashed before being passed to the repository.
*/
func (s *UserService) UpdateProfile(
	ctx context.Context,
	userID string,
	req *model.UpdateProfileRequest,
) (*model.PublicUser, error) {

	updates := bson.M{}

	if req.Username != "" {
		updates["username"] = req.Username
	}

	if req.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, errors.New("failed to process new password")
		}
		updates["password"] = string(hashed)
	}

	/**
	 * Reject requests that contain zero actionable changes.
	 * Without this guard, an empty PATCH would still hit the database and
	 * stamp a new updated_at with no actual change — misleading and wasteful.
	*/
	if len(updates) == 0 {
		return nil, ErrNoFieldsToUpdate
	}

	updated, err := s.userRepo.UpdateProfile(ctx, userID, updates)
	if err != nil {
		return nil, err
	}

	return updated.ToPublic(), nil
}

/**
 * UploadProfilePicture validates the uploaded file, pushes it to Cloudinary,
 * then persists the returned HTTPS URL in the user document.
 *
 * Validation happens in the service — not the handler — because the handler's
 * job is to parse HTTP, not to know what constitutes a valid image.
*/
func (s *UserService) UploadProfilePicture(
	ctx context.Context,
	userID string,
	fileHeader *multipart.FileHeader,
) (*model.PublicUser, error) {

	// Guard 1: size — reject before opening the file
	if fileHeader.Size > maxProfilePictureBytes {
		return nil, ErrFileTooLarge
	}

	// Guard 2: MIME type from the Content-Type header the client sent
	contentType := fileHeader.Header.Get("Content-Type")
	if !allowedImageTypes[contentType] {
		return nil, ErrInvalidFileType
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, errors.New("failed to open uploaded file")
	}
	defer file.Close()

	// Cast to io.Reader — Cloudinary SDK accepts any io.Reader
	url, err := s.cloudinary.UploadProfilePicture(ctx, io.Reader(file), userID)
	if err != nil {
		return nil, errors.New("image upload failed — please try again")
	}

	updated, err := s.userRepo.UpdateProfile(ctx, userID, bson.M{"profile_picture": url})
	if err != nil {
		return nil, err
	}

	return updated.ToPublic(), nil
}

/**
 * SendOTP generates a 6-digit OTP, stores its SHA-256 hash in MongoDB with a
 * 10-minute TTL, and sends the plain OTP to the user's email address.
 *
 * Re-sending is allowed — each call overwrites the previous OTP hash and
 * resets the TTL. This lets users recover from a lost email without needing
 * a separate "resend" endpoint.
*/
func (s *UserService) SendOTP(ctx context.Context, userID string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return errors.New("failed to retrieve user")
	}

	if user.IsVerified {
		return ErrAlreadyVerified
	}

	otp, err := utils.GenerateOTP()
	if err != nil {
		return errors.New("failed to generate OTP — please try again")
	}

	otpHash := utils.HashOTP(otp)
	expiresAt := time.Now().Add(otpTTL)

	if err := s.userRepo.SaveOTP(ctx, userID, otpHash, expiresAt); err != nil {
		return errors.New("failed to save OTP — please try again")
	}

	// Send email asynchronously — the HTTP response should not block on SMTP.
	// If the email fails we log it; the user can request a new OTP.
	// The OTP hash is already stored, so a retry will send the same hash window.
	go func() {
		if err := s.mailer.SendOTP(user.Email, user.Username, otp); err != nil {
			// In production, replace this with your structured logger.
			// We do not return this error to the caller — the HTTP response
			// has already been sent by the time this goroutine runs.
			_ = err
		}
	}()

	return nil
}

/**
 * VerifyOTP compares the submitted plain OTP against the stored hash.
 * On success, marks the user as verified and clears the OTP fields atomically.
 * On failure, the stored OTP is NOT cleared — the user may retry until expiry.
*/
func (s *UserService) VerifyOTP(ctx context.Context, userID, plainOTP string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return errors.New("failed to retrieve user")
	}

	if user.IsVerified {
		return ErrAlreadyVerified
	}

	// No OTP stored — user never called send-otp or the TTL index cleared it.
	if user.OTPHash == "" {
		return ErrOTPNotFound
	}

	// Clock check before hash comparison — cheaper operation first.
	if time.Now().After(user.OTPExpiresAt) {
		return ErrOTPExpired
	}

	// Hash what the user submitted and compare — we never store plain OTPs.
	if utils.HashOTP(plainOTP) != user.OTPHash {
		return ErrOTPInvalid
	}

	// Atomic: set is_verified=true and unset otp_hash + otp_expires_at.
	return s.userRepo.MarkVerified(ctx, userID)
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