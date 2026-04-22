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

// GetPlayer godoc
// @Summary      Get a player
// @Description  Returns a single player by UUID. Any authenticated user can view any player.
// @Tags         players
// @Produce      json
// @Security     BearerAuth
// @Param        id  path     string true "Player UUID"
// @Success      200 {object} domain.Player
// @Failure      400 {object} ErrorResponse "Invalid UUID"
// @Failure      401 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Router       /players/{id} [get]
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

// UpdatePlayer godoc
// @Summary      Update a player
// @Description  Updates first name, last name, and country of a player you own.
// @Tags         players
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path     string              true "Player UUID"
// @Param        request body     UpdatePlayerRequest true "Player update payload"
// @Success      200     {object} domain.Player
// @Failure      400     {object} ErrorResponse "Invalid UUID or body"
// @Failure      401     {object} ErrorResponse
// @Failure      403     {object} ErrorResponse "Player belongs to another team"
// @Failure      404     {object} ErrorResponse
// @Router       /players/{id} [patch]
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
