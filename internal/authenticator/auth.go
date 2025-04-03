package authenticator

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	uuid "github.com/satori/go.uuid"
)

const tokenExp = time.Hour * 3

type claims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID
}

// BuildJWTString makes token and returns it as a string.
func BuildJWTString(id uuid.UUID, key []byte) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
		},
		// собственное утверждение
		UserID: id,
	})

	// создаём строку токена
	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", err
	}

	// возвращаем строку токена
	return tokenString, nil
}
