package util

import (
	"regexp"
	"testing"
)

func TestGenerateMTProtoSecret(t *testing.T) {
	secret, err := GenerateMTProtoSecret()
	if err != nil {
		t.Fatalf("GenerateMTProtoSecret() returned error: %v", err)
	}

	// Should be 32 hex characters (16 bytes)
	if len(secret) != 32 {
		t.Errorf("GenerateMTProtoSecret() returned secret of length %d, expected 32", len(secret))
	}

	// Should only contain hex characters
	if !ValidateMTProtoSecret(secret) {
		t.Errorf("GenerateMTProtoSecret() returned invalid secret: %s", secret)
	}

	// Generate another and ensure they're different (randomness check)
	secret2, err := GenerateMTProtoSecret()
	if err != nil {
		t.Fatalf("Second GenerateMTProtoSecret() returned error: %v", err)
	}
	if secret == secret2 {
		t.Error("GenerateMTProtoSecret() returned same secret twice (very unlikely)")
	}
}

func TestValidateMTProtoSecret(t *testing.T) {
	tests := []struct {
		name     string
		secret   string
		expected bool
	}{
		{"valid 32 hex lowercase", "0123456789abcdef0123456789abcdef", true},
		{"valid 32 hex uppercase", "0123456789ABCDEF0123456789ABCDEF", true},
		{"valid 32 hex mixed case", "0123456789AbCdEf0123456789aBcDeF", true},
		{"too short (31 chars)", "0123456789abcdef0123456789abcde", false},
		{"too long (33 chars)", "0123456789abcdef0123456789abcdef0", false},
		{"contains non-hex (g)", "0123456789abcdef0123456789abcdeg", false},
		{"empty string", "", false},
		{"contains space", "0123456789abcdef 123456789abcdef", false},
		{"contains dash", "0123456789abcdef-123456789abcdef", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateMTProtoSecret(tt.secret)
			if result != tt.expected {
				t.Errorf("ValidateMTProtoSecret(%q) = %v, expected %v", tt.secret, result, tt.expected)
			}
		})
	}
}

func TestGenerateUUID(t *testing.T) {
	uuid := GenerateUUID()

	// Should match UUID format
	uuidPattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	if !uuidPattern.MatchString(uuid) {
		t.Errorf("GenerateUUID() returned invalid UUID format: %s", uuid)
	}

	// Generate another and ensure they're different
	uuid2 := GenerateUUID()
	if uuid == uuid2 {
		t.Error("GenerateUUID() returned same UUID twice (very unlikely)")
	}
}

func TestValidateUUID(t *testing.T) {
	tests := []struct {
		name     string
		uuid     string
		expected bool
	}{
		{"valid UUID lowercase", "103b0aae-3384-4d23-9f5b-2d15be377a23", true},
		{"valid UUID uppercase", "103B0AAE-3384-4D23-9F5B-2D15BE377A23", true},
		{"valid UUID mixed case", "103B0aae-3384-4d23-9F5B-2d15be377a23", true},
		{"missing dashes", "103b0aae33844d239f5b2d15be377a23", false},
		{"wrong dash positions", "103b0aa-e3384-4d23-9f5b-2d15be377a23", false},
		{"too short", "103b0aae-3384-4d23-9f5b-2d15be377a2", false},
		{"too long", "103b0aae-3384-4d23-9f5b-2d15be377a234", false},
		{"empty string", "", false},
		{"contains non-hex", "103b0aag-3384-4d23-9f5b-2d15be377a23", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateUUID(tt.uuid)
			if result != tt.expected {
				t.Errorf("ValidateUUID(%q) = %v, expected %v", tt.uuid, result, tt.expected)
			}
		})
	}
}

func TestParseUUID(t *testing.T) {
	// Valid UUID should parse
	validUUID := "103b0aae-3384-4d23-9f5b-2d15be377a23"
	parsed, err := ParseUUID(validUUID)
	if err != nil {
		t.Errorf("ParseUUID(%q) returned error: %v", validUUID, err)
	}
	if parsed.String() != validUUID {
		t.Errorf("ParseUUID(%q) returned %q", validUUID, parsed.String())
	}

	// Invalid UUID should error
	invalidUUID := "not-a-uuid"
	_, err = ParseUUID(invalidUUID)
	if err == nil {
		t.Errorf("ParseUUID(%q) should have returned error", invalidUUID)
	}
}

func TestGenerateSecureToken(t *testing.T) {
	// Test different lengths
	lengths := []int{8, 16, 32}

	for _, length := range lengths {
		token, err := GenerateSecureToken(length)
		if err != nil {
			t.Errorf("GenerateSecureToken(%d) returned error: %v", length, err)
			continue
		}

		// Output is hex encoded, so length should be 2x
		expectedLen := length * 2
		if len(token) != expectedLen {
			t.Errorf("GenerateSecureToken(%d) returned token of length %d, expected %d", length, len(token), expectedLen)
		}

		// Should be valid hex
		hexPattern := regexp.MustCompile(`^[0-9a-f]+$`)
		if !hexPattern.MatchString(token) {
			t.Errorf("GenerateSecureToken(%d) returned non-hex token: %s", length, token)
		}
	}
}

func TestMaskSecret(t *testing.T) {
	tests := []struct {
		name     string
		secret   string
		expected string
	}{
		{"32 char secret", "0123456789abcdef0123456789abcdef", "0123****cdef"},
		{"short secret (8 chars)", "12345678", "****"},
		{"shorter secret", "1234567", "****"},
		{"very short", "abc", "****"},
		{"9 char secret", "123456789", "1234****6789"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskSecret(tt.secret)
			if result != tt.expected {
				t.Errorf("MaskSecret(%q) = %q, expected %q", tt.secret, result, tt.expected)
			}
		})
	}
}

func TestMaskUUID(t *testing.T) {
	tests := []struct {
		name     string
		uuid     string
		expected string
	}{
		{"valid UUID", "103b0aae-3384-4d23-9f5b-2d15be377a23", "103b0aae-****-****-****-************"},
		{"short string", "1234567", "****"},
		{"exactly 8 chars", "12345678", "12345678-****-****-****-************"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskUUID(tt.uuid)
			if result != tt.expected {
				t.Errorf("MaskUUID(%q) = %q, expected %q", tt.uuid, result, tt.expected)
			}
		})
	}
}

func TestMTProtoSecretConstant(t *testing.T) {
	if MTProtoSecretLength != 16 {
		t.Errorf("MTProtoSecretLength = %d, expected 16", MTProtoSecretLength)
	}
}
