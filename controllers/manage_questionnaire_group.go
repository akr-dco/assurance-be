package controllers

import (
	"fmt"
	"go-api/config"
	"go-api/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ManageGroupQuestionnaireBulk(c *gin.Context) {
	groupID := c.Param("groupId")

	var req struct {
		QuestionnaireIDs []uint `json:"questionnaire_ids" binding:"required"`
		Type             string `json:"type" binding:"required"` // Pre-Inspection atau Post-Inspection
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	// Ambil group
	var group models.MstrGroup
	if err := config.DB.First(&group, groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Group not found"})
		return
	}

	// Hapus hanya questionnaire dari group berdasarkan type
	if err := config.DB.Exec(`
		DELETE FROM mstr_group_questionnaire 
		WHERE mstr_group_id = ? 
		AND questionnaire_id IN (
			SELECT id FROM questionnaires WHERE type = ?
		)
	`, group.ID, req.Type).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to clear questionnaires by type"})
		return
	}

	// Ambil questionnaires baru (hanya sesuai type)
	var questionnaires []models.Questionnaire
	if len(req.QuestionnaireIDs) > 0 {
		if err := config.DB.Where("id IN ? AND type = ?", req.QuestionnaireIDs, req.Type).Find(&questionnaires).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
			return
		}
	}

	// Assign kembali questionnaire baru
	if len(questionnaires) > 0 {
		if err := config.DB.Model(&group).Association("Questionnaires").Append(&questionnaires); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to assign questionnaires to group"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": fmt.Sprintf("Group questionnaires updated successfully for type %s", req.Type),
	})
}

// GET /groups/:id/questionnaires
func GetGroupQuestionnaires(c *gin.Context) {
	groupID := c.Param("id")

	// Ambil company_id dari JWT context
	companyID, ok := c.Get("company_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Company ID not found in token"})
		return
	}

	var group models.MstrGroup
	if err := config.DB.
		Preload("Questionnaires", "company_id = ?", companyID).
		Where("company_id = ?", companyID).
		First(&group, groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Group not found"})
		return
	}

	var allQuestionnaires []models.Questionnaire
	if err := config.DB.Where("company_id = ? AND is_active = ?", companyID, true).Find(&allQuestionnaires).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	// Ambil hanya id questionnaire yang sudah ter-assign ke group
	assignedIDs := make([]uint, len(group.Questionnaires))
	for i, q := range group.Questionnaires {
		assignedIDs[i] = q.ID
	}

	c.JSON(http.StatusOK, gin.H{
		"status":                     "success",
		"all_questionnaires":         allQuestionnaires,
		"assigned_questionnaire_ids": assignedIDs,
	})
}
