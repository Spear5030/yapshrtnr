package handler

import (
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/Spear5030/yapshrtnr/internal/domain"
	"github.com/Spear5030/yapshrtnr/internal/module"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type Handler struct {
	Storage storage
	BaseURL string
}

type storage interface {
	SetURL(user, short, long string)
	GetURL(short string) string
	GetURLsByUser(user string) (urls map[string]string)
	SetBatchURLs(ctx context.Context, urls []domain.URL) error
	Ping() error
}

type link struct {
	Short string `json:"short_url"`
	Long  string `json:"original_url"`
}

type input struct {
	URL string `json:"url"`
}

type result struct {
	Result string `json:"result"`
}

type batchInput struct {
	Long          string `json:"original_url"`
	CorrelationID string `json:"correlation_id"`
}

type batchTmp struct {
	Short         string
	Long          string
	CorrelationID string
}

type batchResult struct {
	Short         string `json:"short_url"`
	CorrelationID string `json:"correlation_id"`
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
	user, err := getUserIDFROMCookie(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	h.Storage.SetURL(user, short, string(b))

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

func (h *Handler) PingDB(w http.ResponseWriter, r *http.Request) {
	if h.Storage.Ping() == nil {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusInternalServerError)

}

func (h *Handler) PostBatch(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	inputs := make([]batchInput, 0)
	if err = json.Unmarshal(b, &inputs); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	user, err := getUserIDFROMCookie(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	tmps := make([]batchTmp, 0, len(inputs))
	urls := make([]domain.URL, 0, len(inputs))
	for _, url := range inputs {
		tmpShort, err := module.ShortingURL(url.Long)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		tmps = append(tmps, batchTmp{
			Short:         tmpShort,
			Long:          url.Long,
			CorrelationID: url.CorrelationID,
		})
		urls = append(urls, domain.URL{
			User:  user,
			Short: tmpShort,
			Long:  url.Long,
		})
	}
	ctx, cancel := context.WithTimeout(r.Context(), time.Minute)
	defer cancel()

	result := make([]batchResult, len(inputs))
	fmt.Println(h.Storage.SetBatchURLs(ctx, urls))
	for i, urlEnt := range tmps {
		result[i] = batchResult{
			Short:         fmt.Sprintf("%s/%s", h.BaseURL, urlEnt.Short),
			CorrelationID: urlEnt.CorrelationID,
		}
	}
	resJSON, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	w.WriteHeader(201)
	w.Write(resJSON)
}

func (h *Handler) PostJSON(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	urlEnt := input{}
	if err := json.Unmarshal(b, &urlEnt); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	short, err := module.ShortingURL(urlEnt.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	user, err := getUserIDFROMCookie(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	h.Storage.SetURL(user, short, urlEnt.URL)

	w.WriteHeader(201)
	res := result{}
	res.Result = fmt.Sprintf("%s/%s", h.BaseURL, short)
	resJSON, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	w.Write(resJSON)
}

func (h *Handler) GetURLsByUser(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	urls := h.Storage.GetURLsByUser(cookie.Value)
	if len(urls) == 0 {
		w.WriteHeader(http.StatusNoContent)
	}
	var one link
	var links []link
	for key, value := range urls {
		one.Short = h.BaseURL + "/" + key
		one.Long = value
		links = append(links, one)
	}
	resJSON, err := json.Marshal(links)
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

func CheckCookies(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookies := r.Cookies()
		if verifyCookies(cookies) { // TODO rewrite

		} else {
			b := make([]byte, 16)
			_, _ = rand.Read(b)
			id := sha256.Sum256(b)
			key := []byte("strongKey") // TODO move to config
			h := hmac.New(sha256.New, key)
			h.Write(id[:])
			dst := h.Sum(nil)
			cookie := &http.Cookie{
				Name:  "id",
				Value: fmt.Sprintf("%x", id),
				Path:  "/",
			}
			http.SetCookie(w, cookie)
			r.AddCookie(cookie)
			cookie = &http.Cookie{
				Name:  "token",
				Value: fmt.Sprintf("%x", dst),
				Path:  "/",
			}
			http.SetCookie(w, cookie)
			r.AddCookie(cookie)
		}
		next.ServeHTTP(w, r)
	})
}

func verifyCookies(cookies []*http.Cookie) bool {
	key := []byte("strongKey") // TODO move to config
	var id, token []byte
	for _, cookie := range cookies {
		switch cookie.Name {
		case "id":
			id, _ = hex.DecodeString(cookie.Value)
		case "token":
			token, _ = hex.DecodeString(cookie.Value)
		}
	}
	if len(id) == 0 {
		return false
	}
	h := hmac.New(sha256.New, key)
	h.Write(id)
	return hmac.Equal(h.Sum(nil), token)
}

func getUserIDFROMCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie("id")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}
