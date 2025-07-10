package handlers

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func genKey(t *testing.T) *rsa.PrivateKey {
	key, err := rsa.GenerateKey(rand.Reader, 512)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func TestGatewayUserProxy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	key := genKey(t)
	pemKey := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})

	hit := false
	userSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit = true
		if r.Header.Get("Authorization") == "" {
			t.Error("auth header missing")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer userSrv.Close()

	gw, err := NewGateway("http://x", userSrv.URL, "http://y", "http://z", string(pemKey), 1)
	if err != nil {
		t.Fatal(err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{"user_id": 1})
	signed, _ := token.SignedString(key)

	r := gin.New()
	r.GET("/users/me", gw.JWTMiddleware(), gw.UserHandler())
	srv := httptest.NewServer(r)
	defer srv.Close()

	req, _ := http.NewRequest("GET", srv.URL+"/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 got %d", resp.StatusCode)
	}
	if !hit {
		t.Fatal("backend not hit")
	}

	// invalid token
	req, _ = http.NewRequest("GET", srv.URL+"/users/me", nil)
	req.Header.Set("Authorization", "Bearer invalid")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 got %d", resp.StatusCode)
	}
}
