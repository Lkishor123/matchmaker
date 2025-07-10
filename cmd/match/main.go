package main

import (
	"matchmaker/internal/config"
	"matchmaker/internal/handlers"
	"matchmaker/internal/logging"
)

func main() {
	logging.Init()
	if _, err := config.LoadMatch(); err != nil {
		logging.Log.Fatal(err)
	}
	r := logging.NewGinEngine()
	r.GET("/ping", handlers.Ping)
	r.POST("/api/v1/analysis", handlers.CreateAnalysis)

	r.Run()
}
