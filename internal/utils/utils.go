package utils

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// jwtSecret хранит секретный ключ для JWT, загружаемый из переменной окружения
var jwtSecret []byte

func init() {
	// Загрузка секретного ключа из переменной окружения
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))
}

// GenerateShortURL генерирует случайную строку для использования в качестве сокращенной ссылки.
func GenerateShortURL() string {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

// GenerateToken создает JWT токен для пользователя.
func GenerateToken(email string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
	})
	return token.SignedString(jwtSecret)
}

// VerifyToken проверяет JWT токен и извлекает email пользователя.
func VerifyToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrInvalidKey
		}
		return jwtSecret, nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		email := claims["email"].(string)
		return email, nil
	} else {
		return "", err
	}
}
