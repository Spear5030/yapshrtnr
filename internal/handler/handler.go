package handler

import (
	"encoding/json"
	"fmt"
	"github.com/Spear5030/yapshrtnr/internal/module"
	"io"
	"net/http"
	"strings"
)

type Handler struct {
	Storage storage
	BaseURL string
}

type storage interface {
	SetURL(short, long string)
	GetURL(short string) string
}

type input struct {
	URL string `json:"url"`
}

type result struct {
	Result string `json:"result"`
}

func New(storage storage, baseURL string) *Handler {
	return &Handler{
		Storage: storage,
		BaseURL: baseURL,
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
	w.Write([]byte(fmt.Sprintf("%s/%s", h.BaseURL, short)))
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

func (h *Handler) PostJSON(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	url := input{}
	if err := json.Unmarshal(b, &url); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	short, err := module.ShortingURL(url.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	h.Storage.SetURL(short, url.URL)

	w.WriteHeader(201)
	res := result{}
	res.Result = fmt.Sprintf("%s/%s", h.BaseURL, short)
	resJSON, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	w.Write(resJSON)
}
