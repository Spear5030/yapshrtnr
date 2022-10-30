package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	testStorage "internal/storage"
)

func TestHandler_ServeHTTP(t *testing.T) {
	type want struct {
		statusCode int
	}
	tests := []struct {
		name    string
		request string
		method  string
		want    want
	}{
		{
			name:   "simple test #1 wrong method",
			method: http.MethodPatch,
			want: want{
				statusCode: 400,
			},
			request: "/gcfCxYsh",
		},
		{
			name:   "simple test #2 get method",
			method: http.MethodGet,
			want: want{
				statusCode: 307,
			},
			request: "/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.request, nil)
			w := httptest.NewRecorder()
			h := Handler{testStorage.NewStorage()}
			h.ServeHTTP(w, request)
			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
		})
	}
}
