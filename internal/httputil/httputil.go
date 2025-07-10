package httputil

import "github.com/gin-gonic/gin"

// JSONError sends a JSON response with a standard error structure.
func JSONError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}

// AbortJSONError aborts the context and sends a JSON error response.
func AbortJSONError(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, gin.H{"error": message})
}
