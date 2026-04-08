package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/leonunix/onyx_storage/dashboard/backend/internal/domain"
)

type JWTManager struct {
	secret []byte
	ttl    time.Duration
}

type Claims struct {
	Username    string              `json:"username"`
	DisplayName string              `json:"displayName"`
	Role        string              `json:"role"`
	Permissions []domain.Permission `json:"permissions"`
	jwt.RegisteredClaims
}

func NewJWTManager(secret string, ttl time.Duration) *JWTManager {
	return &JWTManager{
		secret: []byte(secret),
		ttl:    ttl,
	}
}

func (m *JWTManager) Issue(user domain.User) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Role:        user.Role,
		Permissions: user.Permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.Username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	})
	return token.SignedString(m.secret)
}

func (m *JWTManager) Parse(tokenString string) (domain.User, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		return m.secret, nil
	})
	if err != nil {
		return domain.User{}, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return domain.User{}, errors.New("invalid token")
	}

	return domain.User{
		Username:    claims.Username,
		DisplayName: claims.DisplayName,
		Role:        claims.Role,
		Permissions: claims.Permissions,
	}, nil
}
