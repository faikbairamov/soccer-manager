package handler

import (
	"context"
	"net/http"

	"github.com/faikbairamov/soccer-manager/internal/domain"
	"github.com/faikbairamov/soccer-manager/internal/i18n"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type teamService interface {
	GetTeam(ctx context.Context, userID uuid.UUID) (domain.Team, error)
	UpdateTeam(ctx context.Context, userID uuid.UUID, name, country string) (domain.Team, error)
}

type TeamHandler struct {
	svc teamService
}

func NewTeamHandler(svc teamService) *TeamHandler {
	return &TeamHandler{svc: svc}
}

type UpdateTeamRequest struct {
	Name    string `json:"name"`
	Country string `json:"country"`
}

func (h *TeamHandler) GetTeam(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	team, err := h.svc.GetTeam(c.Request.Context(), userID)
	if err != nil {
		httpError(c, err, "team.not_found")
		return
	}
	c.JSON(http.StatusOK, team)
}

func (h *TeamHandler) UpdateTeam(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var req UpdateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorBody(i18n.Translate(c, "validation.required")))
		return
	}

	team, err := h.svc.UpdateTeam(c.Request.Context(), userID, req.Name, req.Country)
	if err != nil {
		httpError(c, err, "team.not_found")
		return
	}
	c.JSON(http.StatusOK, team)
}
