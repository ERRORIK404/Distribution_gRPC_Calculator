package tokenezation

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func GenerateToken(login string, secret string) (string, error) {
	iat := time.Now()
	exp := iat.Add(time.Minute * 10)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"login": login,
		"iat": iat.Unix(),
		"nbf": iat.Unix(),
		"exp": exp.Unix(),
	})
	return token.SignedString([]byte(secret))
}

func CheckToken(tokenString string, secret string) (string, error) {
    // Парсим токен с проверкой секрета
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        // Проверяем метод подписи
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(secret), nil // Исправлено JMT → JWT
    })

    if err != nil {
        return "", fmt.Errorf("token parsing failed: %w", err)
    }

    // Проверяем валидность токена и claims
    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        // Извлекаем login из claims
        login, ok := claims["login"].(string)
        if !ok {
            return "", fmt.Errorf("login claim is missing or invalid")
        }
        return login, nil
    }

    return "", fmt.Errorf("invalid token")
}