package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

/**
 * Claims defines exactly what lives inside our JWT payload.
 * We embed jwt.RegisteredClaims to get standard fields like ExpiresAt
 * for free — then add our own application-specific fields on top.
*/
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// Access Token TTL
const AccessTokenTTLSeconds = 15 * 60 // 900 seconds

// Refresh Token TTL
const RefreshTokenTTL = 7 * 24 * time.Hour // 7 days

/**
 * GenerateAccessToken creates a signed JWT that expires in 15 minutes.
 * Short TTL limits the damage window if a token is stolen.
*/
func GenerateAccessToken(userID, email, secret string) (string, error) {
	if secret == "" {
		return "", fmt.Errorf("JWT_SECRET cannot be empty")
	}

	/**
	 * Create a new JWT token with the given user ID and email.
	 * The token is signed with the JWT_SECRET.
	*/
	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenTTLSeconds * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "flixor",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

/**
 * GenerateRefreshToken creates a cryptographically random 32-byte token
 * encoded as a 64-character hex string.
 * We use crypto/rand — not math/rand — because predictable tokens are exploitable.
*/
func GenerateRefreshToken() (rawToken string, hash string, expiresAt time.Time, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	rawToken = hex.EncodeToString(b)
	hash = HashToken(rawToken)
	expiresAt = time.Now().Add(RefreshTokenTTL)
	return rawToken, hash, expiresAt, nil
}

/**
 * HashToken returns the SHA-256 hex digest of a raw token string.
 * This is what gets stored in MongoDB — never the raw token.
*/
func HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// RefreshTokenTTLSeconds returns the TTL as an integer for API responses.
func RefreshTokenTTLSeconds() int {
	return int(RefreshTokenTTL.Seconds())
}

/**
 * ValidateToken parses a JWT string and returns the claims if valid.
 * Returns an error if the token is expired, tampered with, or malformed.
 * The middleware calls this on every protected request.
*/
func ValidateToken(tokenString, secret string) (*Claims, error) {
	if tokenString == "" {
		return nil, errors.New("token cannot be empty")
	}
	if secret == "" {
		return nil, errors.New("JWT secret cannot be empty")
	}

	/**
	 * Parse the token — the key function runs first to provide the signing key.
	 * Checking the method inside the key function is critical:
	 * if an attacker sends a token signed with "alg: none", we reject it
	 * before even attempting verification. This prevents the "alg:none" attack.
	*/
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			// Reject anything that isn't HMAC — never trust the header blindly
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(secret), nil
		},
	)

	if err != nil {
		/**
		 * jwt library returns specific errors for expiry vs tampering.
		 * We collapse them into two clean messages — enough info for the
		 * client to act, not enough to help an attacker diagnose failures.
		*/
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token has expired")
		}
		return nil, errors.New("token is invalid")
	}

	// Type-assert the claims back to our Claims struct
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("token is invalid")
	}

	return claims, nil
}
