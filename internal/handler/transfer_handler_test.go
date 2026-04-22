package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/faikbairamov/soccer-manager/internal/domain"
	"github.com/faikbairamov/soccer-manager/internal/i18n"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type stubTransferService struct {
	buyPlayerFn func(ctx context.Context, userID, transferID uuid.UUID) error
}

func (s *stubTransferService) BuyPlayer(ctx context.Context, userID, transferID uuid.UUID) error {
	return s.buyPlayerFn(ctx, userID, transferID)
}
func (s *stubTransferService) ListTransfer(ctx context.Context, userID, playerID uuid.UUID, price int64) (domain.Transfer, error) {
	return domain.Transfer{}, nil
}
func (s *stubTransferService) GetTransfers(ctx context.Context, page, limit int) ([]domain.Transfer, int64, error) {
	return nil, 0, nil
}
func (s *stubTransferService) DelistTransfer(ctx context.Context, userID, transferID uuid.UUID) error {
	return nil
}
func newTransferRouter(svc transferService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(i18n.Middleware())
	r.Use(func(c *gin.Context) {
		c.Set("userID", uuid.UUID{1})
		c.Next()
	})
	h := NewTransferHandler(svc)
	r.POST("/transfers/:id/buy", h.BuyPlayer)
	r.POST("/transfers", h.ListTransfer)
	r.GET("/transfers", h.GetTransfers)
	r.DELETE("/transfers/:id", h.DelistTransfer)
	return r
}
func TestBuyPlayerHandler_InsufficientFunds(t *testing.T) {
	svc := &stubTransferService{
		buyPlayerFn: func(_ context.Context, _, _ uuid.UUID) error {
			return domain.ErrInsufficientFunds
		},
	}
	r := newTransferRouter(svc)
	transferID := uuid.New()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/transfers/%s/buy", transferID), nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}
func TestBuyPlayerHandler_OwnPlayer(t *testing.T) {
	svc := &stubTransferService{
		buyPlayerFn: func(_ context.Context, _, _ uuid.UUID) error {
			return domain.ErrOwnPlayer
		},
	}
	r := newTransferRouter(svc)
	transferID := uuid.New()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/transfers/%s/buy", transferID), nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}
func TestBuyPlayerHandler_Success(t *testing.T) {
	svc := &stubTransferService{
		buyPlayerFn: func(_ context.Context, _, _ uuid.UUID) error { return nil },
	}
	r := newTransferRouter(svc)
	transferID := uuid.New()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/transfers/%s/buy", transferID), nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
