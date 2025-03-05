package sym_encrypt

import (
	"testing"
)

func TestNewEncryptor(t *testing.T) {
	t.Run("valid password", func(t *testing.T) {
		enc, err := NewEncryptor("test123")
		if err != nil {
			t.Errorf("NewEncryptor() unexpected error = %v", err)
		}
		if enc == nil {
			t.Error("NewEncryptor() returned nil encryptor")
		}
	})

	t.Run("empty password", func(t *testing.T) {
		enc, err := NewEncryptor("")
		if err != ErrEmptyPassword {
			t.Errorf("NewEncryptor() error = %v, expected %v", err, ErrEmptyPassword)
		}
		if enc != nil {
			t.Error("NewEncryptor() should return nil encryptor for empty password")
		}
	})
}

func TestEncryptor_Encrypt(t *testing.T) {
	enc, err := NewEncryptor("test123")
	if err != nil {
		t.Fatalf("Failed to create encryptor: %v", err)
	}

	t.Run("valid plaintext", func(t *testing.T) {
		ciphertext, err := enc.Encrypt("hello world")
		if err != nil {
			t.Errorf("Encrypt() unexpected error = %v", err)
		}
		if ciphertext == "" {
			t.Error("Encrypt() returned empty ciphertext")
		}
	})

	t.Run("empty plaintext", func(t *testing.T) {
		ciphertext, err := enc.Encrypt("")
		if err != ErrEmptyPlaintext {
			t.Errorf("Encrypt() error = %v, expected %v", err, ErrEmptyPlaintext)
		}
		if ciphertext != "" {
			t.Error("Encrypt() should return empty string for empty plaintext")
		}
	})
}

func TestEncryptor_Decrypt(t *testing.T) {
	enc, err := NewEncryptor("test123")
	if err != nil {
		t.Fatalf("Failed to create encryptor: %v", err)
	}

	originalText := "hello world"
	ciphertext, err := enc.Encrypt(originalText)
	if err != nil {
		t.Fatalf("Failed to encrypt test data: %v", err)
	}

	t.Run("valid ciphertext", func(t *testing.T) {
		plaintext, err := enc.Decrypt(ciphertext)
		if err != nil {
			t.Errorf("Decrypt() unexpected error = %v", err)
		}
		if plaintext != originalText {
			t.Errorf("Decrypt() = %v, want %v", plaintext, originalText)
		}
	})

	t.Run("invalid base64", func(t *testing.T) {
		_, err := enc.Decrypt("invalid base64!@#$")
		if err == nil {
			t.Error("Decrypt() expected error but got nil")
		}
	})

	t.Run("invalid ciphertext format", func(t *testing.T) {
		_, err := enc.Decrypt("aGVsbG8=") // valid base64 but too short
		if err != ErrInvalidCipher {
			t.Errorf("Decrypt() error = %v, expected %v", err, ErrInvalidCipher)
		}
	})
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	enc, err := NewEncryptor("test123")
	if err != nil {
		t.Fatalf("Failed to create encryptor: %v", err)
	}

	t.Run("basic ascii", func(t *testing.T) {
		original := "hello world"
		testRoundTrip(t, enc, original)
	})

	t.Run("special characters", func(t *testing.T) {
		original := "special chars !@#$%^&*()"
		testRoundTrip(t, enc, original)
	})

	t.Run("unicode characters", func(t *testing.T) {
		original := "unicode chars 你好世界"
		testRoundTrip(t, enc, original)
	})

	t.Run("numbers", func(t *testing.T) {
		original := "1234567890"
		testRoundTrip(t, enc, original)
	})
}

func testRoundTrip(t *testing.T, enc *Encryptor, original string) {
	ciphertext, err := enc.Encrypt(original)
	if err != nil {
		t.Errorf("Encrypt() error = %v", err)
		return
	}

	plaintext, err := enc.Decrypt(ciphertext)
	if err != nil {
		t.Errorf("Decrypt() error = %v", err)
		return
	}

	if plaintext != original {
		t.Errorf("Round trip failed: got %v, want %v", plaintext, original)
	}
}
