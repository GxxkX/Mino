package jwt

import (
	"crypto/rsa"
	"errors"
	"os"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID    string `json:"sub"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	TokenType string `json:"token_type"` // "access" or "refresh"
	gojwt.RegisteredClaims
}

type Manager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewManager(privateKeyPath, publicKeyPath string, accessTTL, refreshTTL time.Duration) (*Manager, error) {
	privBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, err
	}
	privKey, err := gojwt.ParseRSAPrivateKeyFromPEM(privBytes)
	if err != nil {
		return nil, err
	}

	pubBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, err
	}
	pubKey, err := gojwt.ParseRSAPublicKeyFromPEM(pubBytes)
	if err != nil {
		return nil, err
	}

	return &Manager{
		privateKey: privKey,
		publicKey:  pubKey,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}, nil
}

func (m *Manager) GenerateAccessToken(userID, username, role string) (string, error) {
	claims := Claims{
		UserID:    userID,
		Username:  username,
		Role:      role,
		TokenType: "access",
		RegisteredClaims: gojwt.RegisteredClaims{
			ExpiresAt: gojwt.NewNumericDate(time.Now().Add(m.accessTTL)),
			IssuedAt:  gojwt.NewNumericDate(time.Now()),
		},
	}
	return gojwt.NewWithClaims(gojwt.SigningMethodRS256, claims).SignedString(m.privateKey)
}

func (m *Manager) GenerateRefreshToken(userID, username, role string) (string, error) {
	claims := Claims{
		UserID:    userID,
		Username:  username,
		Role:      role,
		TokenType: "refresh",
		RegisteredClaims: gojwt.RegisteredClaims{
			ExpiresAt: gojwt.NewNumericDate(time.Now().Add(m.refreshTTL)),
			IssuedAt:  gojwt.NewNumericDate(time.Now()),
		},
	}
	return gojwt.NewWithClaims(gojwt.SigningMethodRS256, claims).SignedString(m.privateKey)
}

func (m *Manager) Validate(tokenStr string) (*Claims, error) {
	token, err := gojwt.ParseWithClaims(tokenStr, &Claims{}, func(t *gojwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*gojwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.publicKey, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
