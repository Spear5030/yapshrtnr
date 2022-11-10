package module

import (
	"errors"
	"math/rand"
	"net/url"
	"strings"
)

var errURLshorting = errors.New("handler: wrong URL")

func ShortingURL(longURL string) (string, error) {
	u, err := url.Parse(longURL)
	if err != nil || u.Hostname() == "" {
		return "", errURLshorting
	}
	// на stackoverflow есть варианты быстрее, но этот более читабелен. первые два символа - визуальная привязка к домену
	const symBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	host := strings.Split(u.Hostname(), ".")
	b := make([]byte, 8)
	b[0] = host[0][0]
	b[1] = host[len(host)-1][0]
	for i := 2; i < len(b); i++ {
		b[i] = symBytes[rand.Intn(len(symBytes))]
	}
	return string(b), nil
}
