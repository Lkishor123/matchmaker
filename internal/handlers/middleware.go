package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"

	"matchmaker/internal/httputil"
	"matchmaker/internal/logging"
)

// RequireUserID parses the JWT in the Authorization header and stores the user_id claim in the context.
func RequireUserID() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			httputil.AbortJSONError(c, http.StatusUnauthorized, "missing bearer token")
			return
		}
		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		claims := jwt.MapClaims{}
		if _, _, err := new(jwt.Parser).ParseUnverified(tokenStr, claims); err != nil {
			logging.Log.WithError(err).Warn("failed to parse jwt")
			httputil.AbortJSONError(c, http.StatusUnauthorized, "invalid token")
			return
		}
		id, ok := claims["user_id"].(float64)
		if !ok {
			logging.Log.Warn("user_id claim missing")
			httputil.AbortJSONError(c, http.StatusUnauthorized, "invalid token")
			return
		}
		c.Set("user_id", uint(id))
		c.Next()
	}
}
