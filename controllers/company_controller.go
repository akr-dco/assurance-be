package controllers

import (
	"go-api/config"
	"go-api/models"
	"go-api/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateCompany(c *gin.Context) {
	var company models.MstrCompany

	// Ambil data form (karena ada file upload)
	company.CompanyName = c.PostForm("company_name")
	company.CompanyID = c.PostForm("company_id")
	company.E2Endpoint = c.PostForm("e2_endpoint")
	company.E2Region = c.PostForm("e2_region")
	company.E2BucketName = c.PostForm("e2_bucket_name")
	company.E2AccessKey = c.PostForm("e2_access_key")
	company.E2SecretKey = c.PostForm("e2_secret_key")

	// === Upload logo ke E2 ===
	fileHeader, err := c.FormFile("image_url")
	if err == nil { // logo opsional
		file, err := fileHeader.Open()
		if err != nil {
			utils.JSONError(c, http.StatusInternalServerError, "Cannot open logo file: "+err.Error())
			return
		}
		defer file.Close()

		override := &E2Config{
			Endpoint:   company.E2Endpoint,
			Region:     company.E2Region,
			BucketName: company.E2BucketName,
			AccessKey:  company.E2AccessKey,
			SecretKey:  company.E2SecretKey,
		}

		fileKey := GenerateE2ObjectKey(c, "Master-Company", fileHeader.Filename)
		objectKey, err := UploadFileToE2(c, file, fileKey, fileHeader.Header.Get("Content-Type"), "Assurance/"+company.CompanyID, override)
		if err != nil {
			utils.JSONError(c, http.StatusInternalServerError, "Failed upload: "+err.Error())
			return
		}

		/*
			fileName := fmt.Sprintf("%d-%s", time.Now().Unix(), fileHeader.Filename)
			objectKey, err := UploadFileToE2(c, file, fileName, fileHeader.Header.Get("Content-Type"), "Assurance")
			if err != nil {
				utils.JSONError(c, http.StatusInternalServerError, "Failed upload: "+err.Error())
				return
			}
		*/

		// Simpan objectKey di DB, bukan URL
		company.ImageUrl = objectKey
	}

	// Ambil username dari context
	username := c.GetString("username")
	company.CreatedBy = username
	company.UpdatedBy = username

	if err := config.DB.WithContext(c.Request.Context()).Create(&company).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Failed to create company")
		return
	}

	utils.JSONSuccess(c, "Company created successfully", company)
}

func UpdateCompany(c *gin.Context) {
	companyID := c.Param("id")

	var company models.MstrCompany
	if err := config.DB.WithContext(c.Request.Context()).First(&company, companyID).Error; err != nil {
		utils.JSONError(c, http.StatusNotFound, "Company not found")
		return
	}

	// Ambil data dari form (karena bisa ada file upload juga)
	companyName := c.PostForm("company_name")
	companyCode := c.PostForm("company_id")
	isActive := c.PostForm("is_active")
	E2Endpoint := c.PostForm("e2_endpoint")
	E2Region := c.PostForm("e2_region")
	E2BucketName := c.PostForm("e2_bucket_name")
	E2AccessKey := c.PostForm("e2_access_key")
	E2SecretKey := c.PostForm("e2_secret_key")

	if companyName != "" {
		company.CompanyName = companyName
	}

	if companyCode != "" {
		company.CompanyID = companyCode
	}

	if isActive != "" {
		// parse string ke bool
		company.IsActive = (isActive == "true" || isActive == "1")
	}

	if E2Endpoint != "" {
		company.E2Endpoint = E2Endpoint
	}

	if E2Region != "" {
		company.E2Region = E2Region
	}

	if E2BucketName != "" {
		company.E2BucketName = E2BucketName
	}

	if E2AccessKey != "" {
		company.E2AccessKey = E2AccessKey
	}

	if E2SecretKey != "" {
		company.E2SecretKey = E2SecretKey
	}

	// === Cek apakah ada file upload baru untuk logo ===
	fileHeader, err := c.FormFile("image_url")
	if err == nil {
		file, err := fileHeader.Open()
		if err != nil {
			utils.JSONError(c, http.StatusInternalServerError, "Cannot open logo file: "+err.Error())
			return
		}
		defer file.Close()

		override := &E2Config{
			Endpoint:   company.E2Endpoint,
			Region:     company.E2Region,
			BucketName: company.E2BucketName,
			AccessKey:  company.E2AccessKey,
			SecretKey:  company.E2SecretKey,
		}

		fileKey := GenerateE2ObjectKey(c, "Master-Company", fileHeader.Filename)
		objectKey, err := UploadFileToE2(c, file, fileKey, fileHeader.Header.Get("Content-Type"), "Assurance/"+companyCode, override)
		if err != nil {
			utils.JSONError(c, http.StatusInternalServerError, "Failed upload: "+err.Error())
			return
		}

		/*
			fileName := fmt.Sprintf("%d-%s", time.Now().Unix(), fileHeader.Filename)
			objectKey, err := UploadFileToE2(c, file, fileName, fileHeader.Header.Get("Content-Type"), "Assurance")
			if err != nil {
				utils.JSONError(c, http.StatusInternalServerError, "Failed upload: "+err.Error())
				return
			}
		*/

		// Simpan objectKey di DB, bukan URL
		company.ImageUrl = objectKey
	}

	// Update audit fields
	username := c.GetString("username")
	company.UpdatedBy = username

	if err := config.DB.WithContext(c.Request.Context()).Save(&company).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Failed to update company")
		return
	}

	utils.JSONSuccess(c, "Company updated successfully", company)
}

func DeleteCompany(c *gin.Context) {
	CompanyID := c.Param("id")

	deletedBy := c.GetString("username")

	// Set DeletedBy
	if err := config.DB.Model(&models.MstrCompany{}).
		Where("id = ?", CompanyID).
		Update("deleted_by", deletedBy).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Cek relasi di tabel lain, misalnya mstr_user
	/*
		var count int64
		config.DB.Model(&models.MstrCompany{}).Where("id = ?", CompanyID).Count(&count)
		if count > 0 {
			utils.JSONError(c, http.StatusBadRequest, "Company masih digunakan di tabel lain")
			return
		}
	*/

	if err := config.DB.Delete(&models.MstrCompany{}, CompanyID).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Company deleted", nil)
}

func GetFilteredCompanies(c *gin.Context) {
	createdBy := c.Query("created_by")
	updatedBy := c.Query("updated_by")
	companyName := c.Query("company_name")
	//companyID := c.Query("company_id")

	var companies []models.MstrCompany
	query := config.DB.Model(&models.MstrCompany{})

	role, _ := c.Get("role")
	userCompanyID, _ := c.Get("company_id")
	if role != "super-admin" {
		query = query.Where("company_id = ?", userCompanyID)
	}

	if createdBy != "" {
		query = query.Where("created_by = ?", createdBy)
	}
	if updatedBy != "" {
		query = query.Where("updated_by = ?", updatedBy)
	}
	if companyName != "" {
		query = query.Where("company_name ILIKE ?", "%"+companyName+"%")
	}

	if err := query.Order("id DESC").Find(&companies).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Filtered companies", companies)
}
