package controllers

import (
	"go-api/config"
	"go-api/models"
	"go-api/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateGroup(c *gin.Context) {
	var group models.MstrGroup
	if err := c.ShouldBindJSON(&group); err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	userCompanyID, _ := c.Get("company_id")
	group.CompanyID = userCompanyID.(string)
	username, _ := c.Get("username")
	group.CreatedBy = username.(string)
	group.UpdatedBy = username.(string)

	// Simpan group
	if err := config.DB.Create(&group).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Group created", group)
}

// Update Group
func UpdateGroupByID(c *gin.Context) {
	DeviceID := c.Param("id")

	// Cek apakah group ada untuk company tersebut
	var group models.MstrGroup
	if err := config.DB.
		Where("id = ?", DeviceID).
		First(&group).Error; err != nil {
		utils.JSONError(c, http.StatusNotFound, "Group not found")
		return
	}

	// Bind data update dari body JSON
	var input struct {
		GroupName string `json:"group_name"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Update field
	group.GroupName = input.GroupName
	username, _ := c.Get("username")
	group.UpdatedBy = username.(string)

	if err := config.DB.Save(&group).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Group updated", group)
}

// Delete Group
func DeleteGroupByID(c *gin.Context) {
	GroupID := c.Param("id")

	deletedBy := c.GetString("username")

	// Set DeletedBy
	if err := config.DB.Model(&models.MstrGroup{}).
		Where("id = ?", GroupID).
		Update("deleted_by", deletedBy).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	if err := config.DB.Delete(&models.MstrGroup{}, GroupID).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Hapus group
	/*
		if err := config.DB.
			Where("id = ?", GroupID).
			Delete(&models.MstrGroup{}).Error; err != nil {
			utils.JSONError(c, http.StatusInternalServerError, err.Error())
			return
		}
	*/

	utils.JSONSuccess(c, "Group deleted", nil)
}

// Get Filtered Groups
func GetFilteredGroups(c *gin.Context) {
	createdBy := c.Query("created_by")
	updatedBy := c.Query("updated_by")
	GroupName := c.Query("group_name")
	//CompanyID := c.Query("company_id")

	var groups []models.MstrGroup
	query := config.DB.Model(&models.MstrGroup{})

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
	if GroupName != "" {
		query = query.Where("group_name ILIKE ?", "%"+GroupName+"%")
	}

	if err := query.Order("id DESC").Find(&groups).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Filtered groups", groups)
}
