package middleware

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing Bearer token"})
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		req, _ := http.NewRequest("GET", "http://localhost:8082/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil || res.StatusCode != 200 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: invalid token"})
			return
		}

		var data struct {
			Username string `json:"username"`
			Role     string `json:"role"`
		}
		json.NewDecoder(res.Body).Decode(&data)

		if data.Role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Access denied: admin only"})
			return
		}

		c.Next()
	}
}
