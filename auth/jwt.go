package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWT 实现 TokenProvider
type JWT struct {
	cfg        Config
	signMethod jwt.SigningMethod
	signKey    any
	verifyKey  any
}

// NewJWT 创建 JWT provider
func NewJWT(cfg Config) (*JWT, error) {
	j := &JWT{cfg: cfg}
	if err := j.init(); err != nil {
		return nil, err
	}
	return j, nil
}

func (j *JWT) init() error {
	method := string(j.cfg.SigningMethod)
	if method == "" {
		method = string(HS256)
	}

	j.signMethod = jwt.GetSigningMethod(method)
	if j.signMethod == nil {
		return ErrInvalidAlgorithm
	}

	switch {
	case strings.HasPrefix(method, "HS"):
		return j.initHMAC()
	case strings.HasPrefix(method, "RS"):
		return j.initRSA()
	case strings.HasPrefix(method, "ES"):
		return j.initECDSA()
	default:
		return ErrInvalidAlgorithm
	}
}

func (j *JWT) initHMAC() error {
	if j.cfg.SecretProvider == nil {
		return errors.New("auth: secret provider required for HMAC")
	}
	secret := j.cfg.SecretProvider()
	j.signKey = secret
	j.verifyKey = secret
	return nil
}

func (j *JWT) initRSA() error {
	kp, err := j.cfg.KeyPairProvider()
	if err != nil {
		return err
	}

	privKey, err := parseRSAPrivateKey(kp.PrivateKey)
	if err != nil {
		return err
	}
	j.signKey = privKey

	pubKey, err := parseRSAPublicKey(kp.PublicKey)
	if err != nil {
		return err
	}
	j.verifyKey = pubKey
	return nil
}

func (j *JWT) initECDSA() error {
	kp, err := j.cfg.KeyPairProvider()
	if err != nil {
		return err
	}

	privKey, err := parseECDSAPrivateKey(kp.PrivateKey)
	if err != nil {
		return err
	}
	j.signKey = privKey

	pubKey, err := parseECDSAPublicKey(kp.PublicKey)
	if err != nil {
		return err
	}
	j.verifyKey = pubKey
	return nil
}

func (j *JWT) GenerateToken(_ context.Context, claims Claims) (string, error) {
	now := time.Now()
	if claims.IssuedAt.IsZero() {
		claims.IssuedAt = now
	}
	if claims.ExpiresAt.IsZero() {
		claims.ExpiresAt = now.Add(j.cfg.Expiry)
	}

	jwtClaims := jwt.MapClaims{
		"sub": claims.UserID,
		"exp": claims.ExpiresAt.Unix(),
		"iat": claims.IssuedAt.Unix(),
		"iss": j.cfg.Issuer,
	}
	for k, v := range claims.Extra {
		jwtClaims[k] = v
	}

	token := jwt.NewWithClaims(j.signMethod, jwtClaims)
	return token.SignedString(j.signKey)
}

func (j *JWT) ValidateToken(_ context.Context, tokenStr string) (*Claims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != j.signMethod.Alg() {
			return nil, ErrInvalidAlgorithm
		}
		return j.verifyKey, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	claims := &Claims{Extra: make(map[string]any)}
	if sub, ok := mapClaims["sub"].(float64); ok {
		claims.UserID = uint64(sub)
	}
	if exp, ok := mapClaims["exp"].(float64); ok {
		claims.ExpiresAt = time.Unix(int64(exp), 0)
	}
	if iat, ok := mapClaims["iat"].(float64); ok {
		claims.IssuedAt = time.Unix(int64(iat), 0)
	}

	if time.Now().After(claims.ExpiresAt) {
		return nil, ErrExpiredToken
	}

	for k, v := range mapClaims {
		if k != "sub" && k != "exp" && k != "iat" && k != "iss" {
			claims.Extra[k] = v
		}
	}
	return claims, nil
}

// --- Key Parsers ---

func parseRSAPrivateKey(data []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("auth: invalid RSA private key PEM")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func parseRSAPublicKey(data []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("auth: invalid RSA public key PEM")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("auth: not RSA public key")
	}
	return rsaPub, nil
}

func parseECDSAPrivateKey(data []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("auth: invalid ECDSA private key PEM")
	}
	return x509.ParseECPrivateKey(block.Bytes)
}

func parseECDSAPublicKey(data []byte) (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("auth: invalid ECDSA public key PEM")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	ecPub, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("auth: not ECDSA public key")
	}
	return ecPub, nil
}

var _ TokenProvider = (*JWT)(nil)
