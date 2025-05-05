package auth

import (
	"testing"
	"golang.org/x/crypto/bcrypt"
)

// TestHashPassword tests the HashPassword function.
func TestHashPassword(t *testing.T) {
	password := "plainpassword123"
	hash, err := HashPassword(password)

	// 1. Check for errors during hashing
	if err != nil {
		t.Fatalf("HashPassword returned an unexpected error: %v", err)
	}

	// 2. Check if the hash is empty
	if hash == "" {
		t.Fatalf("HashPassword returned an empty hash string")
	}

	// 3. Verify the hash against the original password using bcrypt
	// This also indirectly checks if the returned string is a valid bcrypt hash.
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		t.Errorf("bcrypt.CompareHashAndPassword failed for the generated hash: %v", err)
	}

	// 4. Optional: Check if hashing the same password again yields a different hash (due to salt)
	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword (second call) returned an unexpected error: %v", err)
	}
	if hash == hash2 {
		t.Errorf("Hashing the same password twice produced the same hash, indicating a potential salt issue")
	}
}

// TestCheckPasswordHash tests the CheckPasswordHash function.
func TestCheckPasswordHash(t *testing.T) {
	password := "correctpassword"
	wrongPassword := "wrongpassword"

	// Generate a hash for the correct password
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Setup failed: HashPassword returned an error: %v", err)
	}

	testCases := []struct {
		name           string
		passwordToCheck string
		hashToCheck    string
		expectedResult bool
	}{
		{
			name:           "Correct password",
			passwordToCheck: password,
			hashToCheck:    hash,
			expectedResult: true,
		},
		{
			name:           "Incorrect password",
			passwordToCheck: wrongPassword,
			hashToCheck:    hash,
			expectedResult: false,
		},
		{
			name:           "Empty password",
			passwordToCheck: "",
			hashToCheck:    hash,
			expectedResult: false,
		},
		{
			name:           "Correct password, empty hash",
			passwordToCheck: password,
			hashToCheck:    "",
			expectedResult: false, // bcrypt compare will fail
		},
		{
			name:           "Correct password, invalid hash",
			passwordToCheck: password,
			hashToCheck:    "not-a-valid-bcrypt-hash",
			expectedResult: false, // bcrypt compare will fail
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CheckPasswordHash(tc.passwordToCheck, tc.hashToCheck)
			if result != tc.expectedResult {
				t.Errorf("CheckPasswordHash(%q, %q) = %v; want %v", tc.passwordToCheck, tc.hashToCheck, result, tc.expectedResult)
			}
		})
	}
}
