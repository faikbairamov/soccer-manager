package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/faikbairamov/soccer-manager/internal/i18n"
	"github.com/gin-gonic/gin"
	validator "github.com/go-playground/validator/v10"
)

type authService interface {
	Register(ctx context.Context, email, password string) (string, error)
	Login(ctx context.Context, email, password string) (string, error)
}

type AuthHandler struct {
	svc authService
}

func NewAuthHandler(svc authService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// Register godoc
// @Summary      Register a new user
// @Description  Creates a user account, seeds a team with 20 players, and returns a JWT.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body     RegisterRequest true "Registration payload"
// @Success      201     {object} AuthResponse
// @Failure      400     {object} ErrorResponse
// @Failure      409     {object} ErrorResponse "Email already taken"
// @Failure      429     {object} ErrorResponse "Rate limit exceeded"
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) && len(ve) > 0 {
			switch ve[0].Field() {
			case "Email":
				c.JSON(http.StatusBadRequest, errorBody(i18n.Translate(c, "validation.invalid_email")))
			case "Password":
				c.JSON(http.StatusBadRequest, errorBody(i18n.Translate(c, "validation.password_too_short")))
			default:
				c.JSON(http.StatusBadRequest, errorBody(i18n.Translate(c, "server.internal_error")))
			}
			return
		}
		c.JSON(http.StatusBadRequest, errorBody(err.Error()))
		return
	}

	token, err := h.svc.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		httpError(c, err, "auth.email_taken")
		return
	}
	c.JSON(http.StatusCreated, AuthResponse{Token: token})
}

// Login godoc
// @Summary      Log in
// @Description  Validates credentials and returns a JWT.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body     LoginRequest true "Login payload"
// @Success      200     {object} AuthResponse
// @Failure      400     {object} ErrorResponse
// @Failure      401     {object} ErrorResponse "Invalid credentials"
// @Failure      429     {object} ErrorResponse "Rate limit exceeded"
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorBody(err.Error()))
		return
	}

	token, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		httpError(c, err, "auth.invalid_credentials")
		return
	}
	c.JSON(http.StatusOK, AuthResponse{Token: token})
}
