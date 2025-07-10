package main

import (
	"context"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"matchmaker/internal/handlers"
	"matchmaker/internal/logging"
)

func main() {
	logging.Init()
	r := logging.NewGinEngine()
	r.GET("/ping", handlers.Ping)

	// Placeholder MongoDB and Redis initialization to reference libraries.
	_, _ = mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	_ = redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	r.Run()
}
