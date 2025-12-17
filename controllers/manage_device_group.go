package controllers

import (
	"go-api/config"
	"go-api/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ManageGroupDeviceBulk(c *gin.Context) {
	groupID := c.Param("groupId")

	var req struct {
		DeviceIDs []uint `json:"device_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	// Ambil group beserta relasi devices
	var group models.MstrGroup
	if err := config.DB.Preload("Devices").First(&group, groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Group not found"})
		return
	}

	// Ambil semua devices yang akan di-assign
	var devices []models.MstrDevice
	if len(req.DeviceIDs) > 0 {
		if err := config.DB.Where("id IN ?", req.DeviceIDs).Find(&devices).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
			return
		}
	}

	// Hapus semua device dari group
	if err := config.DB.Model(&group).Association("Devices").Clear(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to clear devices from group"})
		return
	}

	// Assign kembali device baru
	if len(devices) > 0 {
		if err := config.DB.Model(&group).Association("Devices").Append(&devices); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to assign devices to group"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Group devices updated successfully",
	})
}

// GET /groups/:id/devices
func GetGroupDevices(c *gin.Context) {
	groupID := c.Param("id")

	// Ambil company_id dari JWT context
	companyID, ok := c.Get("company_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Company ID not found in token"})
		return
	}

	var group models.MstrGroup
	if err := config.DB.
		Preload("Devices", "company_id = ?", companyID).
		Where("company_id = ?", companyID).
		First(&group, groupID).Error; err != nil {
		c.JSON(404, gin.H{"status": "error", "message": "Group not found"})
		return
	}

	var allDevices []models.MstrDevice
	if err := config.DB.Where("company_id = ? AND is_active = ?", companyID, true).Find(&allDevices).Error; err != nil {
		c.JSON(500, gin.H{"status": "error", "message": err.Error()})
		return
	}

	// Ambil hanya id device yang sudah ter-assign ke group
	assignedIDs := make([]uint, len(group.Devices))
	for i, d := range group.Devices {
		assignedIDs[i] = d.Id
	}

	c.JSON(200, gin.H{
		"status":              "success",
		"all_devices":         allDevices,
		"assigned_device_ids": assignedIDs,
	})
}
