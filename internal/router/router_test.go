package router

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"github.com/Spear5030/yapshrtnr/internal/config"
	"github.com/Spear5030/yapshrtnr/internal/handler"
	testStorage "github.com/Spear5030/yapshrtnr/internal/storage"
	"github.com/Spear5030/yapshrtnr/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path, body string) (int, string) {
	req, err := http.NewRequest(method, ts.URL+path, strings.NewReader(body))
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	defer resp.Body.Close()

	return resp.StatusCode, string(respBody)
}

func TestRouter(t *testing.T) {
	cfg, err := config.New()
	require.NoError(t, err)
	lg, _ := logger.New(true)
	h := handler.New(lg, testStorage.NewMemoryStorage(), cfg.BaseURL, cfg.Key, net.IPNet(cfg.TrustedSubnet))
	r := New(h)
	ts := httptest.NewServer(r)
	defer ts.Close()
	h.Storage.SetURL(context.Background(), "user1", "tt123456", "http://ya.ru")

	statusCode, body := testRequest(t, ts, "POST", "/", "http://longlonglong.lg")
	assert.Equal(t, http.StatusCreated, statusCode)
	assert.NotEmpty(t, body)

	statusCode, _ = testRequest(t, ts, "GET", "/tt123456", "")
	assert.Equal(t, http.StatusOK, statusCode)

	statusCode, _ = testRequest(t, ts, "GET", "/", "")
	assert.Equal(t, http.StatusMethodNotAllowed, statusCode)

	statusCode, body = testRequest(t, ts, "POST", "/api/shorten", "{\"url\":\"http://longlonglong.lg\"}")
	assert.Equal(t, http.StatusCreated, statusCode)
	assert.NotEmpty(t, body)
}

func TestGZRequest(t *testing.T) {
	cfg, err := config.New()
	require.NoError(t, err)
	lg, _ := logger.New(true)
	h := handler.New(lg, testStorage.NewMemoryStorage(), cfg.BaseURL, cfg.Key, net.IPNet(cfg.TrustedSubnet))
	r := New(h)
	ts := httptest.NewServer(r)
	defer ts.Close()
	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	_, _ = gzw.Write([]byte("https://ya.ru"))
	_ = gzw.Close()
	req, err := http.NewRequest("POST", ts.URL+"/", bytes.NewReader(buf.Bytes()))
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Encoding", "gzip")
	require.NoError(t, err)
}

func TestJSON(t *testing.T) {
	cfg, err := config.New()
	require.NoError(t, err)
	lg, _ := logger.New(true)
	h := handler.New(lg, testStorage.NewMemoryStorage(), cfg.BaseURL, cfg.Key, net.IPNet(cfg.TrustedSubnet))
	r := New(h)
	ts := httptest.NewServer(r)
	defer ts.Close()
	req, err := http.NewRequest("POST", ts.URL+"/api/shorten", strings.NewReader("{\"url\":\"http://yandex.ru\"}"))
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	defer resp.Body.Close()

	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	assert.Equal(t, true, json.Valid(respBody))

}
