package auth

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword generates a bcrypt hash for the given password.
// It uses the default cost provided by the bcrypt package.
func HashPassword(password string) (string, error) {
	// GenerateFromPassword automatically handles salt generation.
	// bcrypt.DefaultCost is typically used, but you can specify others (4-31).
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		// Wrap the error for more context, although bcrypt errors are usually clear.
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

// CheckPasswordHash compares a plaintext password with a stored bcrypt hash.
// Returns true if the password matches the hash, false otherwise.
func CheckPasswordHash(password, hash string) bool {
	// CompareHashAndPassword handles comparing the provided password
	// with the stored hash, extracting the salt and cost from the hash itself.
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	// A nil error means the password matches the hash.
	// Any error (like ErrMismatchedHashAndPassword) means they don't match.
	return err == nil
}
