package handlers

import "github.com/gin-gonic/gin"

// Ping is a simple handler used across services.
func Ping(c *gin.Context) {
	c.JSON(200, gin.H{"message": "pong"})
}
