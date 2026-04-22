package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/faikbairamov/soccer-manager/internal/domain"
	"github.com/faikbairamov/soccer-manager/internal/i18n"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubTeamService struct {
	getTeamFn    func(ctx context.Context, userID uuid.UUID) (domain.Team, error)
	updateTeamFn func(ctx context.Context, userID uuid.UUID, name, country string) (domain.Team, error)
}

func (s *stubTeamService) GetTeam(ctx context.Context, userID uuid.UUID) (domain.Team, error) {
	return s.getTeamFn(ctx, userID)
}
func (s *stubTeamService) UpdateTeam(ctx context.Context, userID uuid.UUID, name, country string) (domain.Team, error) {
	return s.updateTeamFn(ctx, userID, name, country)
}

func newAuthedRouter(method, path string, handler gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(i18n.Middleware())
	r.Use(func(c *gin.Context) {
		c.Set("userID", uuid.UUID{1})
		c.Next()
	})
	r.Handle(method, path, handler)
	return r
}

func TestGetTeamHandler_Success(t *testing.T) {
	svc := &stubTeamService{
		getTeamFn: func(_ context.Context, _ uuid.UUID) (domain.Team, error) {
			return domain.Team{Name: "Test FC", Budget: 5_000_000, TotalValue: 20_000_000}, nil
		},
	}
	h := NewTeamHandler(svc)
	r := newAuthedRouter(http.MethodGet, "/teams/me", h.GetTeam)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/teams/me", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "Test FC", body["Name"])
}

func TestGetTeamHandler_NotFound(t *testing.T) {
	svc := &stubTeamService{
		getTeamFn: func(_ context.Context, _ uuid.UUID) (domain.Team, error) {
			return domain.Team{}, domain.ErrNotFound
		},
	}
	h := NewTeamHandler(svc)
	r := newAuthedRouter(http.MethodGet, "/teams/me", h.GetTeam)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/teams/me", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
