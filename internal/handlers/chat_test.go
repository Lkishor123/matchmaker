package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"

	"matchmaker/internal/database"
	"matchmaker/internal/logging"
)

func TestChat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logging.Init()
	mr, _ := miniredis.Run()
	defer mr.Close()
	database.Redis = redis.NewClient(&redis.Options{Addr: mr.Addr()})

	llm := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hi"))
	}))
	defer llm.Close()
	os.Setenv("LLM_API_URL", llm.URL)
	os.Setenv("LLM_API_KEY", "key")

	r := gin.New()
	r.GET("/chat", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		Chat(c)
	})
	srv := httptest.NewServer(r)
	defer srv.Close()

	url := "ws" + srv.URL[len("http"):] + "/chat"
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ws.Close()

	if err := ws.WriteMessage(websocket.TextMessage, []byte("hello")); err != nil {
		t.Fatal(err)
	}
	ws.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err := ws.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	if string(msg) != "hi" {
		t.Fatalf("unexpected message %s", msg)
	}

	time.Sleep(20 * time.Millisecond)
	if v, err := mr.Get("chat_context:1"); err != nil || v == "" {
		t.Fatalf("context not stored")
	}

	// failure path
	llmFail := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer llmFail.Close()
	os.Setenv("LLM_API_URL", llmFail.URL)

	ws2, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ws2.Close()
	if err := ws2.WriteMessage(websocket.TextMessage, []byte("hi")); err != nil {
		t.Fatal(err)
	}
	ws2.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, _, err = ws2.ReadMessage()
	if err == nil {
		t.Fatalf("expected error")
	}
}
