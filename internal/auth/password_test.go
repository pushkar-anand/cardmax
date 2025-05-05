package auth

import (
	"github.com/stretchr/testify/require"
	"testing"
)

const testPass = "UyG7f3A2wB5Jdq8d3B"

func TestPasswordHashing(t *testing.T) {
	t.Run("valid password", func(t *testing.T) {
		hash, err := HashPassword(testPass)
		require.NoError(t, err)
		require.NotEmpty(t, hash)

		err = CheckPasswordHash(testPass, hash)
		require.NoError(t, err)
	})

	t.Run("invalid password", func(t *testing.T) {
		hash, err := HashPassword(testPass)
		require.NoError(t, err)
		require.NotEmpty(t, hash)

		err = CheckPasswordHash("abcd", hash)
		require.ErrorIs(t, err, ErrMismatchedHashAndPassword)
	})

	t.Run("hash format", func(t *testing.T) {
		hash, err := HashPassword(testPass)
		require.NoError(t, err)
		t.Logf("Hash: %s", hash)

		// Test our regex can parse the hash
		matches := hashRegex.FindStringSubmatch(hash)
		require.NotNil(t, matches)
		require.Len(t, matches, 8)
	})
}
