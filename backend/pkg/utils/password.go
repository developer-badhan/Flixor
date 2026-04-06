package utils

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// bcryptCost controls how slow the hashing is.
const bcryptCost = 12

/**
 * HashPassword takes a plain text password and returns a bcrypt hash.
 * The hash is safe to store in MongoDB — it cannot be reversed.
 * Returns an error if hashing fails (extremely rare — only on system-level issues).
*/
func HashPassword(plain string) (string, error) {
	if plain == "" {
		return "", errors.New("password cannot be empty")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", errors.New("failed to hash password")
	}

	return string(hashed), nil
}

/**
 * CheckPassword compares a plain text password against a stored bcrypt hash.
 * Returns nil if they match, an error if they do not.
 * Never compare passwords with == — always use this function.
 * Timing-safe: takes the same time whether the password is wrong
 * by one character or completely different — prevents timing attacks.
*/
func CheckPassword(plain, hashed string) error {
	if plain == "" || hashed == "" {
		return errors.New("password and hash cannot be empty")
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
	if err != nil {
		return errors.New("invalid credentials")
	}

	return nil
}