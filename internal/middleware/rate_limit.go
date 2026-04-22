package middleware

import (
	"net/http"
	"sync"

	"github.com/faikbairamov/soccer-manager/internal/i18n"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func RateLimitMiddleware(r rate.Limit, burst int) gin.HandlerFunc {
	var mu sync.Mutex
	limiters := make(map[string]*rate.Limiter)

	getLimiter := func(ip string) *rate.Limiter {
		mu.Lock()
		defer mu.Unlock()
		if l, ok := limiters[ip]; ok {
			return l
		}
		l := rate.NewLimiter(r, burst)
		limiters[ip] = l
		return l
	}
	return func(c *gin.Context) {
		if !getLimiter(c.ClientIP()).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": i18n.Translate(c, "error.rate_limit")})
			return
		}
		c.Next()
	}
}
