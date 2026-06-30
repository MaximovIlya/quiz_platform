package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	db "github.com/MaximovIlya/vk_practice_project/db/sqlc"
)

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

type Claims struct {
	UserID string    `json:"user_id"`
	Role   db.Role   `json:"role"`
	Type   TokenType `json:"type"`
	jwt.RegisteredClaims
}

type Manager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewManager(secret, accessTTL, refreshTTL string) (*Manager, error) {
	access, err := time.ParseDuration(accessTTL)
	if err != nil {
		return nil, err
	}
	refresh, err := time.ParseDuration(refreshTTL)
	if err != nil {
		return nil, err
	}
	return &Manager{
		secret:     []byte(secret),
		accessTTL:  access,
		refreshTTL: refresh,
	}, nil
}

func (m *Manager) GenerateAccess(userID string, role db.Role) (string, error) {
	return m.generate(userID, role, AccessToken, m.accessTTL)
}

func (m *Manager) GenerateRefresh(userID string, role db.Role) (string, error) {
	return m.generate(userID, role, RefreshToken, m.refreshTTL)
}

func (m *Manager) generate(userID string, role db.Role, tokenType TokenType, ttl time.Duration) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

func (m *Manager) Parse(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
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
