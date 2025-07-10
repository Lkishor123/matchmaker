package main

import (
	"matchmaker/internal/handlers"
	"matchmaker/internal/logging"
)

func main() {
	logging.Init()
	r := logging.NewGinEngine()
	r.GET("/ping", handlers.Ping)
	r.POST("/api/v1/analysis", handlers.CreateAnalysis)

	r.Run()
}
