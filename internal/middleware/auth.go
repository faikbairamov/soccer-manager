package middleware

import (
	"net/http"
	"strings"

	"github.com/faikbairamov/soccer-manager/internal/auth"
	"github.com/faikbairamov/soccer-manager/internal/i18n"
	"github.com/gin-gonic/gin"
)

func Middleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": i18n.Translate(c, "auth.unauthorized")})
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		userID, err := auth.ValidateToken(tokenStr, secret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": i18n.Translate(c, "auth.unauthorized")})
			return
		}
		c.Set("userID", userID)
		c.Next()
	}
}
