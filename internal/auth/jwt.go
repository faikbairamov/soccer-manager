package auth

import (
	"errors"
	"time"

	"github.com/faikbairamov/soccer-manager/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

func GenerateToken(userID uuid.UUID, secret string, expiryHours int) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiryHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ValidateToken(tokenStr, secret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{},
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(secret), nil
		})

	if err != nil {
		return uuid.Nil, domain.ErrUnauthorized
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return uuid.Nil, domain.ErrUnauthorized
	}
	return claims.UserID, nil
}
