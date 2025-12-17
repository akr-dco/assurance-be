package controllers

import (
	"go-api/config"
	"go-api/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// POST /groups/:groupId/chainings (bulk assign chaining ke group)
func ManageGroupChainingBulk(c *gin.Context) {
	groupID := c.Param("groupId")

	var req struct {
		ChainingIDs []uint `json:"chaining_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	// Ambil group beserta relasi chainings
	var group models.MstrGroup
	if err := config.DB.Preload("Chainings").First(&group, groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Group not found"})
		return
	}

	// Ambil semua chaining yang akan di-assign
	var chainings []models.MstrChaining
	if len(req.ChainingIDs) > 0 {
		if err := config.DB.Where("id IN ?", req.ChainingIDs).Find(&chainings).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
			return
		}
	}

	// Hapus semua chaining dari group
	if err := config.DB.Model(&group).Association("Chainings").Clear(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to clear chainings from group"})
		return
	}

	// Assign kembali chaining baru
	if len(chainings) > 0 {
		if err := config.DB.Model(&group).Association("Chainings").Append(&chainings); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to assign chainings to group"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Group chainings updated successfully",
	})
}

// GET /groups/:id/chainings
func GetGroupChainings(c *gin.Context) {
	groupID := c.Param("id")

	// Ambil company_id dari JWT context
	companyID, ok := c.Get("company_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Company ID not found in token"})
		return
	}

	var group models.MstrGroup
	if err := config.DB.
		Preload("Chainings", "company_id = ?", companyID).
		Where("company_id = ?", companyID).
		First(&group, groupID).Error; err != nil {
		c.JSON(404, gin.H{"status": "error", "message": "Group not found"})
		return
	}

	var allChainings []models.MstrChaining
	if err := config.DB.Where("company_id = ? AND is_active = ?", companyID, true).Find(&allChainings).Error; err != nil {
		c.JSON(500, gin.H{"status": "error", "message": err.Error()})
		return
	}

	// Ambil hanya id chaining yang sudah ter-assign ke group
	assignedIDs := make([]uint, len(group.Chainings))
	for i, ch := range group.Chainings {
		assignedIDs[i] = ch.Id
	}

	c.JSON(200, gin.H{
		"status":                "success",
		"all_chainings":         allChainings,
		"assigned_chaining_ids": assignedIDs,
	})
}
