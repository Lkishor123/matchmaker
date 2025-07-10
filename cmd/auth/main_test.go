package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"matchmaker/internal/logging"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestGoogleLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logging.Init()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/login", nil)

	// when oauthConfig is nil
	oauthConfig = nil
	googleLoginHandler(c)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}

	// success case
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()
	oauthConfig = &oauth2.Config{ClientID: "id", ClientSecret: "sec", Endpoint: oauth2.Endpoint{AuthURL: srv.URL + "/auth"}, RedirectURL: "http://example.com"}
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/login", nil)
	googleLoginHandler(c)
	if w.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", w.Code)
	}
	loc := w.Header().Get("Location")
	if !strings.HasPrefix(loc, srv.URL+"/auth") {
		t.Fatalf("unexpected redirect %s", loc)
	}
}

func TestGoogleCallback(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logging.Init()

	// missing code
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/callback", nil)
	googleCallbackHandler(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	// success path
	oauthSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"access_token":"tok","token_type":"Bearer"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer oauthSrv.Close()

	oldTransport := http.DefaultTransport
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if strings.Contains(req.URL.Host, "googleapis.com") {
			rec := httptest.NewRecorder()
			rec.Header().Set("Content-Type", "application/json")
			rec.WriteString(`{"email":"a@b.com","name":"Alice"}`)
			return rec.Result(), nil
		}
		return oldTransport.RoundTrip(req)
	})
	defer func() { http.DefaultTransport = oldTransport }()

	userSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":1}`))
	}))
	defer userSrv.Close()

	oauthConfig = &oauth2.Config{
		ClientID:     "id",
		ClientSecret: "sec",
		RedirectURL:  "http://example.com",
		Endpoint: oauth2.Endpoint{
			AuthURL:  oauthSrv.URL + "/auth",
			TokenURL: oauthSrv.URL + "/token",
		},
	}
	userServiceURL = userSrv.URL
	key, _ := rsa.GenerateKey(rand.Reader, 512)
	jwtPrivateKey = key

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/callback?code=abc", nil)
	googleCallbackHandler(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp struct{ Token string }
	if err := json.NewDecoder(bytes.NewReader(w.Body.Bytes())).Decode(&resp); err != nil || resp.Token == "" {
		t.Fatalf("expected token, err=%v", err)
	}
}
