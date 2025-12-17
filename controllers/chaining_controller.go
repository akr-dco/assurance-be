package controllers

import (
	"go-api/config"
	"go-api/models"
	"go-api/utils"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Create Chaining
func CreateChaining(c *gin.Context) {
	var chaining models.MstrChaining
	if err := c.ShouldBindJSON(&chaining); err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	userCompanyID, _ := c.Get("company_id")
	username, _ := c.Get("username")

	chaining.CompanyID = userCompanyID.(string)
	chaining.CreatedBy = username.(string)
	chaining.UpdatedBy = username.(string)

	if err := config.DB.Create(&chaining).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Chaining created", chaining)
}

func UpdateChainingByID(c *gin.Context) {
	chainingID := c.Param("id")
	var chaining models.MstrChaining

	// 1. Ambil data master
	if err := config.DB.First(&chaining, chainingID).Error; err != nil {
		utils.JSONError(c, http.StatusNotFound, "Chaining not found")
		return
	}

	var input models.MstrChaining
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	// 2. Gunakan GORM Transaction untuk operasi atomik
	err := config.DB.Transaction(func(tx *gorm.DB) error {
		// Hapus detail yang ada di DB tapi tidak di input
		var inputDetailIDs []uint
		for _, detail := range input.Details {
			if detail.Id != 0 {
				inputDetailIDs = append(inputDetailIDs, detail.Id)
			}
		}

		// Cek apakah ada detail lama yang harus dihapus
		var oldDetails []models.MstrChainingDetail
		tx.Where("id_chaining = ?", chaining.Id).Find(&oldDetails)

		for _, oldDetail := range oldDetails {
			found := false
			for _, newID := range inputDetailIDs {
				if oldDetail.Id == newID {
					found = true
					break
				}
			}
			if !found {
				if err := tx.Delete(&oldDetail).Error; err != nil {
					return err // Rollback jika gagal
				}
			}
		}

		// Update master
		chaining.NameChaining = input.NameChaining
		chaining.EventTriggerID = input.EventTriggerID
		chaining.TriggerDatetime = input.TriggerDatetime
		chaining.FrequencyValue = input.FrequencyValue
		chaining.FrequencyUnit = input.FrequencyUnit
		chaining.IsActive = input.IsActive
		//username := c.GetString("username")
		chaining.UpdatedBy = c.GetString("username")
		if err := tx.Save(&chaining).Error; err != nil {
			return err
		}

		// Loop dan simpan/buat detail baru atau yang diperbarui
		for _, detailInput := range input.Details {
			detailInput.IdChaining = chaining.Id
			detailInput.CreatedBy = chaining.UpdatedBy
			detailInput.UpdatedBy = chaining.UpdatedBy
			if err := tx.Save(&detailInput).Error; err != nil {
				return err // Rollback jika gagal
			}
		}
		return nil // Transaksi sukses
	})

	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 3. Muat ulang data untuk respon
	var updated models.MstrChaining
	config.DB.Preload("Details").First(&updated, chaining.Id)
	utils.JSONSuccess(c, "Chaining updated successfully", updated)
}

// Delete Chaining
func DeleteChainingByID(c *gin.Context) {
	id := c.Param("id")

	deletedBy := c.GetString("username")

	// Set DeletedBy
	if err := config.DB.Model(&models.MstrChaining{}).
		Where("id = ?", id).
		Update("deleted_by", deletedBy).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	if err := config.DB.Delete(&models.MstrChaining{}, id).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.JSONSuccess(c, "Chaining deleted", nil)
}

// Get Filtered Chaining
func GetFilteredChainings(c *gin.Context) {
	createdBy := c.Query("created_by")
	updatedBy := c.Query("updated_by")
	nameChaining := c.Query("name_chaining")

	var chainings []models.MstrChaining
	query := config.DB.Model(&models.MstrChaining{}).
		Preload("Events").
		Preload("Details")

	//config.DB.Debug().Preload("Events").Preload("Details").First(&chainings, id)

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
	if nameChaining != "" {
		query = query.Where("name_chaining ILIKE ?", "%"+nameChaining+"%")
	}

	if err := query.Order("id DESC").Find(&chainings).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Filtered chainings", chainings)
}

func GetChainingByID(c *gin.Context) {
	id := c.Param("id")
	var chaining models.MstrChaining

	// Ambil chaining + details (urut berdasarkan sequence)
	if err := config.DB.
		Preload("Events").
		Preload("Details", func(db *gorm.DB) *gorm.DB {
			return db.Order("sequence ASC")
		}).
		First(&chaining, id).Error; err != nil {
		utils.JSONError(c, http.StatusNotFound, "Chaining not found")
		return
	}

	// Loop isi details, inject data questionnaire / inspection
	for i, d := range chaining.Details {
		if d.ItemType == "inspection" {
			var ins models.MstrInspection
			if err := config.DB.Preload("Details.Questions.Options").First(&ins, d.ItemID).Error; err == nil {
				chaining.Details[i].Inspection = &ins
			}
		} else if d.ItemType == "questionnaire" {
			var q models.Questionnaire
			if err := config.DB.Preload("Questions.Options").First(&q, d.ItemID).Error; err == nil {
				chaining.Details[i].Questionnaire = &q
			}
		}
	}

	utils.JSONSuccess(c, "Chaining found", chaining)
}

func GetChainingByDeviceNew(c *gin.Context) {
	deviceID := c.Param("deviceID")
	userID := c.GetString("username")

	if userID == "" {
		utils.JSONError(c, http.StatusBadRequest, "user parameter is required")
		return
	}

	// ============================
	// GET DEVICE TIMEZONE
	// ============================
	tz := c.GetHeader("X-Timezone")
	if tz == "" {
		tz = "UTC"
	}

	userLoc, err := time.LoadLocation(tz)
	if err != nil {
		log.Println("[WARN] Invalid timezone from header, fallback to UTC")
		userLoc = time.UTC
	}

	log.Println("[TIMEZONE] Using device timezone:", userLoc)

	query := `
	SELECT 
		mc.id, 
		mc.name_chaining, 
		mc.trigger_datetime, 
		mc.frequency_value, 
		mc.frequency_unit,
		mc.event_trigger_id,
		et.event_name,               
		et.trigger AS event_trigger_active,
		mc.is_active
	FROM mstr_device md
	JOIN mstr_group_device mgd ON mgd.mstr_device_id = md.id 
	JOIN mstr_group mg ON mg.id = mgd.mstr_group_id AND mg.deleted_at IS NULL
	JOIN mstr_group_chaining mgc ON mgc.mstr_group_id = mg.id 
	JOIN mstr_chainings mc ON mc.id = mgc.mstr_chaining_id AND mc.deleted_at IS NULL
	LEFT JOIN mstr_event_triggers et ON et.id = mc.event_trigger_id AND et.deleted_at IS NULL
	WHERE mc.is_active = true 
	AND md.device_id = ?
	AND md.deleted_at IS NULL
	`

	type ChainingWithTrigger struct {
		Id                 uint
		NameChaining       string
		TriggerDatetime    time.Time // stored UTC from DB
		FrequencyValue     uint
		FrequencyUnit      string
		EventTriggerID     *uint
		EventName          string
		EventTriggerActive bool
		IsActive           bool
	}

	var chainings []ChainingWithTrigger
	if err := config.DB.Raw(query, deviceID).Scan(&chainings).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// CURRENT TIME IN DEVICE TIMEZONE
	nowUTC := time.Now().UTC()
	nowLocal := nowUTC.In(userLoc)

	var activeChains []gin.H

	for _, chain := range chainings {

		// Convert trigger from DB UTC → DEVICE TIMEZONE
		triggerLocal := chain.TriggerDatetime.UTC().In(userLoc)

		// Abaikan chaining yang belum waktunya
		if triggerLocal.IsZero() || triggerLocal.After(nowLocal) {
			continue
		}

		var windowStartLocal, windowEndLocal time.Time

		// =====================
		// MODE A: ONE-TIME ONLY
		// =====================
		if chain.FrequencyValue == 0 || chain.FrequencyUnit == "" {
			windowStartLocal = triggerLocal
			windowEndLocal = time.Date(9999, 12, 31, 23, 59, 59, 0, userLoc)

		} else {
			// =====================
			// MODE B: PERIODIC
			// =====================
			windowStartLocal = getCurrentTriggerWindowDevice(
				triggerLocal,
				int(chain.FrequencyValue),
				chain.FrequencyUnit,
				nowLocal,
			)

			windowEndLocal = windowStartLocal.Add(
				calcDuration(chain.FrequencyValue, chain.FrequencyUnit),
			)
		}

		// Ambil detail chaining
		var details []struct {
			ID       int
			ItemType string
			ItemID   int
			Sequence int
		}
		config.DB.Table("mstr_chaining_details").
			Select("id, item_type, item_id, sequence").
			Where("id_chaining = ? AND deleted_at IS NULL", chain.Id).
			Order("sequence ASC").
			Scan(&details)

		if len(details) == 0 {
			continue
		}

		var activeItems []gin.H
		allDone := true

		// CEK apakah item sudah dikerjakan oleh user pada window ini
		for _, d := range details {
			var doneAt time.Time
			var itemName string

			switch d.ItemType {
			case "inspection":
				config.DB.Raw(`SELECT name_inspection FROM mstr_inspection WHERE id = ? LIMIT 1`,
					d.ItemID,
				).Scan(&itemName)

				config.DB.Raw(`
					SELECT created_at FROM trx_inspection
					WHERE id_inspection = ? AND chaining_id = ? AND created_by = ?
					AND created_at >= ? AND created_at < ?
					ORDER BY created_at DESC LIMIT 1`,
					d.ItemID, chain.Id, userID,
					windowStartLocal.UTC(), windowEndLocal.UTC(),
				).Scan(&doneAt)

			case "questionnaire":
				config.DB.Raw(`SELECT title FROM questionnaires WHERE id = ? LIMIT 1`,
					d.ItemID,
				).Scan(&itemName)

				config.DB.Raw(`
					SELECT created_at FROM mstr_answer
					WHERE questionnaire_id = ? AND chaining_id = ? AND created_by = ?
					AND created_at >= ? AND created_at < ?
					ORDER BY created_at DESC LIMIT 1`,
					d.ItemID, chain.Id, userID,
					windowStartLocal.UTC(), windowEndLocal.UTC(),
				).Scan(&doneAt)
			}

			if doneAt.IsZero() {
				allDone = false
				activeItems = append(activeItems, gin.H{
					"id_detail": d.ID,
					"item_type": d.ItemType,
					"item_id":   d.ItemID,
					"item_name": itemName,
					"sequence":  d.Sequence,
				})
			}
		}

		if !allDone {
			if (chain.EventName == "" || chain.EventName == "<nil>") && len(activeItems) > 0 {
				chain.EventTriggerActive = true
			}

			activeChains = append(activeChains, gin.H{
				"id":                   chain.Id,
				"name_chaining":        chain.NameChaining,
				"trigger_time_local":   windowStartLocal,
				"trigger_time_utc":     windowStartLocal.UTC(),
				"frequency_unit":       chain.FrequencyUnit,
				"frequency_value":      chain.FrequencyValue,
				"event_trigger_id":     chain.EventTriggerID,
				"event_name":           chain.EventName,
				"event_trigger_active": chain.EventTriggerActive,
				"active_items":         activeItems,
				"timezone":             userLoc.String(),
			})
		}
	}

	if len(activeChains) == 0 {
		utils.JSONError(c, http.StatusNotFound, "No active chaining for this user")
		return
	}

	utils.JSONSuccess(c, "Active chaining fetched successfully", activeChains)
}

func getCurrentTriggerWindowDevice(triggerLocal time.Time, freqValue int, freqUnit string, nowLocal time.Time) time.Time {
	if freqValue <= 0 || freqUnit == "" {
		return triggerLocal
	}

	diff := nowLocal.Sub(triggerLocal)
	if diff < 0 {
		return triggerLocal
	}

	var cycle time.Duration

	switch strings.ToLower(freqUnit) {
	case "hour", "hours", "hourly":
		cycle = time.Duration(freqValue) * time.Hour
	case "day", "daily":
		cycle = time.Duration(freqValue) * 24 * time.Hour
	case "week", "weekly":
		cycle = time.Duration(freqValue) * 7 * 24 * time.Hour
	default:
		cycle = time.Duration(freqValue) * time.Hour
	}

	cycles := int(diff / cycle)
	return triggerLocal.Add(time.Duration(cycles) * cycle)
}

func calcDuration(freqValue uint, freqUnit string) time.Duration {
	switch strings.ToLower(freqUnit) {
	case "hour", "hours", "hourly":
		return time.Duration(freqValue) * time.Hour
	case "day", "daily":
		return time.Duration(freqValue) * 24 * time.Hour
	case "week", "weekly":
		return time.Duration(freqValue) * 7 * 24 * time.Hour
	default:
		return time.Hour
	}
}
