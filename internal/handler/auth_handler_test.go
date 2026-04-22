package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/faikbairamov/soccer-manager/internal/domain"
	"github.com/faikbairamov/soccer-manager/internal/i18n"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubAuthService struct {
	registerFn func(ctx context.Context, email, password string) (string, error)
	loginFn    func(ctx context.Context, email, password string) (string, error)
}

func (s *stubAuthService) Register(ctx context.Context, email, password string) (string, error) {
	return s.registerFn(ctx, email, password)
}

func (s *stubAuthService) Login(ctx context.Context, email, password string) (string, error) {
	return s.loginFn(ctx, email, password)
}

func newAuthRouter(svc authService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(i18n.Middleware())
	h := NewAuthHandler(svc)
	r.POST("/auth/register", h.Register)
	r.POST("/auth/login", h.Login)
	return r
}

func postJSON(r *gin.Engine, path string, body any) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func TestRegisterHandler_InvalidEmail(t *testing.T) {
	w := postJSON(newAuthRouter(nil), "/auth/register",
		map[string]string{"email": "not-an-email", "password": "password123"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegisterHandler_PasswordTooShort(t *testing.T) {
	w := postJSON(newAuthRouter(nil), "/auth/register",
		map[string]string{"email": "test@example.com", "password": "short"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegisterHandler_DuplicateEmail(t *testing.T) {
	svc := &stubAuthService{
		registerFn: func(_ context.Context, _, _ string) (string, error) {
			return "", domain.ErrConflict
		},
	}
	w := postJSON(newAuthRouter(svc), "/auth/register",
		map[string]string{"email": "dup@example.com", "password": "password123"})
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestRegisterHandler_Success(t *testing.T) {
	svc := &stubAuthService{
		registerFn: func(_ context.Context, _, _ string) (string, error) {
			return "jwt-token", nil
		},
	}
	w := postJSON(newAuthRouter(svc), "/auth/register",
		map[string]string{"email": "new@example.com", "password": "password123"})
	require.Equal(t, http.StatusCreated, w.Code)

	var resp AuthResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "jwt-token", resp.Token)
}

func TestLoginHandler_WrongPassword(t *testing.T) {
	svc := &stubAuthService{
		loginFn: func(_ context.Context, _, _ string) (string, error) {
			return "", domain.ErrUnauthorized
		},
	}
	w := postJSON(newAuthRouter(svc), "/auth/login",
		map[string]string{"email": "test@example.com", "password": "wrong"})
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLoginHandler_Success(t *testing.T) {
	svc := &stubAuthService{
		loginFn: func(_ context.Context, _, _ string) (string, error) {
			return "jwt-token", nil
		},
	}
	w := postJSON(newAuthRouter(svc), "/auth/login",
		map[string]string{"email": "test@example.com", "password": "password123"})
	require.Equal(t, http.StatusOK, w.Code)

	var resp AuthResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "jwt-token", resp.Token)
}
