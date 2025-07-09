package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"matchmaker/internal/handlers"
)

func main() {
	r := gin.Default()
	r.GET("/ping", handlers.Ping)

	// Placeholder websocket usage to reference the library.
	_ = websocket.Upgrader{}

	r.Run()
}
