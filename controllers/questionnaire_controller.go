package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go-api/config"
	"go-api/models"
	"go-api/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ---------- Questionnaire CRUD ----------

type CreateQuestionnaireReq struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Type        string `json:"type"`
	CompanyID   string `json:"company_id"`
	IsActive    bool   `json:"is_active"`
	CreatedBy   string `json:"created_by"`
}

func CreateQuestionnaire(c *gin.Context) {
	var req CreateQuestionnaireReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	q := models.Questionnaire{
		Title:       req.Title,
		Description: req.Description,
		Type:        req.Type,
		IsActive:    req.IsActive,
	}

	userCompanyID, _ := c.Get("company_id")
	q.CompanyID = userCompanyID.(string)
	username, _ := c.Get("username")
	q.CreatedBy = username.(string)
	q.UpdatedBy = username.(string)

	if err := config.DB.Create(&q).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": q})
}

func ListQuestionnaires(c *gin.Context) {
	role, _ := c.Get("role")
	userCompanyID, _ := c.Get("company_id")

	// mulai query dari DB
	query := config.DB.Preload("Questions.Options").Order("id desc")

	// filter jika bukan super-admin
	if role != "super-admin" {
		query = query.Where("company_id = ?", userCompanyID)
	}

	var list []models.Questionnaire
	if err := query.Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   list,
	})
}

func GetQuestionnaire(c *gin.Context) {
	id := c.Param("id")
	var q models.Questionnaire
	if err := config.DB.Preload("Questions.Options").First(&q, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "questionnaire not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": q})
}

type UpdateQuestionnaireReq struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Type        *string `json:"type"`
	CompanyID   *string `json:"company_id"`
	IsActive    *bool   `json:"is_active"`
	UpdatedBy   *string `json:"updated_by"`
}

func UpdateQuestionnaire(c *gin.Context) {
	id := c.Param("id")
	var body UpdateQuestionnaireReq
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	var q models.Questionnaire
	if err := config.DB.First(&q, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "questionnaire not found"})
		return
	}

	if body.Title != nil {
		q.Title = *body.Title
	}
	if body.Description != nil {
		q.Description = *body.Description
	}
	if body.Type != nil {
		q.Type = *body.Type
	}
	if body.IsActive != nil {
		q.IsActive = *body.IsActive
	}
	// Update field
	userCompanyID, _ := c.Get("company_id")
	q.CompanyID = userCompanyID.(string)

	username, _ := c.Get("username")
	q.UpdatedBy = username.(string)

	if err := config.DB.Save(&q).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": q})
}

func DeleteQuestionnaire(c *gin.Context) {
	id := c.Param("id")

	deletedBy := c.GetString("username")

	// Set DeletedBy
	if err := config.DB.Model(&models.Questionnaire{}).
		Where("id = ?", id).
		Update("deleted_by", deletedBy).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	if err := config.DB.Delete(&models.Questionnaire{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ---------- Question CRUD ----------

type OptionReq struct {
	Label     string `json:"label"`      // A..E
	Text      string `json:"text"`       // option text
	IsCorrect bool   `json:"is_correct"` // optional
}

type CreateQuestionReq struct {
	Text    string      `json:"text" binding:"required"`
	Type    string      `json:"type" binding:"required"` // yesno|multiple|essay|image
	Options []OptionReq `json:"options"`                 // only for multiple
}

func CreateQuestion(c *gin.Context) {
	qnID := c.Param("questionnaireId")

	var req CreateQuestionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	qt := strings.ToLower(req.Type)
	if qt != "yesno" && qt != "multiple" && qt != "text" && qt != "number" && qt != "image" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid question type"})
		return
	}

	newQ := models.Question{
		QuestionnaireID: parseUint(qnID),
		Text:            req.Text,
		Type:            qt,
	}

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&newQ).Error; err != nil {
			return err
		}

		// For multiple choice, save options (A..E)
		if qt == "multiple" {
			if len(req.Options) < 2 {
				return errors.New("multiple choice requires at least 2 options")
			}
			opts := make([]models.Option, 0, len(req.Options))
			for _, o := range req.Options {
				lbl := strings.ToUpper(strings.TrimSpace(o.Label))
				if lbl == "" {
					return errors.New("option label is required")
				}
				opts = append(opts, models.Option{
					QuestionID: newQ.ID,
					Label:      lbl,
					Text:       o.Text,
					IsCorrect:  o.IsCorrect,
				})
			}
			if err := tx.Create(&opts).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	config.DB.Preload("Options").First(&newQ, newQ.ID)
	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": newQ})
}

type UpdateQuestionReq struct {
	Text    *string      `json:"text"`
	Type    *string      `json:"type"`    // if changed to multiple, must also send options
	Options *[]OptionReq `json:"options"` // replace all options
}

func UpdateQuestion(c *gin.Context) {
	id := c.Param("id")
	var body UpdateQuestionReq
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	var q models.Question
	if err := config.DB.Preload("Options").First(&q, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "question not found"})
		return
	}

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		if body.Text != nil {
			q.Text = *body.Text
		}
		if body.Type != nil {
			t := strings.ToLower(*body.Type)
			if t != "yesno" && t != "multiple" && t != "text" && t != "number" && t != "image" {
				return errors.New("invalid question type")
			}
			q.Type = t
		}
		if err := tx.Save(&q).Error; err != nil {
			return err
		}

		// Replace options if provided (for multiple)
		if body.Options != nil {
			if q.Type != "multiple" {
				// if type not multiple, options should be cleared
				if err := tx.Where("question_id = ?", q.ID).Delete(&models.Option{}).Error; err != nil {
					return err
				}
			} else {
				// delete old
				if err := tx.Where("question_id = ?", q.ID).Delete(&models.Option{}).Error; err != nil {
					return err
				}
				// insert new
				opts := make([]models.Option, 0, len(*body.Options))
				for _, o := range *body.Options {
					opts = append(opts, models.Option{
						QuestionID: q.ID,
						Label:      strings.ToUpper(o.Label),
						Text:       o.Text,
						IsCorrect:  o.IsCorrect,
					})
				}
				if len(opts) < 2 {
					return errors.New("multiple choice requires at least 2 options")
				}
				if err := tx.Create(&opts).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	config.DB.Preload("Options").First(&q, q.ID)
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": q})
}

func DeleteQuestion(c *gin.Context) {
	id := c.Param("id")

	deletedBy := c.GetString("username")
	// Set DeletedBy
	if err := config.DB.Model(&models.Question{}).
		Where("id = ?", id).
		Update("deleted_by", deletedBy).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	if err := config.DB.Delete(&models.Question{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ---------- Submit Answer (JSON / multipart) ----------

type AnswerJSONReq struct {
	UserID uint `json:"user_id" binding:"required"`
	// Isi salah satu sesuai tipe:
	AnswerText string `json:"answer_text"` // yesno/multiple/essay
	// Untuk image gunakan multipart (file)
}

func SubmitAnswer(c *gin.Context) {
	questionID := parseUint(c.Param("id"))

	var question models.Question
	if err := config.DB.Preload("Options").First(&question, questionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "question not found"})
		return
	}

	ct := c.GetHeader("Content-Type")
	var ans models.Answer

	if strings.HasPrefix(ct, "multipart/form-data") {
		// image upload route
		userIDStr := c.PostForm("user_id")
		if userIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "user_id is required"})
			return
		}
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "file is required"})
			return
		}
		// only valid for image type
		if question.Type != "image" {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "this question does not accept image"})
			return
		}
		path, err := saveUploadedFile(c, file, "uploads")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
			return
		}
		ans = models.Answer{
			QuestionID: question.ID,
			UserID:     parseUint(userIDStr),
			AnswerFile: path,
		}
	} else {
		// JSON
		var body AnswerJSONReq
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
			return
		}
		if body.UserID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "user_id is required"})
			return
		}

		normalized := strings.TrimSpace(body.AnswerText)
		switch question.Type {
		case "yesno":
			up := strings.ToUpper(normalized)
			if up != "YES" && up != "NO" {
				c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "answer must be Yes or No"})
				return
			}
		case "multiple":
			up := strings.ToUpper(normalized)
			if up == "" {
				c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "answer must be option label (A..E)"})
				return
			}
			valid := false
			for _, o := range question.Options {
				if strings.ToUpper(o.Label) == up {
					valid = true
					break
				}
			}
			if !valid {
				c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid option label"})
				return
			}
		case "essay":
			if normalized == "" {
				c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "essay cannot be empty"})
				return
			}
		case "image":
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "image answer must be multipart/form-data with file"})
			return
		}

		ans = models.Answer{
			QuestionID: question.ID,
			UserID:     body.UserID,
			AnswerText: normalized,
		}
	}

	if err := config.DB.Create(&ans).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": ans})
}

func ListAnswers(c *gin.Context) {
	qid := parseUint(c.Param("id"))
	var list []models.Answer
	if err := config.DB.Where("question_id = ?", qid).Order("id desc").Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": list})
}

// ---------- Helpers ----------

func parseUint(s string) uint {
	var id uint
	fmt.Sscan(s, &id)
	return id
}

func saveUploadedFile(c *gin.Context, fileHeader *multipart.FileHeader, baseDir string) (string, error) {
	_ = os.MkdirAll(baseDir, 0755)
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext == "" {
		ext = ".bin"
	}
	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), sanitize(fileHeader.Filename))
	fullpath := filepath.Join(baseDir, filename)

	if err := saveFile(c, fileHeader, fullpath); err != nil {
		return "", err
	}
	return fullpath, nil
}

func saveFile(c *gin.Context, fh *multipart.FileHeader, dst string) error {
	return c.SaveUploadedFile(fh, dst)
}

func sanitize(s string) string {
	s = strings.ReplaceAll(s, " ", "_")
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '_' || r == '.' || r == '-' {
			return r
		}
		return -1
	}, s)
}

// ANSWER
type UserAnswerFlat struct {
	QuestionID   uint   `json:"question_id"`
	QuestionText string `json:"question_text"`
	QuestionType string `json:"question_type"`
	OptionLabel  string `json:"option_label,omitempty"`
	OptionText   string `json:"option_text,omitempty"`
	AnswerID     uint   `json:"answer_id,omitempty"`
	AnswerText   string `json:"answer_text,omitempty"`
	AnswerFile   string `json:"answer_file,omitempty"`
}

func GetUserAnswersFlat(c *gin.Context) {
	questionnaireID := parseUint(c.Param("id"))
	userID := parseUint(c.Param("userId"))

	var results []UserAnswerFlat

	err := config.DB.Table("questions").
		Select(`
            questions.id as question_id,
            questions.text as question_text,
            questions.type as question_type,
            options.label as option_label,
            options.text as option_text,
            answers.id as answer_id,
            answers.answer_text,
            answers.answer_file
        `).
		Joins("LEFT JOIN options ON options.question_id = questions.id").
		Joins("LEFT JOIN answers ON answers.question_id = questions.id AND answers.user_id = ?", userID).
		Where("questions.questionnaire_id = ?", questionnaireID).
		Order("questions.id, options.label").
		Scan(&results).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   results,
	})
}

func GetUserAnswers(c *gin.Context) {
	questionnaireID := parseUint(c.Param("id"))
	userID := parseUint(c.Param("userId"))

	var questionnaire models.Questionnaire
	err := config.DB.
		Preload("Questions.Options").                        // preload pertanyaan + opsi
		Preload("Questions.Answers", "user_id = ?", userID). // preload jawaban user tertentu
		First(&questionnaire, questionnaireID).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "questionnaire not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   questionnaire,
	})
}

func SubmitAllAnswers(c *gin.Context) {
	answersJSON := c.PostForm("answers")
	userIDStr := c.PostForm("user_id")
	if userIDStr == "" || answersJSON == "" {
		utils.JSONError(c, http.StatusBadRequest, "user_id and answers are required")
		return
	}

	userID := parseUint(userIDStr)

	type AnswerPayload struct {
		QuestionID uint   `json:"question_id"`
		AnswerText string `json:"answer_text"`
		AnswerFile string `json:"answer_file"` // nama field filenya (ex: "file_0")
	}

	var payloads []AnswerPayload
	if err := json.Unmarshal([]byte(answersJSON), &payloads); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid JSON in answers: "+err.Error())
		return
	}

	var allAnswers []models.Answer

	for _, p := range payloads {
		var question models.Question
		if err := config.DB.Preload("Options").First(&question, p.QuestionID).Error; err != nil {
			utils.JSONError(c, http.StatusBadRequest, fmt.Sprintf("Question %d not found", p.QuestionID))
			return
		}

		var answer models.Answer
		answer.QuestionID = p.QuestionID
		answer.UserID = userID

		if question.Type == "image" {
			if p.AnswerFile == "" {
				utils.JSONError(c, http.StatusBadRequest, fmt.Sprintf("answer_file required for image question %d", p.QuestionID))
				return
			}

			fileHeader, err := c.FormFile(p.AnswerFile)
			if err != nil {
				utils.JSONError(c, http.StatusBadRequest, fmt.Sprintf("File %s missing: %v", p.AnswerFile, err))
				return
			}

			file, err := fileHeader.Open()
			if err != nil {
				utils.JSONError(c, http.StatusInternalServerError, fmt.Sprintf("Cannot open file %s: %v", p.AnswerFile, err))
				return
			}
			defer file.Close()

			fileKey := GenerateE2ObjectKey(c, "Trn-Questionnaire", fileHeader.Filename)
			objectKey, err := UploadFileToE2(c, file, fileKey, fileHeader.Header.Get("Content-Type"), "Assurance", nil)
			if err != nil {
				utils.JSONError(c, http.StatusInternalServerError, "Failed upload: "+err.Error())
				return
			}

			answer.AnswerFile = objectKey
		} else {
			normalized := strings.TrimSpace(p.AnswerText)
			if question.Type == "essay" && normalized == "" {
				utils.JSONError(c, http.StatusBadRequest, fmt.Sprintf("Essay cannot be empty for question %d", p.QuestionID))
				return
			}
			answer.AnswerText = normalized
		}

		allAnswers = append(allAnswers, answer)
	}

	if err := config.DB.Create(&allAnswers).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "All answers submitted", allAnswers)
}

func SubmitAllAnswersWithMaster(c *gin.Context) {
	answersJSON := c.PostForm("answers")
	userIDStr := c.PostForm("user_id")
	questionnaireIDStr := c.PostForm("questionnaire_id")
	deviceID := c.PostForm("device_id")
	chaningIDStr := c.PostForm("chaining_id")
	userCompanyID := c.GetString("company_id")
	username := c.GetString("username")

	if userIDStr == "" || answersJSON == "" || questionnaireIDStr == "" {
		utils.JSONError(c, http.StatusBadRequest, "user_id, questionnaire_id, and answers are required")
		return
	}

	userID := parseUint(userIDStr)
	questionnaireID := parseUint(questionnaireIDStr)
	chaningID := parseUint(chaningIDStr)

	type AnswerPayload struct {
		QuestionID uint   `json:"question_id"`
		AnswerText string `json:"answer_text"`
		AnswerFile string `json:"answer_file"` // key file
	}

	var payloads []AnswerPayload
	if err := json.Unmarshal([]byte(answersJSON), &payloads); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	// Gunakan transaksi
	tx := config.DB.Begin()

	master := models.MstrAnswer{
		QuestionnaireID: questionnaireID,
		UserID:          userID,
		CompanyID:       userCompanyID,
		DeviceID:        deviceID,
		ChainingID:      chaningID,
		CreatedBy:       username,
		UpdatedBy:       username,
	}
	if err := tx.Create(&master).Error; err != nil {
		tx.Rollback()
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	var details []models.MstrAnswerDetail

	for _, p := range payloads {
		var question models.Question
		if err := tx.First(&question, p.QuestionID).Error; err != nil {
			tx.Rollback()
			utils.JSONError(c, http.StatusBadRequest, fmt.Sprintf("Question %d not found", p.QuestionID))
			return
		}

		detail := models.MstrAnswerDetail{
			MasterAnswerID: master.ID,
			QuestionID:     p.QuestionID,
			CreatedBy:      username,
			UpdatedBy:      username,
		}

		if question.Type == "image" {
			fileHeader, err := c.FormFile(p.AnswerFile)
			if err != nil {
				tx.Rollback()
				utils.JSONError(c, http.StatusBadRequest, fmt.Sprintf("File missing: %v", err))
				return
			}
			file, _ := fileHeader.Open()
			defer file.Close()

			fileKey := GenerateE2ObjectKey(c, "Trn-Questionnaire", fileHeader.Filename)
			objectKey, err := UploadFileToE2(c, file, fileKey, fileHeader.Header.Get("Content-Type"), "Assurance/"+userCompanyID, nil)
			if err != nil {
				tx.Rollback()
				utils.JSONError(c, http.StatusInternalServerError, "Failed upload: "+err.Error())
				return
			}

			detail.AnswerFile = objectKey
		} else {
			detail.AnswerText = strings.TrimSpace(p.AnswerText)
		}

		details = append(details, detail)
	}

	if err := tx.Create(&details).Error; err != nil {
		tx.Rollback()
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	tx.Commit()
	utils.JSONSuccess(c, "All answers submitted", master)
}
