package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

var ErrInvalidPasswordHash = errors.New("invalid password hash")

type PasswordHasher struct {
	MemoryKiB   uint32
	Iterations  uint32
	Parallelism uint8
	SaltBytes   uint32
	KeyBytes    uint32
}

func DefaultPasswordHasher() PasswordHasher {
	return PasswordHasher{
		MemoryKiB:   64 * 1024,
		Iterations:  1,
		Parallelism: 2,
		SaltBytes:   16,
		KeyBytes:    32,
	}
}

func (h PasswordHasher) Hash(password string) (string, error) {
	salt := make([]byte, h.SaltBytes)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(password), salt, h.Iterations, h.MemoryKiB, h.Parallelism, h.KeyBytes)
	return fmt.Sprintf(
		"argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		h.MemoryKiB,
		h.Iterations,
		h.Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

func (h PasswordHasher) Verify(password, encoded string) (bool, error) {
	params, salt, expected, err := parseHash(encoded)
	if err != nil {
		return false, err
	}
	actual := argon2.IDKey([]byte(password), salt, params.iterations, params.memoryKiB, params.parallelism, uint32(len(expected)))
	return subtle.ConstantTimeCompare(actual, expected) == 1, nil
}

type hashParams struct {
	memoryKiB   uint32
	iterations  uint32
	parallelism uint8
}

func parseHash(encoded string) (hashParams, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 5 || parts[0] != "argon2id" || parts[1] != "v=19" {
		return hashParams{}, nil, nil, ErrInvalidPasswordHash
	}

	params := hashParams{}
	for _, part := range strings.Split(parts[2], ",") {
		keyValue := strings.SplitN(part, "=", 2)
		if len(keyValue) != 2 {
			return hashParams{}, nil, nil, ErrInvalidPasswordHash
		}
		value, err := strconv.ParseUint(keyValue[1], 10, 32)
		if err != nil {
			return hashParams{}, nil, nil, ErrInvalidPasswordHash
		}
		switch keyValue[0] {
		case "m":
			params.memoryKiB = uint32(value)
		case "t":
			params.iterations = uint32(value)
		case "p":
			if value > 255 {
				return hashParams{}, nil, nil, ErrInvalidPasswordHash
			}
			params.parallelism = uint8(value)
		default:
			return hashParams{}, nil, nil, ErrInvalidPasswordHash
		}
	}
	if params.memoryKiB == 0 || params.iterations == 0 || params.parallelism == 0 {
		return hashParams{}, nil, nil, ErrInvalidPasswordHash
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return hashParams{}, nil, nil, ErrInvalidPasswordHash
	}
	key, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return hashParams{}, nil, nil, ErrInvalidPasswordHash
	}
	return params, salt, key, nil
}
