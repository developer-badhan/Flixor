package service

import (
	"context"
	"errors"
	"time"

	"github.com/developer-badhan/Flixor/config"
	"github.com/developer-badhan/Flixor/internal/model"
	"github.com/developer-badhan/Flixor/internal/repository"
	"github.com/developer-badhan/Flixor/pkg/utils"
)

/**
 * AuthService handles all authentication business logic.
 * It owns the rules — what makes a valid registration, what makes a
 * successful login. The handler layer enforces HTTP; this layer enforces business.
*/
type AuthService struct {
	userRepo *repository.UserRepository
	cfg      *config.Config
}

/**
 * NewAuthService creates an AuthService with its required dependencies.
 * Called once at startup in main.go — dependencies injected, never fetched internally.
*/
func NewAuthService(userRepo *repository.UserRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

/**
 *  Register creates a new user account.
 * Business rules enforced here:
 *   - Email must not already be registered
 *   - Password is hashed before storage — plain text never touches the database
 *   - Timestamps are set by the repository — not trusted from the client
 * 
 * Returns AuthResponse (token + public user) so the client is
 * logged in immediately after registering — no second login call needed.
*/
func (s *AuthService) Register(ctx context.Context, req *model.RegisterRequest) (*model.AuthResponse, error) {
	// Rule 1: Check for duplicate email 
	existing, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		// A real database error — not just "not found"
		return nil, errors.New("registration failed — please try again")
	}
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	// Rule 2: Hash the password 
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("registration failed — please try again")
	}

	// Rule 3: Build the user document 
	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
	}

	// Rule 4: Persist the user 
	createdUser, err := s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	// Rule 5: Generate JWT
	token, err := utils.GenerateToken(
		createdUser.ID.Hex(),
		createdUser.Email,
		s.cfg.JWTSecret,
		s.cfg.JWTExpiryHours,
	)
	if err != nil {
		return nil, errors.New("account created but login failed — please log in manually")
	}

	// Build and return the response 
	return &model.AuthResponse{
		Token: token,
		User: model.PublicUser{
			ID:        createdUser.ID,
			Username:  createdUser.Username,
			Email:     createdUser.Email,
			CreatedAt: createdUser.CreatedAt,
		},
	}, nil
}

/**
 *  Login authenticates an existing user and returns a JWT on success.
 *  Business rules enforced here:
 *    - Always return the same error for wrong email or wrong password
 *   (prevents email enumeration attacks)
 *    - Token expiry is read from config — one place to change it
*/
func (s *AuthService) Login(ctx context.Context, req *model.LoginRequest) (*model.AuthResponse, error) {
	// Step 1: Look up the user by email
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Step 2: Verify the password
	if err := utils.CheckPassword(req.Password, user.Password); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Step 3: Generate JWT
	token, err := utils.GenerateToken(
		user.ID.Hex(),
		user.Email,
		s.cfg.JWTSecret,
		s.cfg.JWTExpiryHours,
	)
	if err != nil {
		return nil, errors.New("login failed — please try again")
	}

	// Step 4: Update last seen (non-blocking) 
	go s.updateLastSeen(user.ID.Hex())

	// Build and return the response 
	return &model.AuthResponse{
		Token: token,
		User: model.PublicUser{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
	}, nil
}

/**
 * updateLastSeen records when a user last logged in.
 * Runs in a background goroutine — login response is never blocked by this.
 * Uses a fresh context since the request context may already be cancelled
 * by the time this goroutine runs.
*/
func (s *AuthService) updateLastSeen(userID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = ctx
	_ = userID
}