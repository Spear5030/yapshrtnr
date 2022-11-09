package handler

import (
	"fmt"
	"github.com/Spear5030/yapshrtnr/internal/module"
	"io"
	"net/http"
	"strings"
)

type Handler struct {
	Storage storage
	Addr    string
}

type storage interface {
	SetURL(short, long string)
	GetURL(short string) string
}

func New(storage storage, addr string) *Handler {
	return &Handler{
		Storage: storage,
		Addr:    addr,
	}
}

func (h *Handler) PostURL(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	short, err := module.ShortingURL(string(b))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	h.Storage.SetURL(short, string(b))

	w.WriteHeader(201)
	w.Write([]byte(fmt.Sprintf("http://%s/%s", h.Addr, short)))
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
