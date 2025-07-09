package main

import (
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"matchmaker/internal/handlers"
	"matchmaker/internal/models"
)

func main() {
	r := gin.Default()
	r.GET("/ping", handlers.Ping)

	// Placeholder GORM initialization to reference the library.
	dsn := "host=localhost user=gorm dbname=gorm password=gorm sslmode=disable"
	_, _ = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	_ = models.User{}

	r.Run()
}
