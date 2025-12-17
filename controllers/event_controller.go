package controllers

import (
	"go-api/config"
	"go-api/models"
	"go-api/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateEvent(c *gin.Context) {
	var event models.MstrEventTrigger
	if err := c.ShouldBindJSON(&event); err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Isi metadata
	userCompanyID := c.GetString("company_id")
	username := c.GetString("username")

	event.CompanyID = userCompanyID
	event.CreatedBy = username
	event.UpdatedBy = username

	if err := config.DB.Create(&event).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Event created", event)
}

func UpdateEventByID(c *gin.Context) {
	eventID := c.Param("id")

	var event models.MstrEventTrigger
	if err := config.DB.
		Where("id = ?", eventID).
		First(&event).Error; err != nil {
		utils.JSONError(c, http.StatusNotFound, "Event not found")
		return
	}

	var input struct {
		EventName   string `json:"event_name"`
		Description string `json:"description"`
		Trigger     bool   `json:"trigger"`
		IsActive    bool   `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Update field
	event.EventName = input.EventName
	event.Description = input.Description
	event.Trigger = input.Trigger
	event.IsActive = input.IsActive
	userCompanyID := c.GetString("company_id")
	event.CompanyID = userCompanyID
	username := c.GetString("username")
	event.UpdatedBy = username

	if err := config.DB.Save(&event).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Event updated", event)
}

func DeleteEventByID(c *gin.Context) {
	eventID := c.Param("id")
	deletedBy := c.GetString("username")

	// Set DeletedBy
	if err := config.DB.Model(&models.MstrEventTrigger{}).
		Where("id = ?", eventID).
		Update("deleted_by", deletedBy).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	if err := config.DB.Delete(&models.MstrEventTrigger{}, eventID).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.JSONSuccess(c, "Event deleted", nil)
}

func GetFilteredEvents(c *gin.Context) {
	createdBy := c.Query("created_by")
	updatedBy := c.Query("updated_by")
	eventName := c.Query("event_name")
	eventID := c.Query("id")

	var events []models.MstrEventTrigger
	query := config.DB.Model(&models.MstrEventTrigger{})

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
	if eventName != "" {
		query = query.Where("event_name ILIKE ?", "%"+eventName+"%")
	}
	if eventID != "" {
		query = query.Where("id = ?", eventID)
	}

	if err := query.Order("id DESC").Find(&events).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Filtered events", events)
}
