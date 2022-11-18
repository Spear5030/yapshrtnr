package handler

import (
	"compress/gzip"
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
	fmt.Println(r)
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println(string(b))
	short, err := module.ShortingURL(string(b))
	fmt.Println(short)
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
	fmt.Println(h.Storage)
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
	fmt.Println(string(b))
	url := input{}
	if err := json.Unmarshal(b, &url); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	fmt.Println(url)
	short, err := module.ShortingURL(url.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	fmt.Println(short)
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

func DecompressGZRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(`Content-Encoding`) == `gzip` {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer gz.Close()
			r.Body = gz
		}
		next.ServeHTTP(w, r)
	})
}
