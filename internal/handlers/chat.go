package handlers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"

	"matchmaker/internal/database"
	"matchmaker/internal/logging"
)

var wsUpgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

// Chat handles GET /api/v1/chat and streams LLM responses over WebSocket.
func Chat(c *gin.Context) {
	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logging.Log.WithError(err).Warn("websocket upgrade failed")
		return
	}
	uid := c.GetUint("user_id")
	logging.Log.WithField("user_id", uid).Info("websocket connected")
	defer func() {
		conn.Close()
		logging.Log.WithField("user_id", uid).Info("websocket disconnected")
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			logging.Log.WithError(err).WithField("user_id", uid).Info("read loop ended")
			return
		}
		if err := handleChatMessage(c.Request.Context(), uid, msg, conn); err != nil {
			logging.Log.WithError(err).WithField("user_id", uid).Error("message handling failed")
			return
		}
	}
}

func handleChatMessage(ctx context.Context, uid uint, msg []byte, conn *websocket.Conn) error {
	key := fmt.Sprintf("chat_context:%d", uid)
	prev, err := database.Redis.Get(ctx, key).Result()
	if err == redis.Nil {
		prev = ""
	} else if err != nil {
		logging.Log.WithError(err).Error("redis get failed")
	}

	prompt := prev + "User: " + string(msg) + "\nAI:"
	body := bytes.NewBuffer(nil)
	_ = json.NewEncoder(body).Encode(map[string]interface{}{"prompt": prompt, "stream": true})

	llmURL := os.Getenv("LLM_API_URL")
	if llmURL == "" {
		llmURL = "https://example.com/api/chat"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, llmURL, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey := os.Getenv("LLM_API_KEY"); apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("llm status %d: %s", resp.StatusCode, string(b))
	}

	reader := bufio.NewReader(resp.Body)
	var respBuf bytes.Buffer
	for {
		chunk := make([]byte, 1024)
		n, err := reader.Read(chunk)
		if n > 0 {
			data := chunk[:n]
			if werr := conn.WriteMessage(websocket.TextMessage, data); werr != nil {
				return werr
			}
			respBuf.Write(data)
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}

	newCtx := prev + "User: " + string(msg) + "\nAI: " + respBuf.String() + "\n"
	if err := database.Redis.Set(ctx, key, newCtx, 10*time.Minute).Err(); err != nil {
		logging.Log.WithError(err).Error("redis set failed")
	}
	return nil
}
