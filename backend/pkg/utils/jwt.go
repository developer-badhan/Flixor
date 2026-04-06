package utils

import (
	"errors"
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

/**
 * GenerateToken creates a signed JWT for a successfully authenticated user.
 * The token is returned as a string — store nothing on the server side.
 * JWTs are stateless: the token itself carries all the information we need.
*/
func GenerateToken(userID, email, secret string, expiryHours int) (string, error) {
	if secret == "" {
		return "", errors.New("JWT secret cannot be empty")
	}

	// Build the claims — this is the payload of the token
	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			// ExpiresAt is enforced automatically by the jwt library on validation.
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiryHours) * time.Hour)),

			// IssuedAt lets us know when the token was created — useful for audit logs
			IssuedAt: jwt.NewNumericDate(time.Now()),

			// Issuer identifies who created this token.
			Issuer: "flixor-api",
		},
	}

	/**
	 * Create the token using HMAC-SHA256 (HS256).
	 * HS256 is symmetric — same secret signs and verifies.
	 * Suitable for a single backend. If you ever have multiple services
	*/
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret — this produces the final string
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", errors.New("failed to sign token")
	}

	return signed, nil
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