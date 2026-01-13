package docs_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/langowen/bodybalance-backend/internal/port/http-server/handler/docs"
	"github.com/stretchr/testify/assert"
)

func TestDocsEndpointsAccessible(t *testing.T) {
	r := docs.Routes()

	tests := []struct {
		name       string
		path       string
		wantStatus int
		checkBody  bool
	}{
		{
			name:       "корневой путь документации",
			path:       "/swagger/",
			wantStatus: http.StatusOK,
			checkBody:  true,
		},
		{
			name:       "swagger JSON спецификация",
			path:       "/swagger/doc.json",
			wantStatus: http.StatusOK,
			checkBody:  true,
		},
		{
			name:       "несуществующий файл",
			path:       "/swagger/nonexistent",
			wantStatus: http.StatusNotFound,
			checkBody:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code, "status code")

			if tt.checkBody && tt.wantStatus == http.StatusOK {
				assert.NotEmpty(t, rec.Body.String(), "body should not be empty")
			}
		})
	}
}
