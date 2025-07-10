package logging

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

// Init configures the global logger.
func Init() {
	Log = logrus.New()
	Log.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
}

// NewGinEngine returns a Gin engine that logs requests and panics using Log.
func NewGinEngine() *gin.Engine {
	if Log == nil {
		Init()
	}
	engine := gin.New()
	engine.Use(gin.RecoveryWithWriter(Log.WriterLevel(logrus.ErrorLevel)))
	engine.Use(requestLogger())
	return engine
}

func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		entry := Log.WithFields(logrus.Fields{
			"status":   c.Writer.Status(),
			"method":   c.Request.Method,
			"path":     c.Request.URL.Path,
			"clientIP": c.ClientIP(),
			"latency":  time.Since(start),
		})
		if len(c.Errors) > 0 {
			entry.Error(c.Errors.String())
		} else {
			entry.Info("request handled")
		}
	}
}
