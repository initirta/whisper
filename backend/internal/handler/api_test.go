package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"whisper/backend/internal/config"
)

func testRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	cfg := config.Config{FrontendOrigin: "http://localhost:3000"}
	api := NewAPI(nil, nil, nil, nil, cfg)
	r := gin.New()
	r.Use(api.CORS())
	api.RegisterRoutes(r)
	return r
}

func TestPrivateRoutesRejectMissingCredentials(t *testing.T) {
	r := testRouter()
	for _, path := range []string{"/api/v1/admin/me", "/api/v1/admin/conversations", "/api/v1/conversations/00000000-0000-4000-8000-000000000001/messages"} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, path, nil)
		r.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusUnauthorized {
			t.Errorf("GET %s: got %d, want 401", path, recorder.Code)
		}
	}
}

func TestLoginRequiresConfiguredOrigin(t *testing.T) {
	r := testRouter()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/auth/login", nil)
	request.Header.Set("Origin", "https://attacker.example")
	r.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("got %d, want 403", recorder.Code)
	}
}

func TestCORSAllowsOnlyConfiguredOrigin(t *testing.T) {
	r := testRouter()
	for _, tc := range []struct {
		origin string
		want   int
	}{{"http://localhost:3000", http.StatusNoContent}, {"https://attacker.example", http.StatusForbidden}} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodOptions, "/api/v1/conversations/x/messages", nil)
		request.Header.Set("Origin", tc.origin)
		r.ServeHTTP(recorder, request)
		if recorder.Code != tc.want {
			t.Errorf("origin %s: got %d, want %d", tc.origin, recorder.Code, tc.want)
		}
	}
}
