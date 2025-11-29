package auth

import "golang.org/x/crypto/bcrypt"

// Bcrypt 实现 PasswordHasher
type Bcrypt struct {
	cost int
}

// NewBcrypt 创建 bcrypt hasher
func NewBcrypt(cost int) *Bcrypt {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}
	return &Bcrypt{cost: cost}
}

func (b *Bcrypt) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), b.cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (b *Bcrypt) ComparePassword(hashed, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
}

var _ PasswordHasher = (*Bcrypt)(nil)
