package handler

import (
	"errors"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
)

type Handler struct {
	Storage storage
}

type storage interface {
	SetURL(short, long string)
	GetURL(short string) string
}

func New(storage storage) *Handler {
	return &Handler{Storage: storage}
}

func (h *Handler) PostURL(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	short, err := shortingURL(string(b))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	h.Storage.SetURL(short, string(b))

	w.WriteHeader(201)
	w.Write([]byte("http://localhost:8080/" + short))
}

func (h *Handler) GetURL(w http.ResponseWriter, r *http.Request) {
	short := strings.TrimLeft(r.URL.Path, "/")
	v := h.Storage.GetURL(short)
	if len(v) > 0 {
		w.Header().Set("Location", v)
		w.WriteHeader(307)
		return
	}
	http.Error(w, "Wrong ID", http.StatusBadRequest)
}

var errURLshorting = errors.New("handler: wrong URL")

func shortingURL(longURL string) (string, error) {
	u, err := url.Parse(longURL)
	if err != nil || u.Hostname() == "" {
		//fmt.Println(err)
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
