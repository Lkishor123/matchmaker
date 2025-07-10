package main

import (
	"matchmaker/internal/config"
	"matchmaker/internal/database"
	"matchmaker/internal/handlers"
	"matchmaker/internal/logging"
)

func main() {
	logging.Init()
	if _, err := config.LoadReport(); err != nil {
		logging.Log.Fatal(err)
	}
	if _, err := database.InitMongo(); err != nil {
		logging.Log.Fatal("mongodb initialization failed")
	}
	if _, err := database.InitRedis(); err != nil {
		logging.Log.Fatal("redis initialization failed")
	}

	r := logging.NewGinEngine()
	r.GET("/ping", handlers.Ping)
	r.POST("/internal/v1/reports", handlers.CreateReport)
	r.Run()
}
