package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
)

/**
 * GenerateOTP creates a cryptographically random 6-digit string.
 * Uses crypto/rand — not math/rand — because predictable OTPs are exploitable.
 * Zero-padded so "000123" is a valid OTP.
*/
func GenerateOTP() (string, error) {
	// Upper bound is 1_000_000 so the result spans 000000–999999
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", fmt.Errorf("failed to generate OTP: %w", err)
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

/**
 * HashOTP returns the SHA-256 hex digest of a plain OTP string.
 * Only this hash is stored in MongoDB — the plain OTP is only ever sent by email.
*/
func HashOTP(otp string) string {
	sum := sha256.Sum256([]byte(otp))
	return hex.EncodeToString(sum[:])
}