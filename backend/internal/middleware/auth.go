package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	jwtpkg "github.com/mino/backend/internal/pkg/jwt"
)

const UserIDKey = "userID"
const UsernameKey = "username"
const RoleKey = "role"

func Auth(jwtMgr *jwtpkg.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string

		// 1. Try Authorization header (Bearer <token>)
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
				token = parts[1]
			}
		}

		// 2. Fallback to ?token= query param (for <audio>/<img> elements
		//    that cannot send custom headers).
		if token == "" {
			token = c.Query("token")
		}

		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "missing authorization"})
			return
		}

		claims, err := jwtMgr.Validate(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "invalid or expired token"})
			return
		}

		c.Set(UserIDKey, claims.UserID)
		c.Set(UsernameKey, claims.Username)
		c.Set(RoleKey, claims.Role)
		c.Next()
	}
}

// GetUserID extracts the authenticated user ID from context
func GetUserID(c *gin.Context) string {
	v, _ := c.Get(UserIDKey)
	s, _ := v.(string)
	return s
}
