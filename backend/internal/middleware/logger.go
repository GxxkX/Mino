package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Logger(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		entry := log.WithFields(logrus.Fields{
			"status":  status,
			"method":  c.Request.Method,
			"path":    path,
			"query":   query,
			"ip":      c.ClientIP(),
			"latency": latency,
		})

		if status >= 500 {
			entry.Error("server error")
		} else if status >= 400 {
			entry.Warn("client error")
		} else {
			entry.Info("request")
		}
	}
}
