package handler

import (
	"github.com/Spear5030/yapshrtnr/internal/config"
	testStorage "github.com/Spear5030/yapshrtnr/internal/storage"
	"github.com/Spear5030/yapshrtnr/pkg/logger"
	"github.com/stretchr/testify/require"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_GetInternalStats(t *testing.T) {
	cfg, err := config.New()
	require.NoError(t, err)
	lg, _ := logger.New(true)
	_, IPNet, _ := net.ParseCIDR("127.0.0.0/8")
	h := New(lg, testStorage.NewMemoryStorage(), cfg.BaseURL, cfg.Key, *IPNet)
	req := httptest.NewRequest("GET", "/api/internal/stats", nil)

	w := httptest.NewRecorder()
	h.GetInternalStats(w, req)
	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, w.Code)

	w = httptest.NewRecorder()
	req.Header.Set("X-Real-IP", "127.0.0.1")
	h.GetInternalStats(w, req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req.Header.Set("X-Real-IP", "192.168.0.1")
	h.GetInternalStats(w, req)
	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, w.Code)
}

func TestHandler_GetInternalStatsWithoutCIDR(t *testing.T) {
	cfg, err := config.New()
	require.NoError(t, err)
	lg, _ := logger.New(true)
	h := New(lg, testStorage.NewMemoryStorage(), cfg.BaseURL, cfg.Key, net.IPNet(cfg.TrustedSubnet))
	req := httptest.NewRequest("GET", "/api/internal/stats", nil)

	w := httptest.NewRecorder()
	h.GetInternalStats(w, req)
	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, w.Code)

	w = httptest.NewRecorder()
	req.Header.Set("X-Real-IP", "127.0.0.1")
	h.GetInternalStats(w, req)
	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, w.Code)

	w = httptest.NewRecorder()
	req.Header.Set("X-Real-IP", "192.168.0.1")
	h.GetInternalStats(w, req)
	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, w.Code)

}
