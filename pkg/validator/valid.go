package validator

import (
	"errors"
	"regexp"
	"unicode/utf8"
)

var (
	ErrLoginTooShort      = errors.New("login must be at least 3 characters long")
	ErrLoginTooLong       = errors.New("login must be no more than 20 characters long")
	ErrLoginInvalidChars  = errors.New("login can only contain letters, numbers, underscores, and hyphens")
	ErrLoginStartsWithNum = errors.New("login cannot start with a number")
)

// ValidateLogin проверяет логин на безопасность и корректность.
func ValidateLogin(login string) error {
	// Проверка длины
	if utf8.RuneCountInString(login) < 3 {
		return ErrLoginTooShort
	}
	if utf8.RuneCountInString(login) > 20 {
		return ErrLoginTooLong
	}

	// Проверка первого символа (не должен быть цифрой)
	if len(login) > 0 && login[0] >= '0' && login[0] <= '9' {
		return ErrLoginStartsWithNum
	}

	// Регулярное выражение: только буквы, цифры, "_" и "-"
	validLoginRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validLoginRegex.MatchString(login) {
		return ErrLoginInvalidChars
	}

	return nil
}