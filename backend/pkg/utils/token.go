package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

/**
 * AccessTokenClaims are the JWT payload fields.
 * We keep it minimal — large payloads slow down every request.
*/
type AccessTokenClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

/**
 * GenerateAccessToken creates a signed JWT that expires in 15 minutes.
 * Short TTL limits the damage window if a token is stolen.
*/
func GenerateAccessToken(userID, email string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", fmt.Errorf("JWT_SECRET not set")
	}

	/**
	 * Create a new JWT token with the given user ID and email.
	 * The token is signed with the JWT_SECRET.
	*/
	claims := AccessTokenClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "flixor",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

/**
 * ValidateAccessToken parses and validates a JWT string.
 * Returns the claims on success, or a descriptive error on failure.
*/
func ValidateAccessToken(tokenStr string) (*AccessTokenClaims, error) {
	secret := os.Getenv("JWT_SECRET")

	token, err := jwt.ParseWithClaims(tokenStr, &AccessTokenClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*AccessTokenClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// Refresh Token TTL
const refreshTokenTTL = 7 * 24 * time.Hour // 7 days

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
	expiresAt = time.Now().Add(refreshTokenTTL)
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
	return int(refreshTokenTTL.Seconds())
}

// AccessTokenTTLSeconds is the access token TTL in seconds (for TokenResponse).
const AccessTokenTTLSeconds = 15 * 60 // 900 seconds