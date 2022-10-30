package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	testStorage "github.com/Spear5030/yapshrtnr/internal/storage"
	"github.com/stretchr/testify/assert"
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.request, nil)
			w := httptest.NewRecorder()
			h := Handler{testStorage.NewStorage()}

			h.ServeHTTP(w, request)
			result := w.Result()
			defer result.Body.Close()
			//resBody, err := io.ReadAll(result.Body)
			assert.Equal(t, tt.want.statusCode, result.StatusCode)
		})
	}
}
