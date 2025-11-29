// Package auth 提供认证相关功能
package auth

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token expired")
	ErrInvalidAlgorithm = errors.New("invalid signing algorithm")
)

// SigningMethod 签名算法
type SigningMethod string

const (
	HS256 SigningMethod = "HS256"
	HS384 SigningMethod = "HS384"
	HS512 SigningMethod = "HS512"
	RS256 SigningMethod = "RS256"
	RS384 SigningMethod = "RS384"
	RS512 SigningMethod = "RS512"
	ES256 SigningMethod = "ES256"
	ES384 SigningMethod = "ES384"
	ES512 SigningMethod = "ES512"
)

// Claims token 声明
type Claims struct {
	UserID    uint64
	ExpiresAt time.Time
	IssuedAt  time.Time
	Extra     map[string]any
}

// TokenProvider token 提供者接口
type TokenProvider interface {
	GenerateToken(ctx context.Context, claims Claims) (string, error)
	ValidateToken(ctx context.Context, token string) (*Claims, error)
}

// PasswordHasher 密码哈希接口
type PasswordHasher interface {
	HashPassword(password string) (string, error)
	ComparePassword(hashed, password string) error
}

// SecretProvider 密钥提供者（避免密钥直接暴露在配置中）
type SecretProvider func() []byte

// KeyPair 非对称密钥对
type KeyPair struct {
	PrivateKey []byte
	PublicKey  []byte
}

// KeyPairProvider 密钥对提供者
type KeyPairProvider func() (*KeyPair, error)

// Config 认证配置
type Config struct {
	SigningMethod   SigningMethod   `yaml:"signing_method" mapstructure:"signing_method"`
	Expiry          time.Duration   `yaml:"expiry" mapstructure:"expiry"`
	RefreshExpiry   time.Duration   `yaml:"refresh_expiry" mapstructure:"refresh_expiry"`
	Issuer          string          `yaml:"issuer" mapstructure:"issuer"`
	SecretProvider  SecretProvider  `yaml:"-" mapstructure:"-"` // 对称密钥
	KeyPairProvider KeyPairProvider `yaml:"-" mapstructure:"-"` // 非对称密钥
}

// DefaultConfig 默认配置
func DefaultConfig() Config {
	return Config{
		SigningMethod: HS256,
		Expiry:        24 * time.Hour,
		RefreshExpiry: 7 * 24 * time.Hour,
		Issuer:        "higo",
	}
}

// WithSecret 设置对称密钥
func (c Config) WithSecret(secret []byte) Config {
	c.SecretProvider = func() []byte { return secret }
	return c
}

// WithSecretProvider 设置密钥提供者
func (c Config) WithSecretProvider(p SecretProvider) Config {
	c.SecretProvider = p
	return c
}

// WithKeyPair 设置非对称密钥对
func (c Config) WithKeyPair(privateKey, publicKey []byte) Config {
	c.KeyPairProvider = func() (*KeyPair, error) {
		return &KeyPair{PrivateKey: privateKey, PublicKey: publicKey}, nil
	}
	return c
}

// WithKeyPairProvider 设置密钥对提供者
func (c Config) WithKeyPairProvider(p KeyPairProvider) Config {
	c.KeyPairProvider = p
	return c
}
