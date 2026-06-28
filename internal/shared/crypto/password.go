package crypto

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	// PasswordArgon2Memory is the Argon2id memory cost in KiB.
	PasswordArgon2Memory uint32 = 64 * 1024
	// PasswordArgon2Iterations is the Argon2id iteration count.
	PasswordArgon2Iterations uint32 = 3
	// PasswordArgon2Parallelism is the Argon2id parallelism factor.
	PasswordArgon2Parallelism uint8 = 1
	// PasswordSaltLength is the random salt length in bytes.
	PasswordSaltLength = 16
	// PasswordKeyLength is the derived key length in bytes.
	PasswordKeyLength uint32 = 32
)

func HashPassword(password string) (string, error) {
	salt := make([]byte, PasswordSaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, PasswordArgon2Iterations, PasswordArgon2Memory, PasswordArgon2Parallelism, PasswordKeyLength)
	saltEncoded := base64.RawStdEncoding.EncodeToString(salt)
	hashEncoded := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		PasswordArgon2Memory,
		PasswordArgon2Iterations,
		PasswordArgon2Parallelism,
		saltEncoded,
		hashEncoded,
	), nil
}

func VerifyPassword(password string, encodedHash string) bool {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" {
		return false
	}

	params := strings.Split(parts[3], ",")
	if len(params) != 3 {
		return false
	}
	memory, err := parseUint32Param(params[0], "m")
	if err != nil {
		return false
	}
	iterations, err := parseUint32Param(params[1], "t")
	if err != nil {
		return false
	}
	parallelism64, err := parseUint32Param(params[2], "p")
	if err != nil {
		return false
	}
	parallelism := uint8(parallelism64)

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}
	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}

	hash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(expectedHash)))
	return subtle.ConstantTimeCompare(hash, expectedHash) == 1
}

func parseUint32Param(value string, key string) (uint32, error) {
	prefix := key + "="
	if !strings.HasPrefix(value, prefix) {
		return 0, fmt.Errorf("invalid argon2 parameter")
	}
	parsed, err := strconv.ParseUint(strings.TrimPrefix(value, prefix), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(parsed), nil
}
