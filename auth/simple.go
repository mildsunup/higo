package auth

import "context"

// SimpleTokenProvider 简化的 Token 接口
type SimpleTokenProvider interface {
	Generate(ctx context.Context, userID uint64, extra map[string]any) (string, error)
	Validate(ctx context.Context, token string) (userID uint64, err error)
}

// SimplePasswordHasher 简化的密码哈希接口
type SimplePasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashed, password string) error
}

// simpleJWT 包装 JWT 实现 SimpleTokenProvider
type simpleJWT struct {
	jwt *JWT
}

// NewSimpleJWT 创建简化的 JWT provider
func NewSimpleJWT(cfg Config) (SimpleTokenProvider, error) {
	jwt, err := NewJWT(cfg)
	if err != nil {
		return nil, err
	}
	return &simpleJWT{jwt: jwt}, nil
}

func (s *simpleJWT) Generate(ctx context.Context, userID uint64, extra map[string]any) (string, error) {
	return s.jwt.GenerateToken(ctx, Claims{UserID: userID, Extra: extra})
}

func (s *simpleJWT) Validate(ctx context.Context, token string) (uint64, error) {
	claims, err := s.jwt.ValidateToken(ctx, token)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}

// simpleBcrypt 包装 Bcrypt 实现 SimplePasswordHasher
type simpleBcrypt struct {
	bcrypt *Bcrypt
}

// NewSimpleBcrypt 创建简化的 bcrypt hasher
func NewSimpleBcrypt(cost int) SimplePasswordHasher {
	return &simpleBcrypt{bcrypt: NewBcrypt(cost)}
}

func (s *simpleBcrypt) Hash(password string) (string, error) {
	return s.bcrypt.HashPassword(password)
}

func (s *simpleBcrypt) Compare(hashed, password string) error {
	return s.bcrypt.ComparePassword(hashed, password)
}
