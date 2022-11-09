package router

import (
	"github.com/Spear5030/yapshrtnr/internal/handler"
	testStorage "github.com/Spear5030/yapshrtnr/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
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
	//cfg, _ := config.New()
	h := handler.New(testStorage.New(), "localhost:8080")
	r := New(h)
	ts := httptest.NewServer(r)
	defer ts.Close()
	h.Storage.SetURL("tt123456", "http://ya.ru")

	statusCode, body := testRequest(t, ts, "POST", "/", "http://longlonglong.lg")
	assert.Equal(t, http.StatusCreated, statusCode)
	assert.NotEmpty(t, body)

	statusCode, _ = testRequest(t, ts, "GET", "/tt123456", "")
	assert.Equal(t, http.StatusOK, statusCode)

	statusCode, _ = testRequest(t, ts, "GET", "/", "")
	assert.Equal(t, http.StatusMethodNotAllowed, statusCode)

}
