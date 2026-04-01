package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS permite requisições de outras origens (ex.: frontend em outro host/porta).
func CORS() gin.HandlerFunc {
	allowed := parseOrigins(os.Getenv("CORS_ALLOW_ORIGINS"))
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && isAllowedOrigin(allowed, origin) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func parseOrigins(v string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, part := range strings.Split(v, ",") {
		o := strings.TrimSpace(part)
		if o == "" {
			continue
		}
		out[o] = struct{}{}
	}
	return out
}

func isAllowedOrigin(allowed map[string]struct{}, origin string) bool {
	if len(allowed) == 0 {
		return false
	}
	_, ok := allowed[origin]
	return ok
}
