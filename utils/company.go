package utils

import (
	"go-api/config"
	"go-api/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CheckCompanyActive(c *gin.Context, companyID string) (*models.MstrCompany, bool) {
	var company models.MstrCompany
	if err := config.DB.Where("company_id = ?", companyID).
		First(&company).Error; err != nil {
		JSONError(c, http.StatusNotFound, "Company not found")
		return nil, false
	}

	if !company.IsActive {
		JSONError(c, http.StatusForbidden, "Company is not active")
		return nil, false
	}

	return &company, true
}

/*
//Cara penggunaan
_, ok := utils.CheckCompanyActive(c, device.CompanyID)
	if !ok {
		return // sudah di-handle errornya di helper
	}
*/
