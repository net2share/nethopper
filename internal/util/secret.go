package util

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"

	"github.com/google/uuid"
)

const (
	// MTProto secret length (32 hex characters = 16 bytes)
	MTProtoSecretLength = 16
)

var (
	// MTProto secret pattern: 32 hex characters
	mtprotoSecretPattern = regexp.MustCompile(`^[0-9a-fA-F]{32}$`)
	// UUID pattern
	uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
)

// GenerateMTProtoSecret generates a random 32-character hex secret for MTProto
// Equivalent to: head -c 16 /dev/urandom | xxd -ps
func GenerateMTProtoSecret() (string, error) {
	bytes := make([]byte, MTProtoSecretLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// ValidateMTProtoSecret validates that a string is a valid MTProto secret
func ValidateMTProtoSecret(secret string) bool {
	return mtprotoSecretPattern.MatchString(secret)
}

// GenerateUUID generates a new random UUID for VMess
func GenerateUUID() string {
	return uuid.New().String()
}

// ValidateUUID validates that a string is a valid UUID
func ValidateUUID(id string) bool {
	return uuidPattern.MatchString(id)
}

// ParseUUID parses and validates a UUID string
func ParseUUID(id string) (uuid.UUID, error) {
	return uuid.Parse(id)
}

// GenerateSecureToken generates a secure random token of specified length
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// MaskSecret masks a secret for display, showing only first and last 4 characters
func MaskSecret(secret string) string {
	if len(secret) <= 8 {
		return "****"
	}
	return secret[:4] + "****" + secret[len(secret)-4:]
}

// MaskUUID masks a UUID for display, showing only the first segment
func MaskUUID(id string) string {
	if len(id) < 8 {
		return "****"
	}
	return id[:8] + "-****-****-****-************"
}
