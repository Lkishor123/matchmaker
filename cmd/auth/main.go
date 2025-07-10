package main

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"matchmaker/internal/config"
	"matchmaker/internal/handlers"
	"matchmaker/internal/httputil"
	"matchmaker/internal/logging"
)

var (
	oauthConfig    *oauth2.Config
	jwtPrivateKey  *rsa.PrivateKey
	userServiceURL string
)

func main() {
	logging.Init()

	cfg, err := config.LoadAuth()
	if err != nil {
		logging.Log.WithError(err).Fatal("config error")
	}

	oauthConfig = &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	keyData := cfg.JWTPrivateKey
	if keyData != "" {
		block, _ := pem.Decode([]byte(keyData))
		if block != nil {
			if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
				jwtPrivateKey = key
			} else {
				logging.Log.WithError(err).Error("failed to parse RSA private key")
			}
		}
	}

	userServiceURL = cfg.UserServiceURL

	r := logging.NewGinEngine()
	r.GET("/ping", handlers.Ping)
	r.GET("/api/v1/auth/google/login", googleLoginHandler)
	r.GET("/api/v1/auth/google/callback", googleCallbackHandler)

	// Example usage of jwt-go to ensure dependency is referenced.
	_ = jwt.New(jwt.SigningMethodHS256)

	r.Run()
}

func googleLoginHandler(c *gin.Context) {
	if oauthConfig == nil {
		logging.Log.Error("oauth config not initialized")
		httputil.JSONError(c, http.StatusInternalServerError, "internal error")
		return
	}
	url := oauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	logging.Log.WithField("url", url).Info("redirecting to google oauth")
	c.Redirect(http.StatusFound, url)
}

func googleCallbackHandler(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		logging.Log.Warn("missing code in callback")
		httputil.JSONError(c, http.StatusBadRequest, "missing code")
		return
	}

	tok, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		logging.Log.WithError(err).Error("token exchange failed")
		httputil.JSONError(c, http.StatusBadRequest, "token exchange failed")
		return
	}

	client := oauthConfig.Client(context.Background(), tok)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		logging.Log.WithError(err).Error("failed to fetch user info")
		httputil.JSONError(c, http.StatusBadGateway, "failed to fetch user info")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logging.Log.WithField("status", resp.StatusCode).WithField("body", string(body)).Error("google userinfo returned non-200")
		httputil.JSONError(c, http.StatusBadGateway, "google userinfo failed")
		return
	}

	var gUser struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&gUser); err != nil {
		logging.Log.WithError(err).Error("failed to decode user info")
		httputil.JSONError(c, http.StatusBadGateway, "invalid user info")
		return
	}

	// Call User Service
	body, err := json.Marshal(map[string]string{
		"email": gUser.Email,
		"name":  gUser.Name,
	})
	if err != nil {
		logging.Log.WithError(err).Error("failed to marshal user request")
		httputil.JSONError(c, http.StatusInternalServerError, "internal error")
		return
	}

	usResp, err := http.Post(userServiceURL+"/internal/v1/users", "application/json", bytes.NewReader(body))
	if err != nil {
		logging.Log.WithError(err).Error("user service request failed")
		httputil.JSONError(c, http.StatusBadGateway, "user service unavailable")
		return
	}
	defer usResp.Body.Close()
	if usResp.StatusCode != http.StatusOK && usResp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(usResp.Body)
		logging.Log.WithFields(map[string]interface{}{"status": usResp.StatusCode, "body": string(b)}).Error("user service returned error")
		httputil.JSONError(c, http.StatusBadGateway, "user service error")
		return
	}
	var userResp struct {
		ID uint `json:"id"`
	}
	if err := json.NewDecoder(usResp.Body).Decode(&userResp); err != nil {
		logging.Log.WithError(err).Error("failed to decode user service response")
		httputil.JSONError(c, http.StatusBadGateway, "invalid user service response")
		return
	}

	if jwtPrivateKey == nil {
		logging.Log.Error("jwt private key not configured")
		httputil.JSONError(c, http.StatusInternalServerError, "internal error")
		return
	}

	claims := jwt.MapClaims{
		"user_id": userResp.ID,
		"email":   gUser.Email,
		"roles":   []string{"user"},
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(jwtPrivateKey)
	if err != nil {
		logging.Log.WithError(err).Error("failed to sign jwt")
		httputil.JSONError(c, http.StatusInternalServerError, "internal error")
		return
	}

	logging.Log.WithField("user_id", userResp.ID).Info("authentication successful")
	c.JSON(http.StatusOK, gin.H{"token": signed})
}
