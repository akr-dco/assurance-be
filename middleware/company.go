package middleware

import (
	"fmt"
	"go-api/config"
	"go-api/models"
	"go-api/utils"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CheckCompanyActive() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("[DEBUG] CheckCompanyActive START")
		// Ambil data dari JWT
		roleVal, _ := c.Get("role")
		userCompanyID, _ := c.Get("company_id")
		emailVal, _ := c.Get("email")

		role := roleVal.(string)
		companyID := userCompanyID.(string)
		email := emailVal.(string)

		//log.Printf("[DEBUG] Email: %v, Role: %s, CompanyID: %s", email, role, companyID)

		// Kalau super-admin → skip pengecekan
		if role == "super-admin" {
			c.Next()
			return
		}

		// === Cek company aktif ===
		var company models.MstrCompany
		if err := config.DB.Where("company_id = ?", companyID).First(&company).Error; err != nil {
			log.Printf("[ERROR] Company not found: %s", companyID)
			utils.JSONError(c, http.StatusNotFound, "Company not found")
			c.Abort()
			return
		}

		if !company.IsActive {
			log.Printf("[ERROR] Company is not active: %s", companyID)
			utils.JSONError(c, http.StatusForbidden, "Company is not active")
			c.Abort()
			return
		}

		// === Cek user aktif di company ===
		var user models.MstrUser
		if err := config.DB.Where("email = ? AND company_id = ?", email, companyID).First(&user).Error; err != nil {
			log.Printf("[ERROR] User not found in this company: %v / %s", email, companyID)
			utils.JSONError(c, http.StatusForbidden, "User not found")
			c.Abort()
			return
		}

		if !user.IsActive {
			log.Printf("[ERROR] User is not active in this company: %v / %s", email, companyID)
			utils.JSONError(c, http.StatusForbidden, "User is not active in this company")
			c.Abort()
			return
		}

		c.Set("company", company)
		c.Set("user", user)
		fmt.Println("[DEBUG] CheckCompanyActive END")
		c.Next()
	}
}
