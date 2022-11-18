package module

import (
	"math/rand"
)

//var errURLshorting = errors.New("handler: wrong URL")

func ShortingURL(longURL string) (string, error) {
	b := make([]byte, 8)
	/*	u, err := url.Parse(longURL)
		if err != nil || u.Hostname() == "" {
			return "", errURLshorting
		}

		host := strings.Split(u.Hostname(), ".")

		b[0] = host[0][0]
		b[1] = host[len(host)-1][0]*/
	// на stackoverflow есть варианты быстрее, но этот более читабелен. первые два символа - визуальная привязка к домену
	const symBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for i := 0; i < len(b); i++ {
		b[i] = symBytes[rand.Intn(len(symBytes))]
	}
	return string(b), nil
}
