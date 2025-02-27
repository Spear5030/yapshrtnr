// Package handler обрабатывает http запросы
package handler

import (
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/Spear5030/yapshrtnr/internal/domain"
	"github.com/Spear5030/yapshrtnr/internal/module"
	pckgstorage "github.com/Spear5030/yapshrtnr/internal/storage"
)

// Handler основная структура обработчика. Storage - интерфейс.
type Handler struct {
	Storage       storage
	logger        *zap.Logger
	BaseURL       string
	SecretKey     string
	trustedSubnet net.IPNet
}

type storage interface {
	SetURL(ctx context.Context, user, short, long string) error
	GetURL(ctx context.Context, short string) (string, bool)
	GetURLsByUser(ctx context.Context, user string) (urls map[string]string)
	SetBatchURLs(ctx context.Context, urls []domain.URL) error
	DeleteURLs(ctx context.Context, user string, shorts []string)
	GetUsersCount(ctx context.Context) (int, error)
	GetUrlsCount(ctx context.Context) (int, error)
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

type statsResult struct {
	Urls  int `json:"urls"`
	Users int `json:"users"`
}

// New возвращает Handler
func New(logger *zap.Logger, storage storage, baseURL string, key string, trustedSubnet net.IPNet) *Handler {
	return &Handler{
		logger:        logger,
		Storage:       storage,
		BaseURL:       baseURL,
		SecretKey:     key,
		trustedSubnet: trustedSubnet,
	}
}

// PostURL получает URL из тела запроса для сокращения. Возвращает 201 статус, либо 409 при дублировании URL
func (h *Handler) PostURL(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Info("Error readBody", zap.String("err", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.logger.Info("will shorting URL", zap.String("long", string(b)))
	short, err := module.ShortingURL(string(b))
	if err != nil {
		h.logger.Info("Error shorting", zap.String("err", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	user, err := getUserIDFROMCookie(r)
	if err != nil {
		h.logger.Info("Error getUserID", zap.String("err", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	err = h.Storage.SetURL(r.Context(), user, short, string(b))

	var de *pckgstorage.DuplicationError
	status := http.StatusCreated
	res := fmt.Sprintf("%s/%s", h.BaseURL, short)
	if err != nil {
		if !errors.As(err, &de) {
			h.logger.Info("Error setURL", zap.String("err", err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			return
		} else {
			status = http.StatusConflict
			res = fmt.Sprintf("%s/%s", h.BaseURL, de.Duplication)
		}
	}
	h.logger.Info("SetURL", zap.String("long", string(b)), zap.String("short", short), zap.Int("status", status))
	w.WriteHeader(status)
	_, err = w.Write([]byte(res))
	if err != nil {
		h.logger.Error("PostURL write error", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

// GetURL получает сокращенную ссылку из URL. Возвращает полную ссылку и Redirect
func (h *Handler) GetURL(w http.ResponseWriter, r *http.Request) {
	short := strings.TrimLeft(r.URL.Path, "/")
	v, deleted := h.Storage.GetURL(r.Context(), short)
	if deleted {
		w.WriteHeader(http.StatusGone)
		return
	}
	if len(v) > 0 {
		w.Header().Set("Location", v)
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	}
	http.Error(w, "Wrong ID", http.StatusBadRequest)
}

// PingDB проверяет соединение с PostgreSQL
func (h *Handler) PingDB(w http.ResponseWriter, r *http.Request) {
	pinger, ok := h.Storage.(pckgstorage.Pinger)
	if ok {
		if pinger.Ping() == nil {
			w.WriteHeader(http.StatusOK)
			return
		}
	}
	//log.Fatal("Storage haven't pinger")
	w.WriteHeader(http.StatusInternalServerError)
}

// PostBatch получает список URL в JSON. Преобразует и отправляет в storage. Возвращает JSON c сокращенными URL и CorrelationID
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
		tmpShort, errInput := module.ShortingURL(url.Long)
		if errInput != nil {
			http.Error(w, errInput.Error(), http.StatusBadRequest)
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
		fmt.Println(tmpShort, ":", url.Long)
	}

	result := make([]batchResult, len(inputs))
	err = h.Storage.SetBatchURLs(r.Context(), urls)
	if err != nil {
		h.logger.Info(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
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
	w.WriteHeader(http.StatusCreated)
	w.Write(resJSON)
}

// PostJSON получает URL в JSON. Преобразует и отправляет в storage. Возвращает JSON c сокращенным URL
func (h *Handler) PostJSON(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	urlEnt := input{}
	if errUnmarshal := json.Unmarshal(b, &urlEnt); errUnmarshal != nil {
		http.Error(w, errUnmarshal.Error(), http.StatusBadRequest)
	}
	short, err := module.ShortingURL(urlEnt.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	user, err := getUserIDFROMCookie(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	err = h.Storage.SetURL(r.Context(), user, short, urlEnt.URL)
	res := result{}

	var de *pckgstorage.DuplicationError
	status := http.StatusCreated
	res.Result = fmt.Sprintf("%s/%s", h.BaseURL, short)
	if err != nil {
		if !errors.As(err, &de) {
			w.WriteHeader(http.StatusBadRequest)
			return
		} else {
			status = http.StatusConflict
			res.Result = fmt.Sprintf("%s/%s", h.BaseURL, de.Duplication)
		}
	}
	w.WriteHeader(status)
	resJSON, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	w.Write(resJSON)
}

// GetURLsByUser возвращает JSON с массивом ссылок, которые созданы текущим пользователем
func (h *Handler) GetURLsByUser(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	urls := h.Storage.GetURLsByUser(r.Context(), cookie.Value)
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

// GetInternalStats возвращает JSON со статистикой, если запрос идет из доверенных подсетей
func (h *Handler) GetInternalStats(w http.ResponseWriter, r *http.Request) {
	ip := r.Header.Get("X-Real-IP")
	netIP := net.ParseIP(ip)
	if netIP == nil {
		h.logger.Info("no x-real-ip header")
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if !h.trustedSubnet.Contains(netIP) {
		h.logger.Info("GetInternalStats from non trusted subnet", zap.String("IP", netIP.String()))
		http.Error(w, "Forbidden", http.StatusForbidden)
		w.WriteHeader(http.StatusForbidden)
		return
	}
	stats := &statsResult{}
	var err error

	stats.Users, err = h.Storage.GetUsersCount(r.Context())
	if err != nil {
		h.logger.Error("GetUsersCount error", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	stats.Urls, err = h.Storage.GetUrlsCount(r.Context())
	if err != nil {
		h.logger.Error("GetUrlsCount error", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	statsJSON, err := json.Marshal(stats)
	if err != nil {
		h.logger.Error("GetInternalStats marshal error", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = w.Write(statsJSON)
	if err != nil {
		h.logger.Error("GetInternalStats write error", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

}

// DecompressGZRequest middleware для работы с gzip
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

// CheckCookies middleware для проверки Cookie
func CheckCookies(secretKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookies := r.Cookies()
			if verifyCookies(cookies, secretKey) {

			} else {
				b := make([]byte, 16)
				_, _ = rand.Read(b)
				id := sha256.Sum256(b)
				key := []byte(secretKey)
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
}

func verifyCookies(cookies []*http.Cookie, secretKey string) bool {
	key := []byte(secretKey)
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
