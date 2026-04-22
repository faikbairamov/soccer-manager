package handler

import (
	"context"
	"net/http"

	"github.com/faikbairamov/soccer-manager/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type transferService interface {
	ListTransfer(ctx context.Context, userID, playerID uuid.UUID, askingPrice int64) (domain.Transfer, error)
	GetTransfers(ctx context.Context, page, limit int) ([]domain.Transfer, int64, error)
	DelistTransfer(ctx context.Context, userID, transferID uuid.UUID) error
	BuyPlayer(ctx context.Context, userID, transferID uuid.UUID) error
}
type TransferHandler struct {
	svc transferService
}

func NewTransferHandler(svc transferService) *TransferHandler {
	return &TransferHandler{svc: svc}
}

type ListTransferRequest struct {
	PlayerID    string `json:"player_id"    binding:"required,uuid"`
	AskingPrice int64  `json:"asking_price" binding:"required,min=1"`
}
type GetTransfersQuery struct {
	Page  int `form:"page"`
	Limit int `form:"limit"`
}

func (h *TransferHandler) ListTransfer(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	var req ListTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorBody(err.Error()))
		return
	}
	playerID, _ := uuid.Parse(req.PlayerID)
	transfer, err := h.svc.ListTransfer(c.Request.Context(), userID, playerID, req.AskingPrice)
	if err != nil {
		httpError(c, err, "transfer.already_listed")
		return
	}
	c.JSON(http.StatusCreated, transfer)
}
func (h *TransferHandler) GetTransfers(c *gin.Context) {
	var q GetTransfersQuery
	_ = c.ShouldBindQuery(&q)
	if q.Page == 0 {
		q.Page = 1
	}
	if q.Limit == 0 {
		q.Limit = 20
	}
	transfers, total, err := h.svc.GetTransfers(c.Request.Context(), q.Page, q.Limit)
	if err != nil {
		httpError(c, err, "server.internal_error")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": transfers,
		"meta": gin.H{"page": q.Page, "limit": q.Limit, "total": total},
	})
}
func (h *TransferHandler) DelistTransfer(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	transferID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorBody("invalid id"))
		return
	}
	if err := h.svc.DelistTransfer(c.Request.Context(), userID, transferID); err != nil {
		httpError(c, err, "transfer.not_found")
		return
	}
	c.Status(http.StatusNoContent)
}
func (h *TransferHandler) BuyPlayer(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	transferID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorBody("invalid id"))
		return
	}
	if err := h.svc.BuyPlayer(c.Request.Context(), userID, transferID); err != nil {
		httpError(c, err, "transfer.not_found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Player purchased successfully"})
}
