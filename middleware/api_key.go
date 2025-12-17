package middleware

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("[DEBUG] APIKeyAuth START")
		key := c.GetHeader("X-API-KEY")
		if key != os.Getenv("API_KEY") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		fmt.Println("[DEBUG] APIKeyAuth END")
		c.Next()
	}
}
