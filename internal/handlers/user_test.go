package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"matchmaker/internal/logging"

	"matchmaker/internal/database"
	"matchmaker/internal/models"
)

func setupTestDB(t *testing.T) {
	logging.Init()
	dsn := fmt.Sprintf("file:%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.BirthDetail{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	database.DB = db
}

func TestCreateUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	// invalid
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/", bytes.NewBufferString(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")
	CreateUser(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d", w.Code)
	}

	// success
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	body := `{"email":"a@b.com","name":"Alice"}`
	c.Request = httptest.NewRequest("POST", "/", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	CreateUser(c)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 got %d", w.Code)
	}
	var resp map[string]uint
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["id"] == 0 {
		t.Fatal("missing id")
	}
}

func TestGetMe(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)
	user := models.User{Email: "u@x.com"}
	database.DB.Create(&user)

	// success
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Set("user_id", user.ID)
	GetMe(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}

	// not found
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Set("user_id", uint(999))
	GetMe(c)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 got %d", w.Code)
	}
}

func TestUpdateMe(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)
	user := models.User{Email: "u@x.com"}
	database.DB.Create(&user)

	// invalid body
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("PUT", "/", bytes.NewBufferString(`x`))
	c.Set("user_id", user.ID)
	UpdateMe(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d", w.Code)
	}

	// success
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	body := `{"location":"LA"}`
	c.Request = httptest.NewRequest("PUT", "/", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", user.ID)
	UpdateMe(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}
	var updated models.User
	database.DB.First(&updated, user.ID)
	if updated.Location != "LA" {
		t.Fatalf("location not updated: %s", updated.Location)
	}
}
