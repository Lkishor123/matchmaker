package main

import (
	"github.com/gin-gonic/gin"

	"matchmaker/internal/handlers"
)

func main() {
	r := gin.Default()
	r.GET("/ping", handlers.Ping)

	r.Run()
}
