package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims определяет структуру полезной нагрузки (payload) JWT токена.
type Claims struct {
	jwt.RegisteredClaims
	UserID int64
}

const tokenTTL = 12 * time.Hour

// BuildJWTString создает новую JWT строку для указанного ID пользователя.
func BuildJWTString(userID int64, secretKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда токен выдан
			IssuedAt: jwt.NewNumericDate(time.Now()),
			// когда токен истекает
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenTTL)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// GetUserID извлекает ID пользователя из JWT строки.
func GetUserID(tokenString string, secretKey string) (int64, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	return claims.UserID, err
}
