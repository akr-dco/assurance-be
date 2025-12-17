package controllers

import (
	"encoding/json"
	"fmt"
	"go-api/config"
	"go-api/models"
	"go-api/utils"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// SAM UPDATE
func UpdateMstrInspectionDetailByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid detail id")
		return
	}

	username, _ := c.Get("username")

	// ambil body JSON
	var req models.MstrInspectionDetail
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	tx := config.DB.Begin()
	if tx.Error != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Failed to start transaction")
		return
	}

	// cek apakah detail ada
	var detail models.MstrInspectionDetail
	if err := tx.First(&detail, id).Error; err != nil {
		tx.Rollback()
		utils.JSONError(c, http.StatusNotFound, "Detail not found")
		return
	}

	// update field detail
	detail.NameCoordinate = req.NameCoordinate
	detail.X = req.X
	detail.Y = req.Y
	detail.TutorialCoordinate = req.TutorialCoordinate
	detail.RequiredCoordinate = req.RequiredCoordinate
	detail.SendNow = req.SendNow
	detail.TypeTriggerID = req.TypeTriggerID
	detail.UpdatedBy = username.(string)
	if err := tx.Save(&detail).Error; err != nil {
		tx.Rollback()
		utils.JSONError(c, http.StatusInternalServerError, "Failed to update detail: "+err.Error())
		return
	}

	// id-id question baru dari request
	var questionIDs []uint
	for _, qd := range req.Questions {
		var q models.MstrInspectionQuestion

		if qd.ID != 0 {
			// update existing
			if err := tx.First(&q, qd.ID).Error; err == nil {
				q.Text = qd.Text
				q.Type = qd.Type
				q.UpdatedBy = username.(string)
				tx.Save(&q)
			}
		} else {
			// insert baru
			q = models.MstrInspectionQuestion{
				InspectionDetailID: detail.Id,
				Text:               qd.Text,
				Type:               qd.Type,
				CreatedBy:          username.(string),
				UpdatedBy:          username.(string),
			}
			tx.Create(&q)
		}

		questionIDs = append(questionIDs, q.ID)

		// === proses options ===
		var optionIDs []uint
		for _, od := range qd.Options {
			var o models.MstrInspectionQuestionOption

			if od.ID != 0 {
				if err := tx.First(&o, od.ID).Error; err == nil {
					o.Label = od.Label
					o.Text = od.Text
					o.IsCorrect = od.IsCorrect
					o.UpdatedBy = username.(string)
					tx.Save(&o)
				}
			} else {
				o = models.MstrInspectionQuestionOption{
					InspectionQuestionID: q.ID,
					Label:                od.Label,
					Text:                 od.Text,
					IsCorrect:            od.IsCorrect,
					CreatedBy:            username.(string),
					UpdatedBy:            username.(string),
				}
				tx.Create(&o)
			}

			optionIDs = append(optionIDs, o.ID)
		}

		// hapus option yang tidak ada di request
		if len(optionIDs) > 0 {
			tx.Where("inspection_question_id = ? AND id NOT IN ?", q.ID, optionIDs).
				Delete(&models.MstrInspectionQuestionOption{})
		} else {
			tx.Where("inspection_question_id = ?", q.ID).
				Delete(&models.MstrInspectionQuestionOption{})
		}
	}

	// hapus question yang tidak ada di request
	if len(questionIDs) > 0 {
		tx.Where("inspection_detail_id = ? AND id NOT IN ?", detail.Id, questionIDs).
			Delete(&models.MstrInspectionQuestion{})
	} else {
		tx.Where("inspection_detail_id = ?", detail.Id).
			Delete(&models.MstrInspectionQuestion{})
	}

	if err := tx.Commit().Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Failed to commit transaction: "+err.Error())
		return
	}

	// ambil kembali detail yang sudah diupdate
	var updated models.MstrInspectionDetail
	if err := config.DB.
		Preload("Questions.Options").
		First(&updated, detail.Id).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Detail updated successfully", updated)
}

// SAM CREATE
func CreateMstrInspectionDetail(c *gin.Context) {
	username, _ := c.Get("username")

	var req models.MstrInspectionDetail
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	// Pastikan ada inspection_id (parent)
	if req.IdMstrInspection == 0 {
		utils.JSONError(c, http.StatusBadRequest, "Missing inspection_id")
		return
	}

	tx := config.DB.Begin()
	if tx.Error != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Failed to start transaction")
		return
	}

	const offset = 5.0
	// Terapkan offset acak ringan (random positif atau negatif)
	randOffsetX := (rand.Float64() - 0.5) * offset
	randOffsetY := (rand.Float64() - 0.5) * offset

	// ===== COPY DETAIL =====
	newDetail := models.MstrInspectionDetail{
		IdMstrInspection:   req.IdMstrInspection,
		NameCoordinate:     req.NameCoordinate,
		X:                  req.X + randOffsetX,
		Y:                  req.Y + randOffsetY,
		TutorialCoordinate: req.TutorialCoordinate,
		RequiredCoordinate: req.RequiredCoordinate,
		SendNow:            req.SendNow,
		TypeTriggerID:      req.TypeTriggerID,
		CreatedBy:          username.(string),
		UpdatedBy:          username.(string),
	}

	if err := tx.Create(&newDetail).Error; err != nil {
		tx.Rollback()
		utils.JSONError(c, http.StatusInternalServerError, "Failed to copy detail: "+err.Error())
		return
	}

	// ===== COPY QUESTIONS & OPTIONS =====
	for _, qd := range req.Questions {
		newQ := models.MstrInspectionQuestion{
			InspectionDetailID: newDetail.Id,
			Text:               qd.Text,
			Type:               qd.Type,
			CreatedBy:          username.(string),
			UpdatedBy:          username.(string),
		}
		if err := tx.Create(&newQ).Error; err != nil {
			tx.Rollback()
			utils.JSONError(c, http.StatusInternalServerError, "Failed to copy question: "+err.Error())
			return
		}

		for _, od := range qd.Options {
			newO := models.MstrInspectionQuestionOption{
				InspectionQuestionID: newQ.ID,
				Label:                od.Label,
				Text:                 od.Text,
				IsCorrect:            od.IsCorrect,
				CreatedBy:            username.(string),
				UpdatedBy:            username.(string),
			}
			if err := tx.Create(&newO).Error; err != nil {
				tx.Rollback()
				utils.JSONError(c, http.StatusInternalServerError, "Failed to copy option: "+err.Error())
				return
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Failed to commit transaction: "+err.Error())
		return
	}

	utils.JSONSuccess(c, "Detail copied successfully", newDetail)
}

// SAM MOVE
func UpdateInspectionPosition(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	if err := config.DB.Model(&models.MstrInspectionDetail{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"x": req.X,
			"y": req.Y,
		}).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Failed to update coordinates: "+err.Error())
		return
	}

	utils.JSONSuccess(c, "Coordinate updated successfully", nil)
}

// SAM DELETE
func DeleteMstrInspectionDetailByID(c *gin.Context) {
	id := c.Param("id")

	deletedBy := c.GetString("username")

	// Set DeletedBy
	if err := config.DB.Model(&models.MstrInspectionDetail{}).
		Where("id = ?", id).
		Update("deleted_by", deletedBy).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	if err := config.DB.Delete(&models.MstrInspectionDetail{}, id).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.JSONSuccess(c, "SAM deleted", nil)
}

// ASSURANCE MASTER CREATED
func CreateMstrInspection(c *gin.Context) {
	name := c.PostForm("name_inspection")
	detailJSON := c.PostForm("details")

	// Ambil info user
	userCompanyID := c.GetString("company_id")
	username := c.GetString("username")

	// === Upload file ===
	fileHeader, err := c.FormFile("image")
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Image upload required: "+err.Error())
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Cannot open file: "+err.Error())
		return
	}
	defer file.Close()

	fileKey := GenerateE2ObjectKey(c, "Master-Assurance", fileHeader.Filename)
	objectKey, err := UploadFileToE2(c, file, fileKey, fileHeader.Header.Get("Content-Type"), "Assurance/"+userCompanyID, nil)
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

	// === Mulai transaksi ===
	tx := config.DB.Begin()
	if tx.Error != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Failed to start transaction")
		return
	}

	// === Simpan master inspection ===
	inspection := models.MstrInspection{
		NameInspection: name,
		ImageUrl:       objectKey,
		CompanyID:      userCompanyID,
		CreatedBy:      username,
		UpdatedBy:      username,
	}

	if err := tx.Create(&inspection).Error; err != nil {
		tx.Rollback()
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// === Proses details + questions + options ===
	if detailJSON != "" {
		var details []models.MstrInspectionDetail
		if err := json.Unmarshal([]byte(detailJSON), &details); err != nil {
			tx.Rollback()
			utils.JSONError(c, http.StatusBadRequest, "Invalid details format: "+err.Error())
			return
		}

		for i := range details {
			// reset ID detail
			detail := models.MstrInspectionDetail{
				IdMstrInspection:   inspection.Id,
				NameCoordinate:     details[i].NameCoordinate,
				X:                  details[i].X,
				Y:                  details[i].Y,
				TutorialCoordinate: details[i].TutorialCoordinate,
				RequiredCoordinate: details[i].RequiredCoordinate,
				SendNow:            details[i].SendNow,
				TypeTriggerID:      details[i].TypeTriggerID,
				CreatedBy:          username,
				UpdatedBy:          username,
			}

			// Simpan detail
			if err := tx.Create(&detail).Error; err != nil {
				tx.Rollback()
				utils.JSONError(c, http.StatusInternalServerError, "Failed to save detail: "+err.Error())
				return
			}

			// === Simpan questions ===
			for j := range details[i].Questions {
				q := models.MstrInspectionQuestion{
					InspectionDetailID: detail.Id,
					Text:               details[i].Questions[j].Text,
					Type:               details[i].Questions[j].Type,
					CreatedBy:          username,
					UpdatedBy:          username,
				}

				if err := tx.Create(&q).Error; err != nil {
					tx.Rollback()
					utils.JSONError(c, http.StatusInternalServerError, "Failed to save question: "+err.Error())
					return
				}

				// === Simpan options ===
				for k := range details[i].Questions[j].Options {
					o := models.MstrInspectionQuestionOption{
						InspectionQuestionID: q.ID,
						Label:                details[i].Questions[j].Options[k].Label,
						Text:                 details[i].Questions[j].Options[k].Text,
						IsCorrect:            details[i].Questions[j].Options[k].IsCorrect,
						CreatedBy:            username,
						UpdatedBy:            username,
					}

					if err := tx.Create(&o).Error; err != nil {
						tx.Rollback()
						utils.JSONError(c, http.StatusInternalServerError, "Failed to save option: "+err.Error())
						return
					}
				}
			}
		}
	}

	// === Commit transaksi ===
	if err := tx.Commit().Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Failed to commit transaction: "+err.Error())
		return
	}

	// === Ambil kembali inspection lengkap dengan relasi ===
	var saved models.MstrInspection
	if err := config.DB.
		Preload("Details.Questions.Options").
		First(&saved, inspection.Id).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Assurance master created", saved)
}

// ASSURANCE MASTER DELETE
func DeleteMstrInspectionByID(c *gin.Context) {
	id := c.Param("id")

	deletedBy := c.GetString("username")

	// Set DeletedBy
	if err := config.DB.Model(&models.MstrInspection{}).
		Where("id = ?", id).
		Update("deleted_by", deletedBy).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	if err := config.DB.Delete(&models.MstrInspection{}, id).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.JSONSuccess(c, "Assurance master deleted", nil)
}

// ASSURANCE MASTER LIST
func GetFilteredInspections(c *gin.Context) {
	createdBy := c.Query("created_by")
	updatedBy := c.Query("updated_by")
	nameInspection := c.Query("name_inspection")
	idInspection := c.Query("id_inspection")

	var inspections []models.MstrInspection
	query := config.DB.Preload("Details.Questions.Options")

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

	if nameInspection != "" {
		query = query.Where("name_inspection ILIKE ?", "%"+nameInspection+"%")
	}

	if idInspection != "" {
		query = query.Where("id = ?", idInspection)
	}

	query = query.Order("id DESC")

	if err := query.Find(&inspections).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Filtered Assurance Master", inspections)
}

// ASSURANCE UPDATE
func UpdateMstrInspectionByID_X(c *gin.Context) {
	id := c.Param("id")

	var inspection models.MstrInspection
	if err := config.DB.Preload("Details.Questions.Options").
		First(&inspection, id).Error; err != nil {
		utils.JSONError(c, http.StatusNotFound, "Inspection not found")
		return
	}

	var payload struct {
		NameInspection string                        `json:"name_inspection"`
		Details        []models.MstrInspectionDetail `json:"details"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid payload: "+err.Error())
		return
	}

	username, _ := c.Get("username")

	// --- Update inspection utama ---
	inspection.NameInspection = payload.NameInspection
	inspection.UpdatedBy = username.(string)
	inspection.UpdatedAt = time.Now()
	if err := config.DB.Save(&inspection).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Buat map detail lama untuk deteksi delete
	existingDetails := make(map[uint]models.MstrInspectionDetail)
	for _, d := range inspection.Details {
		existingDetails[d.Id] = d
	}

	// --- Proses details ---
	for _, detail := range payload.Details {
		if detail.Id == 0 {
			// INSERT detail baru
			detail.IdMstrInspection = inspection.Id
			detail.CreatedBy = username.(string)
			detail.UpdatedBy = username.(string)
			if err := config.DB.Create(&detail).Error; err != nil {
				utils.JSONError(c, http.StatusInternalServerError, "Failed insert detail: "+err.Error())
				return
			}
		} else {
			// UPDATE detail lama
			if _, ok := existingDetails[detail.Id]; ok {
				detail.UpdatedBy = username.(string)
				if err := config.DB.Model(&models.MstrInspectionDetail{}).
					Where("id = ?", detail.Id).
					Updates(detail).Error; err != nil {
					utils.JSONError(c, http.StatusInternalServerError, "Failed update detail: "+err.Error())
					return
				}
			}
			delete(existingDetails, detail.Id)
		}

		// Proses questions
		existingQuestions := make(map[uint]models.MstrInspectionQuestion)
		if detail.Id != 0 {
			var oldQ []models.MstrInspectionQuestion
			config.DB.Where("inspection_detail_id = ?", detail.Id).
				Preload("Options").
				Find(&oldQ)
			for _, q := range oldQ {
				existingQuestions[q.ID] = q
			}
		}

		for _, q := range detail.Questions {
			if q.ID == 0 {
				// INSERT question baru
				q.InspectionDetailID = detail.Id
				q.CreatedBy = username.(string)
				q.UpdatedBy = username.(string)
				if err := config.DB.Create(&q).Error; err != nil {
					utils.JSONError(c, http.StatusInternalServerError, "Failed insert question: "+err.Error())
					return
				}
			} else {
				// UPDATE question lama
				if _, ok := existingQuestions[q.ID]; ok {
					q.UpdatedBy = username.(string)
					if err := config.DB.Model(&models.MstrInspectionQuestion{}).
						Where("id = ?", q.ID).
						Updates(q).Error; err != nil {
						utils.JSONError(c, http.StatusInternalServerError, "Failed update question: "+err.Error())
						return
					}
				}
				delete(existingQuestions, q.ID)
			}

			// Proses options
			existingOptions := make(map[uint]models.MstrInspectionQuestionOption)
			if q.ID != 0 {
				var oldO []models.MstrInspectionQuestionOption
				config.DB.Where("inspection_question_id = ?", q.ID).Find(&oldO)
				for _, o := range oldO {
					existingOptions[o.ID] = o
				}
			}

			for _, opt := range q.Options {
				if opt.ID == 0 {
					opt.InspectionQuestionID = q.ID
					opt.CreatedBy = username.(string)
					opt.UpdatedBy = username.(string)
					if err := config.DB.Create(&opt).Error; err != nil {
						utils.JSONError(c, http.StatusInternalServerError, "Failed insert option: "+err.Error())
						return
					}
				} else {
					if _, ok := existingOptions[opt.ID]; ok {
						opt.UpdatedBy = username.(string)
						if err := config.DB.Model(&models.MstrInspectionQuestionOption{}).
							Where("id = ?", opt.ID).
							Updates(opt).Error; err != nil {
							utils.JSONError(c, http.StatusInternalServerError, "Failed update option: "+err.Error())
							return
						}
					}
					delete(existingOptions, opt.ID)
				}
			}

			// DELETE options yang tidak ada di payload
			for oid := range existingOptions {
				config.DB.Delete(&models.MstrInspectionQuestionOption{}, oid)
			}
		}

		// DELETE questions yang tidak ada di payload
		for qid := range existingQuestions {
			config.DB.Delete(&models.MstrInspectionQuestion{}, qid)
		}
	}

	// DELETE details yang tidak ada di payload
	for did := range existingDetails {
		config.DB.Delete(&models.MstrInspectionDetail{}, did)
	}

	// Ambil hasil akhir
	var updated models.MstrInspection
	if err := config.DB.Preload("Details.Questions.Options").
		First(&updated, inspection.Id).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Mstr Inspection updated", updated)
}

// Update hanya nama inspection
func UpdateMstrInspectionByID(c *gin.Context) {
	id := c.Param("id")

	var payload struct {
		NameInspection string `json:"name_inspection"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid payload: "+err.Error())
		return
	}

	if payload.NameInspection == "" {
		utils.JSONError(c, http.StatusBadRequest, "Assurance name cannot be empty")
		return
	}

	username := c.GetString("username")

	// Update langsung tanpa preload
	result := config.DB.Model(&models.MstrInspection{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"name_inspection": payload.NameInspection,
			"updated_by":      username,
			//"updated_at":      time.Now(),
		})

	if result.Error != nil {
		utils.JSONError(c, http.StatusInternalServerError, result.Error.Error())
		return
	}

	if result.RowsAffected == 0 {
		utils.JSONError(c, http.StatusNotFound, "Assurance not found")
		return
	}

	utils.JSONSuccess(c, "Assurance name updated successfully", nil)
}

// COPY ASSURANCE MASTER
func CopyMstrInspectionByID(c *gin.Context) {
	id := c.Param("id")

	// Ambil info user
	userCompanyID, _ := c.Get("company_id")
	username, _ := c.Get("username")

	// === Ambil data inspection asli lengkap dengan relasi ===
	var original models.MstrInspection
	if err := config.DB.
		Preload("Details.Questions.Options").
		First(&original, "id = ?", id).Error; err != nil {
		utils.JSONError(c, http.StatusNotFound, "Original inspection not found")
		return
	}

	// === Mulai transaksi ===
	tx := config.DB.Begin()
	if tx.Error != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Failed to start transaction")
		return
	}

	// === Buat record baru dengan nama baru ===
	newInspection := models.MstrInspection{
		NameInspection: fmt.Sprintf("%s (Copy)", original.NameInspection),
		ImageUrl:       original.ImageUrl, // pakai image lama
		CompanyID:      userCompanyID.(string),
		CreatedBy:      username.(string),
		UpdatedBy:      username.(string),
	}

	if err := tx.Create(&newInspection).Error; err != nil {
		tx.Rollback()
		utils.JSONError(c, http.StatusInternalServerError, "Failed to copy inspection: "+err.Error())
		return
	}

	// === Salin semua detail, question, option ===
	for _, d := range original.Details {
		newDetail := models.MstrInspectionDetail{
			IdMstrInspection:   newInspection.Id,
			NameCoordinate:     d.NameCoordinate,
			X:                  d.X,
			Y:                  d.Y,
			TutorialCoordinate: d.TutorialCoordinate,
			RequiredCoordinate: d.RequiredCoordinate,
			SendNow:            d.SendNow,
			TypeTriggerID:      d.TypeTriggerID,
			CreatedBy:          username.(string),
			UpdatedBy:          username.(string),
		}

		if err := tx.Create(&newDetail).Error; err != nil {
			tx.Rollback()
			utils.JSONError(c, http.StatusInternalServerError, "Failed to copy detail: "+err.Error())
			return
		}

		for _, q := range d.Questions {
			newQ := models.MstrInspectionQuestion{
				InspectionDetailID: newDetail.Id,
				Text:               q.Text,
				Type:               q.Type,
				CreatedBy:          username.(string),
				UpdatedBy:          username.(string),
			}

			if err := tx.Create(&newQ).Error; err != nil {
				tx.Rollback()
				utils.JSONError(c, http.StatusInternalServerError, "Failed to copy question: "+err.Error())
				return
			}

			for _, o := range q.Options {
				newO := models.MstrInspectionQuestionOption{
					InspectionQuestionID: newQ.ID,
					Label:                o.Label,
					Text:                 o.Text,
					IsCorrect:            o.IsCorrect,
					CreatedBy:            username.(string),
					UpdatedBy:            username.(string),
				}

				if err := tx.Create(&newO).Error; err != nil {
					tx.Rollback()
					utils.JSONError(c, http.StatusInternalServerError, "Failed to copy option: "+err.Error())
					return
				}
			}
		}
	}

	// === Commit ===
	if err := tx.Commit().Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Failed to commit copy: "+err.Error())
		return
	}

	utils.JSONSuccess(c, "Inspection copied successfully", newInspection)
}
