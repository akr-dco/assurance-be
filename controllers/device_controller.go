package controllers

import (
	"go-api/config"
	"go-api/models"
	"go-api/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateDevice(c *gin.Context) {
	var device models.MstrDevice
	if err := c.ShouldBindJSON(&device); err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Cek apakah device_id sudah ada
	var existing models.MstrDevice
	if err := config.DB.Where("device_id = ? and company_id = ?", device.DeviceID, device.CompanyID).First(&existing).Error; err == nil {

		utils.JSONSuccess(c, "Device ID already exists", existing)
		return
	}

	// Jika tidak ada, lanjutkan simpan
	device.CreatedBy = "system"
	device.UpdatedBy = "system"
	device.IsActive = false

	if err := config.DB.Create(&device).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Device created", device)
}

// Update Device
func UpdateDeviceByID(c *gin.Context) {
	deviceID := c.Param("id")

	// Cek apakah device ada untuk company tersebut
	var device models.MstrDevice
	if err := config.DB.
		Where("id = ?", deviceID).
		First(&device).Error; err != nil {
		utils.JSONError(c, http.StatusNotFound, "Device not found")
		return
	}

	// Bind data update dari body JSON
	var input struct {
		DeviceName string `json:"device_name"`
		IsActive   bool   `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Update field
	device.DeviceName = input.DeviceName
	device.IsActive = input.IsActive
	userCompanyID, _ := c.Get("company_id")
	device.CompanyID = userCompanyID.(string)
	username, _ := c.Get("username")
	device.UpdatedBy = username.(string)

	if err := config.DB.Save(&device).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Device updated", device)
}

func DeleteDeviceByID(c *gin.Context) {
	id := c.Param("id")

	deletedBy := c.GetString("username")

	// Set DeletedBy
	if err := config.DB.Model(&models.MstrDevice{}).
		Where("id = ?", id).
		Update("deleted_by", deletedBy).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	if err := config.DB.Delete(&models.MstrDevice{}, id).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.JSONSuccess(c, "Device deleted", nil)
}

// Get Filtered Devices
func GetFilteredDevices(c *gin.Context) {
	createdBy := c.Query("created_by")
	updatedBy := c.Query("updated_by")
	deviceName := c.Query("device_name")
	deviceID := c.Query("device_id")

	var devices []models.MstrDevice
	query := config.DB.Model(&models.MstrDevice{})

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
	if deviceName != "" {
		query = query.Where("device_name ILIKE ?", "%"+deviceName+"%")
	}
	if deviceID != "" {
		query = query.Where("device_id = ?", deviceID)
	}

	if err := query.Order("id DESC").Find(&devices).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Filtered devices", devices)
}

func GetDeviceInspection(c *gin.Context) {
	deviceID := c.Param("deviceID")

	query := `
        SELECT mi.id, mi.name_inspection, mi.image_url, mi.created_at, 
               mi.updated_at, mi.created_by, mi.updated_by, mi.company_id
        FROM mstr_device md
        JOIN mstr_group_device mgd ON mgd.mstr_device_id = md.id
        JOIN mstr_group mg ON mg.id = mgd.mstr_group_id
        JOIN mstr_group_inspection mgi ON mgi.mstr_group_id = mg.id
        JOIN mstr_inspection mi ON mi.id = mgi.mstr_inspection_id
        WHERE md.device_id = ?
    `

	var inspections []models.MstrInspection
	if err := config.DB.Raw(query, deviceID).Scan(&inspections).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	if len(inspections) == 0 {
		utils.JSONError(c, http.StatusNotFound, "No inspections found for this device")
		return
	}

	utils.JSONSuccess(c, "Device inspections fetched successfully", inspections)
}

func GetDevicePreInspection(c *gin.Context) {
	deviceID := c.Param("deviceID")

	query := `
        SELECT q.id, q.title, q.description, q.company_id, q.created_by,
               q.created_at, q.updated_by, q.updated_at, q.type, q.is_active
        FROM mstr_device md
        JOIN mstr_group_device mgd ON mgd.mstr_device_id = md.id
        JOIN mstr_group mg ON mg.id = mgd.mstr_group_id
        JOIN mstr_group_questionnaire mgq ON mgq.mstr_group_id = mg.id
        JOIN questionnaires q ON q.id = mgq.questionnaire_id
        WHERE q.type = 'Pre-Inspection' AND md.device_id = ?
    `

	var questionnaires []models.Questionnaire
	if err := config.DB.Raw(query, deviceID).Scan(&questionnaires).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	if len(questionnaires) == 0 {
		utils.JSONError(c, http.StatusNotFound, "No Pre-Inspections found for this device")
		return
	}

	utils.JSONSuccess(c, "Device Pre-Inspections fetched successfully", questionnaires)
}

func GetDevicePostInspection(c *gin.Context) {
	deviceID := c.Param("deviceID")

	query := `
        SELECT q.id, q.title, q.description, q.company_id, q.created_by,
               q.created_at, q.updated_by, q.updated_at, q.type, q.is_active
        FROM mstr_device md
        JOIN mstr_group_device mgd ON mgd.mstr_device_id = md.id
        JOIN mstr_group mg ON mg.id = mgd.mstr_group_id
        JOIN mstr_group_questionnaire mgq ON mgq.mstr_group_id = mg.id
        JOIN questionnaires q ON q.id = mgq.questionnaire_id
        WHERE q.type = 'Post-Inspection' AND md.device_id = ?
    `

	var questionnaires []models.Questionnaire
	if err := config.DB.Raw(query, deviceID).Scan(&questionnaires).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	if len(questionnaires) == 0 {
		utils.JSONError(c, http.StatusNotFound, "No Post-Inspections found for this device")
		return
	}

	utils.JSONSuccess(c, "Device Post-Inspections fetched successfully", questionnaires)
}

func GetChainingByDevice(c *gin.Context) {
	deviceID := c.Param("deviceID")

	query := `
        SELECT mc.id, mc.name_chaining, mc.is_active, mc.company_id, mc.created_at, 
               mc.updated_at, mc.created_by, mc.updated_by
        FROM mstr_device md
        JOIN mstr_group_device mgd ON mgd.mstr_device_id = md.id
        JOIN mstr_group mg ON mg.id = mgd.mstr_group_id
        JOIN mstr_group_chaining mgc ON mgc.mstr_group_id = mg.id
        JOIN mstr_chainings mc ON mc.id = mgc.mstr_chaining_id
        WHERE mc.is_active = true and md.device_id = ?
    `

	var chaining []models.MstrChaining
	if err := config.DB.Raw(query, deviceID).Scan(&chaining).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	if len(chaining) == 0 {
		utils.JSONError(c, http.StatusNotFound, "No chaining found for this device")
		return
	}

	utils.JSONSuccess(c, "Device chaining fetched successfully", chaining)
}
