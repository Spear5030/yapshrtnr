package handler

import (
	"github.com/Spear5030/yapshrtnr/internal/config"
	testStorage "github.com/Spear5030/yapshrtnr/internal/storage"
	"github.com/Spear5030/yapshrtnr/pkg/logger"
	"net/http/httptest"
	"testing"
)

func BenchmarkHandler_PostURLMemory(b *testing.B) {
	cfg, _ := config.New()
	lg, _ := logger.New(true)
	h := New(lg, testStorage.NewMemoryStorage(), cfg.BaseURL, cfg.Key)
	r := httptest.NewRequest("Post", "/", nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		w := httptest.NewRecorder()
		b.StartTimer()
		h.PostURL(w, r)
	}
}

func BenchmarkHandler_PostURLFile(b *testing.B) {
	cfg, _ := config.New()
	lg, _ := logger.New(true)
	s, _ := testStorage.NewFileStorage("bench.base")
	h := New(lg, s, cfg.BaseURL, cfg.Key)
	r := httptest.NewRequest("Post", "/", nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		w := httptest.NewRecorder()
		b.StartTimer()
		h.PostURL(w, r)
	}
}
