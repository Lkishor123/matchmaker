package main

import (
	"matchmaker/internal/database"
	"matchmaker/internal/handlers"
	"matchmaker/internal/logging"
	"matchmaker/internal/models"
)

func main() {
	logging.Init()
	if _, err := database.Init(); err != nil {
		logging.Log.Fatal("database initialization failed")
	}
	if err := database.DB.AutoMigrate(&models.User{}, &models.BirthDetail{}); err != nil {
		logging.Log.WithError(err).Fatal("auto-migrate failed")
	}

	r := logging.NewGinEngine()
	r.GET("/ping", handlers.Ping)

	r.POST("/internal/v1/users", handlers.CreateUser)

	api := r.Group("/api/v1")
	api.Use(handlers.RequireUserID())
	api.GET("/users/me", handlers.GetMe)
	api.PUT("/users/me", handlers.UpdateMe)

	r.Run()
}
