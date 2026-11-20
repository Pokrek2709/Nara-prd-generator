package security_test

import (
	"testing"

	"github.com/your-org/prd-generator/backend/internal/infrastructure/security"
)

func TestEncryptDecrypt(t *testing.T) {
	secret := "test-super-secret-key-that-is-at-least-32-chars-long"
	plaintexts := []string{
		"sk-proj-1234567890abcdef",
		"gsk_my_groq_api_key_secret_value",
		"AIzaSyDemoGoogleKey12345",
		"short",
		"a very long api key with special characters: !@#$%^&*()_+{}[]:;<>?,./~`",
	}

	for _, pt := range plaintexts {
		enc, err := security.Encrypt(pt, secret)
		if err != nil {
			t.Fatalf("Encrypt failed for %q: %v", pt, err)
		}
		if enc == pt {
			t.Fatalf("Encrypted value should not match plaintext: %q", enc)
		}

		dec, err := security.Decrypt(enc, secret)
		if err != nil {
			t.Fatalf("Decrypt failed for %q: %v", pt, err)
		}
		if dec != pt {
			t.Fatalf("Decrypted value %q does not match original %q", dec, pt)
		}
	}
}

func TestEncryptEmptyString(t *testing.T) {
	secret := "test-super-secret-key-that-is-at-least-32-chars-long"
	enc, err := security.Encrypt("", secret)
	if err != nil {
		t.Fatalf("Encrypt empty string failed: %v", err)
	}
	if enc != "" {
		t.Fatalf("Expected empty ciphertext, got %q", enc)
	}

	dec, err := security.Decrypt("", secret)
	if err != nil {
		t.Fatalf("Decrypt empty string failed: %v", err)
	}
	if dec != "" {
		t.Fatalf("Expected empty plaintext, got %q", dec)
	}
}

func TestDecryptLegacyPlaintext(t *testing.T) {
	secret := "test-super-secret-key-that-is-at-least-32-chars-long"
	legacyKey := "sk-1234567890abcdef"

	dec, err := security.Decrypt(legacyKey, secret)
	if err != nil {
		t.Fatalf("Decrypt legacy key failed: %v", err)
	}
	if dec != legacyKey {
		t.Fatalf("Expected %q, got %q", legacyKey, dec)
	}
}

func TestDecryptWithWrongSecret(t *testing.T) {
	secret1 := "secret-one-key-32-chars-long-123456"
	secret2 := "secret-two-key-32-chars-long-654321"

	enc, err := security.Encrypt("my-secret-key-data", secret1)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	_, err = security.Decrypt(enc, secret2)
	if err == nil {
		t.Fatal("Expected decryption with wrong secret to fail, but it succeeded")
	}
}
