package main

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"

	"matchmaker/internal/handlers"
)

func main() {
	r := gin.Default()
	r.GET("/ping", handlers.Ping)

	// Example usage of jwt-go to ensure dependency is referenced.
	_ = jwt.New(jwt.SigningMethodHS256)

	r.Run()
}
