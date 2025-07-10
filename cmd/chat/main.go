package main

import (
	"matchmaker/internal/config"
	"matchmaker/internal/database"
	"matchmaker/internal/handlers"
	"matchmaker/internal/logging"
)

func main() {
	logging.Init()
	if _, err := config.LoadChat(); err != nil {
		logging.Log.Fatal(err)
	}
	if _, err := database.InitRedis(); err != nil {
		logging.Log.Fatal("redis initialization failed")
	}
	r := logging.NewGinEngine()
	r.GET("/ping", handlers.Ping)

	api := r.Group("/api/v1")
	api.Use(handlers.RequireUserID())
	api.GET("/chat", handlers.Chat)

	r.Run()
}
