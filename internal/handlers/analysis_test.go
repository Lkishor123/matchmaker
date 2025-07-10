package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"matchmaker/internal/logging"
)

func TestCreateAnalysis(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logging.Init()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"report":true}`))
	}))
	defer srv.Close()
	os.Setenv("REPORT_SERVICE_URL", srv.URL)

	body := `{"personA":{"dob":"2000","tob":"12:00:00","lat":1,"lon":2},"personB":{"dob":"2001","tob":"12:00:00","lat":1,"lon":2}}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	CreateAnalysis(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}
	var resp map[string]int
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["score"] == 0 {
		t.Fatalf("missing score")
	}

	// failure when report service returns error
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer bad.Close()
	os.Setenv("REPORT_SERVICE_URL", bad.URL)
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	CreateAnalysis(c)
	if w.Code != http.StatusBadGateway {
		t.Fatalf("expected 502 got %d", w.Code)
	}
}
