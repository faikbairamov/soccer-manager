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

// TransferListResponse is returned by GetTransfers.
type TransferListResponse struct {
	Data []domain.Transfer `json:"data"`
	Meta struct {
		Page  int   `json:"page"`
		Limit int   `json:"limit"`
		Total int64 `json:"total"`
	} `json:"meta"`
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

// ListTransfer godoc
// @Summary      List a player on the transfer market
// @Description  Creates a transfer listing for a player you own. Player must not already be listed.
// @Tags         transfers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body     ListTransferRequest true "Listing payload"
// @Success      201     {object} domain.Transfer
// @Failure      400     {object} ErrorResponse
// @Failure      401     {object} ErrorResponse
// @Failure      403     {object} ErrorResponse "Player belongs to another team"
// @Failure      409     {object} ErrorResponse "Player already listed"
// @Router       /transfers [post]
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
// GetTransfers godoc
// @Summary      Browse transfer market
// @Description  Returns a paginated list of all active transfer listings.
// @Tags         transfers
// @Produce      json
// @Security     BearerAuth
// @Param        page  query    int false "Page number (default 1)"
// @Param        limit query    int false "Results per page (default 20)"
// @Success      200   {object} TransferListResponse
// @Failure      401   {object} ErrorResponse
// @Router       /transfers [get]
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
// DelistTransfer godoc
// @Summary      Remove a transfer listing
// @Description  Deletes a transfer listing. Only the owner of the listed player can delist.
// @Tags         transfers
// @Produce      json
// @Security     BearerAuth
// @Param        id  path string true "Transfer UUID"
// @Success      204
// @Failure      400 {object} ErrorResponse "Invalid UUID"
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse "Not the owner"
// @Failure      404 {object} ErrorResponse
// @Router       /transfers/{id} [delete]
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
// BuyPlayer godoc
// @Summary      Buy a listed player
// @Description  Purchases a player from the transfer market. Atomically credits the seller, debits the buyer, transfers the player, and removes the listing.
// @Tags         transfers
// @Produce      json
// @Security     BearerAuth
// @Param        id  path     string true "Transfer UUID"
// @Success      200 {object} map[string]string
// @Failure      400 {object} ErrorResponse "Invalid UUID"
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse "Cannot buy your own player"
// @Failure      404 {object} ErrorResponse
// @Failure      422 {object} ErrorResponse "Insufficient budget"
// @Router       /transfers/{id}/buy [post]
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
