package handler

import (
	"context"
	"net/http"

	"github.com/faikbairamov/soccer-manager/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type playerService interface {
	GetPlayer(ctx context.Context, id uuid.UUID) (domain.Player, error)
	UpdatePlayer(ctx context.Context, userID, playerID uuid.UUID, firstName, lastName, country string) (domain.Player, error)
}

type PlayerHandler struct {
	svc playerService
}

func NewPlayerHandler(svc playerService) *PlayerHandler {
	return &PlayerHandler{svc: svc}
}

type UpdatePlayerRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Country   string `json:"country"`
}

func (h *PlayerHandler) GetPlayer(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorBody("invalid id"))
		return
	}

	player, err := h.svc.GetPlayer(c.Request.Context(), id)
	if err != nil {
		httpError(c, err, "player.not_found")
		return
	}
	c.JSON(http.StatusOK, player)
}

func (h *PlayerHandler) UpdatePlayer(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	playerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorBody("invalid id"))
		return
	}

	var req UpdatePlayerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorBody(err.Error()))
		return
	}

	player, err := h.svc.UpdatePlayer(c.Request.Context(), userID, playerID, req.FirstName, req.LastName, req.Country)
	if err != nil {
		httpError(c, err, "player.not_owner")
		return
	}
	c.JSON(http.StatusOK, player)
}
