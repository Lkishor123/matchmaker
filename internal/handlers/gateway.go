package handlers

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	stdproxy "net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"

	"matchmaker/internal/httputil"
	"matchmaker/internal/logging"
)

// workerPool limits concurrent proxy requests.
type workerPool struct {
	jobs chan func()
}

func newWorkerPool(n int) *workerPool {
	p := &workerPool{jobs: make(chan func())}
	for i := 0; i < n; i++ {
		go func() {
			for fn := range p.jobs {
				fn()
			}
		}()
	}
	return p
}

func (p *workerPool) Do(fn func()) {
	done := make(chan struct{})
	p.jobs <- func() { fn(); close(done) }
	<-done
}

// Gateway proxies requests to internal services.
type Gateway struct {
	auth  *stdproxy.ReverseProxy
	user  *stdproxy.ReverseProxy
	match *stdproxy.ReverseProxy
	chat  *stdproxy.ReverseProxy
	key   *rsa.PublicKey
	pool  *workerPool
}

// NewGateway constructs a Gateway using service URLs and RSA key.
func NewGateway(authURL, userURL, matchURL, chatURL, pemKey string, workers int) (*Gateway, error) {
	parse := func(u string) (*stdproxy.ReverseProxy, error) {
		url, err := url.Parse(u)
		if err != nil {
			return nil, err
		}
		p := stdproxy.NewSingleHostReverseProxy(url)
		return p, nil
	}
	auth, err := parse(authURL)
	if err != nil {
		return nil, err
	}
	user, err := parse(userURL)
	if err != nil {
		return nil, err
	}
	match, err := parse(matchURL)
	if err != nil {
		return nil, err
	}
	chat, err := parse(chatURL)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode([]byte(pemKey))
	if block == nil {
		return nil, fmt.Errorf("invalid RSA key")
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return &Gateway{auth, user, match, chat, &key.PublicKey, newWorkerPool(workers)}, nil
}

func (g *Gateway) proxy(p *stdproxy.ReverseProxy) gin.HandlerFunc {
	return func(c *gin.Context) {
		g.pool.Do(func() { p.ServeHTTP(c.Writer, c.Request) })
	}
}

// JWTMiddleware verifies tokens using the configured RSA key.
func (g *Gateway) JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			httputil.AbortJSONError(c, http.StatusUnauthorized, "missing bearer token")
			return
		}
		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return g.key, nil
		})
		if err != nil || !token.Valid {
			logging.Log.WithError(err).Warn("jwt verification failed")
			httputil.AbortJSONError(c, http.StatusUnauthorized, "invalid token")
			return
		}
		c.Next()
	}
}

func (g *Gateway) AuthHandler() gin.HandlerFunc  { return g.proxy(g.auth) }
func (g *Gateway) UserHandler() gin.HandlerFunc  { return g.proxy(g.user) }
func (g *Gateway) MatchHandler() gin.HandlerFunc { return g.proxy(g.match) }
func (g *Gateway) ChatHandler() gin.HandlerFunc  { return g.proxy(g.chat) }
