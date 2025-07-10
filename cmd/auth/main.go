package main

import (
	"github.com/golang-jwt/jwt/v4"

	"matchmaker/internal/handlers"
	"matchmaker/internal/logging"
)

func main() {
	logging.Init()
	r := logging.NewGinEngine()
	r.GET("/ping", handlers.Ping)

	// Example usage of jwt-go to ensure dependency is referenced.
	_ = jwt.New(jwt.SigningMethodHS256)

	r.Run()
}
