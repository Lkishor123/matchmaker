package main

import (
	"github.com/gorilla/websocket"

	"matchmaker/internal/handlers"
	"matchmaker/internal/logging"
)

func main() {
	logging.Init()
	r := logging.NewGinEngine()
	r.GET("/ping", handlers.Ping)

	// Placeholder websocket usage to reference the library.
	_ = websocket.Upgrader{}

	r.Run()
}
