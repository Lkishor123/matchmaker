package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"

	"matchmaker/internal/database"
	"matchmaker/internal/logging"
)

func TestCreateReport(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logging.Init()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()
	database.Redis = redis.NewClient(&redis.Options{Addr: mr.Addr()})

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("success", func(mt *mtest.T) {
		database.Mongo = mt.DB
		mt.AddMockResponses(
			mtest.CreateCursorResponse(0, "astrology.reports", mtest.FirstBatch),
			mtest.CreateSuccessResponse(),
		)

		engine := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"ok":true}`))
		}))
		defer engine.Close()
		os.Setenv("ASTROLOGY_ENGINE_URL", engine.URL)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := `{"dob":"2000-02-02","tob":"12:00:00","lat":5,"lon":3}`
		c.Request = httptest.NewRequest("POST", "/", bytes.NewBufferString(body))
		c.Request.Header.Set("Content-Type", "application/json")
		CreateReport(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 got %d", w.Code)
		}
		time.Sleep(20 * time.Millisecond)
	})

	mt.Run("engine failure", func(mt *mtest.T) {
		database.Mongo = mt.DB
		mt.AddMockResponses(
			mtest.CreateCursorResponse(0, "astrology.reports", mtest.FirstBatch),
		)
		engine := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
		}))
		defer engine.Close()
		os.Setenv("ASTROLOGY_ENGINE_URL", engine.URL)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := `{"dob":"2000-01-01","tob":"12:00:00","lat":1,"lon":2}`
		c.Request = httptest.NewRequest("POST", "/", bytes.NewBufferString(body))
		c.Request.Header.Set("Content-Type", "application/json")
		CreateReport(c)
		if w.Code != http.StatusBadGateway {
			t.Fatalf("expected 502 got %d", w.Code)
		}
	})
}
