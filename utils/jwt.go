package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte(getEnv("JWT_SECRET", "cxLIKW0ASke084UUdrLmW9JIrSp7sjy0DwGorBAF"))

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

func GenerateAccessToken(userID uint, username string, email string, role string, company_id string, ttl time.Duration) (string, int64, error) {
	expiresAt := time.Now().Add(ttl).Unix()
	claims := jwt.MapClaims{
		"sub":        userID,
		"username":   username,
		"email":      email,
		"role":       role,
		"company_id": company_id,
		"exp":        expiresAt,
		"iat":        time.Now().Unix(),
		"scope":      "access",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(jwtSecret)
	return signed, expiresAt, err
}

func GenerateRefreshToken(userID uint, ttl time.Duration) (string, error) {
	expiresAt := time.Now().Add(ttl).Unix()
	claims := jwt.MapClaims{
		"sub":   userID,
		"exp":   expiresAt,
		"iat":   time.Now().Unix(),
		"scope": "refresh",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ParseToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, err
	}
	return claims, nil
}

func GetJwtSecret() string {
	return string(jwtSecret)
}
