package admin

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/langowen/bodybalance-backend/internal/config"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/logdiscart"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func makeJWTToken(signingKey string, isAdmin bool, username string) string {
	claims := Claims{
		Username: username,
		IsAdmin:  isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, _ := token.SignedString([]byte(signingKey))
	return t
}

func TestAuthMiddleware_NoCookie(t *testing.T) {
	h := &Handler{
		logger: logdiscart.NewDiscardLogger(),
		cfg:    &config.Config{HTTPServer: config.HTTPServer{SigningKey: "testkey"}},
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	called := false
	h.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.False(t, called)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	h := &Handler{
		logger: logdiscart.NewDiscardLogger(),
		cfg:    &config.Config{HTTPServer: config.HTTPServer{SigningKey: "testkey"}},
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: "invalidtoken"})
	w := httptest.NewRecorder()

	called := false
	h.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.False(t, called)
}

func TestAuthMiddleware_NotAdmin(t *testing.T) {
	h := &Handler{
		logger: logdiscart.NewDiscardLogger(),
		cfg:    &config.Config{HTTPServer: config.HTTPServer{SigningKey: "testkey"}},
	}
	token := makeJWTToken("testkey", false, "user1")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	w := httptest.NewRecorder()

	called := false
	h.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})).ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.False(t, called)
}

func TestAuthMiddleware_AdminOK(t *testing.T) {
	h := &Handler{
		logger: logdiscart.NewDiscardLogger(),
		cfg:    &config.Config{HTTPServer: config.HTTPServer{SigningKey: "testkey"}},
	}
	token := makeJWTToken("testkey", true, "adminuser")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	w := httptest.NewRecorder()

	called := false
	h.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, called)
}
