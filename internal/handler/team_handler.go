package handler

import (
	"context"
	"errors"
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

// GetTeam godoc
// @Summary      Get my team
// @Description  Returns the authenticated user's team including total player value.
// @Tags         teams
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} domain.Team
// @Failure      401 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Router       /teams/me [get]
func (h *TeamHandler) GetTeam(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	team, err := h.svc.GetTeam(c.Request.Context(), userID)
	if err != nil {
		httpError(c, err, "team.not_found")
		return
	}
	c.JSON(http.StatusOK, team)
}

// UpdateTeam godoc
// @Summary      Update my team
// @Description  Updates the name and/or country of the authenticated user's team.
// @Tags         teams
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body     UpdateTeamRequest true "Team update payload"
// @Success      200     {object} domain.Team
// @Failure      400     {object} ErrorResponse
// @Failure      401     {object} ErrorResponse
// @Failure      422     {object} ErrorResponse "Empty name or country"
// @Router       /teams/me [patch]
func (h *TeamHandler) UpdateTeam(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var req UpdateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorBody(i18n.Translate(c, "validation.required")))
		return
	}

	team, err := h.svc.UpdateTeam(c.Request.Context(), userID, req.Name, req.Country)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			httpError(c, err, "team.invalid_input")
		default:
			httpError(c, err, "team.not_found")
		}
		return
	}
	c.JSON(http.StatusOK, team)
}
