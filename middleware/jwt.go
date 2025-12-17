package middleware

import (
	"fmt"
	"go-api/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("[DEBUG] AuthMiddleware START")
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid Authorization header"})
			return
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return []byte(utils.GetJwtSecret()), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		// Simpan ke context
		if username, exists := claims["username"]; exists {
			c.Set("username", username)
		}
		if role, exists := claims["role"]; exists {
			c.Set("role", role)
		}
		if companyID, exists := claims["company_id"]; exists {
			c.Set("company_id", companyID)
		}
		if email, exists := claims["email"]; exists {
			c.Set("email", email)
		}
		fmt.Println("[DEBUG] AuthMiddleware END")
		c.Next()
	}
}
