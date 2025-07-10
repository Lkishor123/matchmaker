package handlers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"matchmaker/internal/database"
	"matchmaker/internal/httputil"
	"matchmaker/internal/logging"
)

type BirthDetails struct {
	DOB string  `json:"dob"`
	TOB string  `json:"tob"`
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// CreateReport handles POST /internal/v1/reports to fetch astrology reports.
func CreateReport(c *gin.Context) {
	var bd BirthDetails
	if err := c.ShouldBindJSON(&bd); err != nil {
		logging.Log.WithError(err).Warn("invalid report payload")
		httputil.JSONError(c, http.StatusBadRequest, "invalid request")
		return
	}

	key := reportKey(bd)
	ctx := context.Background()

	// L1 Cache (Redis)
	if val, err := database.Redis.Get(ctx, key).Result(); err == nil {
		c.Data(http.StatusOK, "application/json", []byte(val))
		return
	} else if err != redis.Nil {
		logging.Log.WithError(err).WithField("key", key).Error("redis get failed")
	}

	// L2 Cache (MongoDB)
	var doc struct {
		Report json.RawMessage `bson:"report"`
	}
	if err := database.Mongo.Collection("reports").FindOne(ctx, bson.M{"_id": key}).Decode(&doc); err == nil {
		c.Data(http.StatusOK, "application/json", doc.Report)
		go writeCaches(key, doc.Report)
		return
	} else if err != mongo.ErrNoDocuments {
		logging.Log.WithError(err).WithField("key", key).Error("mongo find failed")
	}

	// Cache miss - fetch from external engine
	data, err := fetchFromEngine(ctx, bd)
	if err != nil {
		logging.Log.WithError(err).Error("engine request failed")
		httputil.JSONError(c, http.StatusBadGateway, "engine error")
		return
	}

	c.Data(http.StatusOK, "application/json", data)
	go writeCaches(key, data)
}

func reportKey(b BirthDetails) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%.8f:%.8f", b.DOB, b.TOB, b.Lat, b.Lon)))
	return hex.EncodeToString(sum[:])
}

func fetchFromEngine(ctx context.Context, b BirthDetails) ([]byte, error) {
	url := os.Getenv("ASTROLOGY_ENGINE_URL")
	apiKey := os.Getenv("ASTROLOGY_ENGINE_API_KEY")
	if url == "" {
		return nil, fmt.Errorf("ASTROLOGY_ENGINE_URL not set")
	}
	body, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		bts, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("engine status %d: %s", resp.StatusCode, string(bts))
	}
	return io.ReadAll(resp.Body)
}

func writeCaches(key string, data []byte) {
	ctx := context.Background()
	if err := database.Redis.Set(ctx, key, data, time.Hour).Err(); err != nil {
		logging.Log.WithError(err).WithField("key", key).Error("redis set failed")
	}
	_, err := database.Mongo.Collection("reports").UpdateOne(ctx,
		bson.M{"_id": key},
		bson.M{"$set": bson.M{"report": data, "createdAt": time.Now()}},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		logging.Log.WithError(err).WithField("key", key).Error("mongo upsert failed")
	}
}
