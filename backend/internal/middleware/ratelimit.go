package middleware

import (
	"fmt"
	"net/http"
	"time"

	"context"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// rateLimitScript atomically increments the counter and sets expiry if it's a new key.
// Returns the current count after increment.
var rateLimitScript = redis.NewScript(`
local count = redis.call("INCR", KEYS[1])
if count == 1 then
    redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
return count
`)

// RateLimit implements a fixed window rate limiter using Redis with atomic Lua script
func RateLimit(rdb *redis.Client, maxRequests int, window time.Duration) gin.HandlerFunc {
	windowMs := int(window.Milliseconds())

	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == "" {
			userID = c.ClientIP()
		}

		key := fmt.Sprintf("ratelimit:%s:%s", userID, c.FullPath())
		ctx := context.Background()

		count, err := rateLimitScript.Run(ctx, rdb, []string{key}, windowMs).Int64()
		if err != nil {
			// If Redis is unavailable, allow the request
			c.Next()
			return
		}

		if count > int64(maxRequests) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "too many requests",
			})
			return
		}

		c.Next()
	}
}
