package database

import (
	"context"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"matchmaker/internal/logging"
)

var Mongo *mongo.Database

// InitMongo connects to MongoDB using MONGO_URL.
func InitMongo() (*mongo.Database, error) {
	uri := os.Getenv("MONGO_URL")
	if uri == "" {
		logging.Log.Warn("MONGO_URL not set")
	}
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		logging.Log.WithError(err).Error("failed to connect to mongodb")
		return nil, err
	}
	Mongo = client.Database("astrology")
	return Mongo, nil
}
