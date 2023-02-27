// Package module содержит функции бизнес-логики
package module

import (
	"errors"
	"math/rand"

	"github.com/asaskevich/govalidator"
)

var errURLshorting = errors.New("handler: wrong URL")

// ShortingURL Сокращение и валидация URL.
func ShortingURL(longURL string) (string, error) {
	b := make([]byte, 8)
	if !govalidator.IsURL(longURL) {
		return "", errURLshorting
	}
	// на stackoverflow есть варианты быстрее, но этот более читабелен. первые два символа - визуальная привязка к домену
	const symBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for i := 0; i < len(b); i++ {
		b[i] = symBytes[rand.Intn(len(symBytes))]
	}
	return string(b), nil
}
