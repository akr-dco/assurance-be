package controllers

import (
	"net/http"

	"go-api/utils"

	"github.com/gin-gonic/gin"
)

// response struct
type GuestTokenResponse struct {
	Token string `json:"token"`
}

// handler
func GetSupersetGuestToken(c *gin.Context) {
	role := c.GetString("role")
	companyID := c.GetString("company_id")

	if role != "super-admin" && companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id not found in token"})
		return
	}

	// panggil helper untuk buat guest token
	token, err := utils.CreateGuestToken(c, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, GuestTokenResponse{Token: token})
}
