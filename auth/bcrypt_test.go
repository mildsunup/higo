package auth

import "testing"

func TestBcrypt_HashPassword(t *testing.T) {
	bcrypt := NewBcrypt(10)
	password := "test123"

	hash, err := bcrypt.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == password {
		t.Error("Hash should not equal password")
	}
}

func TestBcrypt_ComparePassword(t *testing.T) {
	bcrypt := NewBcrypt(10)
	password := "test123"
	hash, _ := bcrypt.HashPassword(password)

	if err := bcrypt.ComparePassword(hash, password); err != nil {
		t.Error("ComparePassword should succeed")
	}

	if err := bcrypt.ComparePassword(hash, "wrong"); err == nil {
		t.Error("ComparePassword should fail")
	}
}
