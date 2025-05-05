package auth

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/crypto/argon2"
	"regexp"
	"strconv"
)

const (
	pepper      = "d9BfPN6mv9rvNJTj"
	version     = argon2.Version // Argon2 version
	algoTime    = 4
	memory      = 64 * 1024
	saltLength  = 16
	parallelism = 2
	keyLength   = 32
)

var (
	ErrPasswordTooLong           = errors.New("password length exceeds RFC 9106 limits")
	ErrInvalidHashFormat         = errors.New("invalid hash format")
	ErrInvalidHashVersion        = errors.New("invalid hash version")
	ErrInvalidTimeParm           = errors.New("invalid time parameter")
	ErrInvalidMemParm            = errors.New("invalid memory parameter")
	ErrInvalidThreadParm         = errors.New("invalid thread parameter")
	ErrMismatchedHashAndPassword = errors.New("mismatched hash and password")
)

const MaxRFC9106PasswdLen int = (2 << 31) - 1

// Format strings for hash encoding/decoding
const (
	hashFormat = "$argon2id$v=%d$m=%d,t=%d,p=%d,k=%d$%s$%s"
)

// Regular expression to parse the hash
var hashRegex = regexp.MustCompile(`^\$argon2id\$v=(\d+)\$m=(\d+),t=(\d+),p=(\d+),k=(\d+)\$([^$]+)\$([^$]+)$`)

// HashPassword generates an Argon2id hash for the given password
func HashPassword(password string) (string, error) {
	if len(password) > MaxRFC9106PasswdLen {
		return "", ErrPasswordTooLong
	}

	saltBuf, err := generateRandomBytes(saltLength)
	if err != nil {
		return "", fmt.Errorf("could not generate salt: %w", err)
	}

	saltBuf = append(saltBuf, []byte(pepper)...)

	// Generate the hash using Argon2id
	hashBuf := argon2.IDKey([]byte(password), saltBuf, algoTime, memory, parallelism, keyLength)

	// Encode salt and hash to base64
	salt := base64.StdEncoding.EncodeToString(saltBuf)
	defer eraseBuf([]byte(salt))

	hash := base64.StdEncoding.EncodeToString(hashBuf)
	defer eraseBuf([]byte(hash))

	// Format the hash string
	return fmt.Sprintf(hashFormat, version, memory, algoTime, parallelism, keyLength, salt, hash), nil
}

// CheckPasswordHash compares a plaintext password with a stored hash
func CheckPasswordHash(password, encodedHash string) error {
	// Parse the hash
	matches := hashRegex.FindStringSubmatch(encodedHash)
	if matches == nil || len(matches) != 8 {
		return ErrInvalidHashFormat
	}

	// Extract parameters
	_, err := strconv.ParseInt(matches[1], 10, 32)
	if err != nil {
		return errors.Join(ErrInvalidHashVersion, err)
	}

	mem, err := strconv.ParseUint(matches[2], 10, 32)
	if err != nil {
		return errors.Join(ErrInvalidMemParm, err)
	}

	t, err := strconv.ParseUint(matches[3], 10, 32)
	if err != nil {
		return errors.Join(ErrInvalidTimeParm, err)
	}

	p, err := strconv.ParseUint(matches[4], 10, 8)
	if err != nil {
		return errors.Join(ErrInvalidThreadParm, err)
	}

	k, err := strconv.ParseUint(matches[5], 10, 32)
	if err != nil {
		return fmt.Errorf("invalid key length parameter: %w", err)
	}

	// Decode salt
	saltEncoded := matches[6]
	saltBuf, err := base64.StdEncoding.DecodeString(saltEncoded)
	if err != nil {
		return fmt.Errorf("invalid salt encoding: %w", err)
	}
	defer eraseBuf(saltBuf)

	// Decode hash
	hashEncoded := matches[7]
	storedHash, err := base64.StdEncoding.DecodeString(hashEncoded)
	if err != nil {
		return fmt.Errorf("invalid hash encoding: %w", err)
	}
	defer eraseBuf(storedHash)

	// Compute the hash with the same parameters
	computedHash := argon2.IDKey(
		[]byte(password),
		saltBuf,
		uint32(t),
		uint32(mem),
		uint8(p),
		uint32(k),
	)
	defer eraseBuf(computedHash)

	// Compare hashes
	if bytes.Equal(storedHash, computedHash) {
		return nil
	}

	return ErrMismatchedHashAndPassword
}

// generateRandomBytes generates n random bytes
func generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)

	_, err := rand.Read(b)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return b, nil
}

// eraseBuf fills len(buf) with space characters.
// if buf is nil or zero length, eraseBuf takes no action.
func eraseBuf(buf []byte) {
	for i := 0; i < len(buf); i++ {
		buf[i] = ' '
	}
}
