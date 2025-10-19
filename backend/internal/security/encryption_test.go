package security

import (
	"encoding/base64"
	"testing"
)

func TestNewEncryptionService(t *testing.T) {
	tests := []struct {
		name    string
		keySize int
		wantErr bool
	}{
		{
			name:    "valid 32 byte key",
			keySize: 32,
			wantErr: false,
		},
		{
			name:    "invalid 16 byte key",
			keySize: 16,
			wantErr: true,
		},
		{
			name:    "invalid 64 byte key",
			keySize: 64,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := make([]byte, tt.keySize)
			_, err := NewEncryptionService(key)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewEncryptionService() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEncryptDecrypt(t *testing.T) {
	key := make([]byte, 32)
	service, err := NewEncryptionService(key)
	if err != nil {
		t.Fatalf("Failed to create encryption service: %v", err)
	}

	tests := []struct {
		name      string
		plaintext string
	}{
		{
			name:      "simple text",
			plaintext: "Hello, World!",
		},
		{
			name:      "empty string",
			plaintext: "",
		},
		{
			name:      "long text",
			plaintext: "This is a very long text that should be encrypted and decrypted correctly without any issues",
		},
		{
			name:      "special characters",
			plaintext: "!@#$%^&*()_+-=[]{}|;':\",./<>?",
		},
		{
			name:      "unicode",
			plaintext: "こんにちは世界 🌍",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt
			encrypted, err := service.Encrypt(tt.plaintext)
			if err != nil {
				t.Fatalf("Encrypt() error = %v", err)
			}

			// Empty string should return empty
			if tt.plaintext == "" && encrypted != "" {
				t.Errorf("Encrypt() empty string should return empty, got %v", encrypted)
			}

			if tt.plaintext == "" {
				return
			}

			// Decrypt
			decrypted, err := service.Decrypt(encrypted)
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}

			// Compare
			if decrypted != tt.plaintext {
				t.Errorf("Decrypt() = %v, want %v", decrypted, tt.plaintext)
			}
		})
	}
}

func TestEncryptBytes(t *testing.T) {
	key := make([]byte, 32)
	service, err := NewEncryptionService(key)
	if err != nil {
		t.Fatalf("Failed to create encryption service: %v", err)
	}

	plaintext := []byte("test data")

	encrypted, err := service.EncryptBytes(plaintext)
	if err != nil {
		t.Fatalf("EncryptBytes() error = %v", err)
	}

	decrypted, err := service.DecryptBytes(encrypted)
	if err != nil {
		t.Fatalf("DecryptBytes() error = %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("DecryptBytes() = %v, want %v", string(decrypted), string(plaintext))
	}
}

func TestGenerateKey(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	if len(key) != 32 {
		t.Errorf("GenerateKey() length = %v, want 32", len(key))
	}
}

func TestGenerateKeyBase64(t *testing.T) {
	keyStr, err := GenerateKeyBase64()
	if err != nil {
		t.Fatalf("GenerateKeyBase64() error = %v", err)
	}

	// Decode to verify it's valid base64
	key, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		t.Fatalf("Invalid base64: %v", err)
	}

	if len(key) != 32 {
		t.Errorf("Decoded key length = %v, want 32", len(key))
	}
}

func TestDecryptInvalidData(t *testing.T) {
	key := make([]byte, 32)
	service, err := NewEncryptionService(key)
	if err != nil {
		t.Fatalf("Failed to create encryption service: %v", err)
	}

	tests := []struct {
		name       string
		ciphertext string
		wantErr    bool
	}{
		{
			name:       "invalid base64",
			ciphertext: "not-valid-base64!!!",
			wantErr:    true,
		},
		{
			name:       "too short",
			ciphertext: base64.StdEncoding.EncodeToString([]byte("short")),
			wantErr:    true,
		},
		{
			name:       "empty string",
			ciphertext: "",
			wantErr:    false, // Should return empty string
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Decrypt(tt.ciphertext)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decrypt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
