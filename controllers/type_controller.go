package controllers

import (
	"go-api/config"
	"go-api/models"
	"go-api/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateType(c *gin.Context) {
	var type_trigger models.MstrTypeTrigger
	if err := c.ShouldBindJSON(&type_trigger); err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Isi metadata
	userCompanyID := c.GetString("company_id")
	username := c.GetString("username")

	type_trigger.CompanyID = userCompanyID
	type_trigger.CreatedBy = username
	type_trigger.UpdatedBy = username

	if err := config.DB.Create(&type_trigger).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Type created", type_trigger)
}

func UpdateTypeByID(c *gin.Context) {
	typeID := c.Param("id")

	var type_trigger models.MstrTypeTrigger
	if err := config.DB.
		Where("id = ?", typeID).
		First(&type_trigger).Error; err != nil {
		utils.JSONError(c, http.StatusNotFound, "Type not found")
		return
	}

	var input struct {
		TypeName    string `json:"type_name"`
		Description string `json:"description"`
		IsActive    bool   `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Update field
	type_trigger.TypeName = input.TypeName
	type_trigger.Description = input.Description
	type_trigger.IsActive = input.IsActive
	userCompanyID := c.GetString("company_id")
	type_trigger.CompanyID = userCompanyID
	username := c.GetString("username")
	type_trigger.UpdatedBy = username

	if err := config.DB.Save(&type_trigger).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Type updated", type_trigger)
}

func DeleteTypeByID(c *gin.Context) {
	typeID := c.Param("id")
	deletedBy := c.GetString("username")

	// Set DeletedBy
	if err := config.DB.Model(&models.MstrTypeTrigger{}).
		Where("id = ?", typeID).
		Update("deleted_by", deletedBy).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	if err := config.DB.Delete(&models.MstrTypeTrigger{}, typeID).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.JSONSuccess(c, "Type deleted", nil)
}

func GetFilteredTypes(c *gin.Context) {
	createdBy := c.Query("created_by")
	updatedBy := c.Query("updated_by")
	typeName := c.Query("type_name")
	typeID := c.Query("id")

	var types []models.MstrTypeTrigger
	query := config.DB.Model(&models.MstrTypeTrigger{})

	role := c.GetString("role")
	userCompanyID := c.GetString("company_id")
	if role != "super-admin" {
		query = query.Where("company_id = ?", userCompanyID)
	}

	if createdBy != "" {
		query = query.Where("created_by = ?", createdBy)
	}
	if updatedBy != "" {
		query = query.Where("updated_by = ?", updatedBy)
	}
	if typeName != "" {
		query = query.Where("type_name ILIKE ?", "%"+typeName+"%")
	}
	if typeID != "" {
		query = query.Where("id = ?", typeID)
	}

	if err := query.Order("id DESC").Find(&types).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Filtered types", types)
}
