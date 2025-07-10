package main

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"matchmaker/internal/handlers"
	"matchmaker/internal/logging"
	"matchmaker/internal/models"
)

func main() {
	logging.Init()
	r := logging.NewGinEngine()
	r.GET("/ping", handlers.Ping)

	// Placeholder GORM initialization to reference the library.
	dsn := "host=localhost user=gorm dbname=gorm password=gorm sslmode=disable"
	_, _ = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	_ = models.User{}

	r.Run()
}
