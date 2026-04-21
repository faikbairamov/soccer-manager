package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/faikbairamov/soccer-manager/internal/domain"
	"github.com/faikbairamov/soccer-manager/internal/i18n"
	"github.com/gin-gonic/gin"
)

func httpError(c *gin.Context, err error, key string) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		c.JSON(http.StatusNotFound, errorBody(i18n.Translate(c, key)))
	case errors.Is(err, domain.ErrUnauthorized):
		c.JSON(http.StatusUnauthorized, errorBody(i18n.Translate(c, key)))
	case errors.Is(err, domain.ErrForbidden):
		c.JSON(http.StatusForbidden, errorBody(i18n.Translate(c, key)))
	case errors.Is(err, domain.ErrConflict):
		c.JSON(http.StatusConflict, errorBody(i18n.Translate(c, key)))
	case errors.Is(err, domain.ErrOwnPlayer):
		c.JSON(http.StatusUnprocessableEntity, errorBody(i18n.Translate(c, key)))
	case errors.Is(err, domain.ErrInsufficientFunds):
		c.JSON(http.StatusUnprocessableEntity, errorBody(i18n.Translate(c, key)))
	case errors.Is(err, domain.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, errorBody(i18n.Translate(c, key)))
	default:
		slog.ErrorContext(c.Request.Context(), "unexpected error", "err", err)
		c.JSON(http.StatusInternalServerError, errorBody(i18n.Translate(c, "server.internal_error")))
	}
}

func errorBody(msg string) gin.H {
	return gin.H{"error": msg}
}
