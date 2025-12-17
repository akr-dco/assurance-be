package controllers

import (
	"go-api/config"
	"go-api/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ManageGroupInspectionBulk(c *gin.Context) {
	groupID := c.Param("groupId")

	var req struct {
		InspectionIDs []uint `json:"inspection_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	// Ambil group beserta relasi inspections
	var group models.MstrGroup
	if err := config.DB.Preload("Inspections").First(&group, groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Group not found"})
		return
	}

	// Ambil semua inspections yang akan di-assign
	var inspections []models.MstrInspection
	if len(req.InspectionIDs) > 0 {
		if err := config.DB.Where("id IN ?", req.InspectionIDs).Find(&inspections).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
			return
		}
	}

	// Hapus semua inspection dari group
	if err := config.DB.Model(&group).Association("Inspections").Clear(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to clear inspections from group"})
		return
	}

	// Assign kembali inspection baru
	if len(inspections) > 0 {
		if err := config.DB.Model(&group).Association("Inspections").Append(&inspections); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to assign inspections to group"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Group inspections updated successfully",
	})
}

func GetGroupInspections(c *gin.Context) {
	groupID := c.Param("id")

	// Ambil company_id dari JWT context
	companyID, ok := c.Get("company_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Company ID not found in token"})
		return
	}

	// Ambil data group + preload inspections tapi filter company_id
	var group models.MstrGroup
	if err := config.DB.
		Preload("Inspections", "company_id = ?", companyID).
		Where("company_id = ?", companyID).
		First(&group, groupID).Error; err != nil {
		c.JSON(404, gin.H{"status": "error", "message": "Group not found"})
		return
	}

	// Ambil semua inspections dengan filter company_id
	var allInspections []models.MstrInspection
	if err := config.DB.Where("company_id = ?", companyID).Find(&allInspections).Error; err != nil {
		c.JSON(500, gin.H{"status": "error", "message": err.Error()})
		return
	}

	// Ambil hanya id inspection yang sudah ter-assign ke group
	assignedIDs := make([]uint, len(group.Inspections))
	for i, insp := range group.Inspections {
		assignedIDs[i] = insp.Id
	}

	c.JSON(200, gin.H{
		"status":                  "success",
		"all_inspections":         allInspections,
		"assigned_inspection_ids": assignedIDs,
	})
}
