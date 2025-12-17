package controllers

import (
	"errors"
	"fmt"
	"go-api/config"
	"go-api/models"
	"go-api/utils"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func GetFilteredUsers(c *gin.Context) {

	role := c.GetString("role")
	userCompanyID := c.GetString("company_id")
	//username := c.GetString("username")

	// Filter dari query param
	filters := map[string]string{
		"created_by": c.Query("created_by"),
		"updated_by": c.Query("updated_by"),
	}

	var users []models.MstrUser
	query := config.DB.WithContext(c.Request.Context()).Model(&models.MstrUser{})

	// Restriksi company kalau bukan super-admin
	/*
		if role != "super-admin" && userCompanyID != "" {
			query = query.Where("company_id = ? AND role = 'user' OR username = ?", userCompanyID, username)
		}
	*/

	if role != "super-admin" && userCompanyID != "" {
		query = query.Where("company_id = ?", userCompanyID)
	}

	// Tambahkan filter dinamis
	for field, value := range filters {
		if value != "" {
			query = query.Where(fmt.Sprintf("%s = ?", field), value)
		}
	}

	// Optional: Pagination
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if err := query.Order("id DESC").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Failed to fetch users")
		return
	}

	utils.JSONSuccess(c, "Filtered users", users)
}

func CreateUser(c *gin.Context) {
	// Ambil data dari context dengan type assertion

	username := c.GetString("username")
	role := c.GetString("role")
	userCompanyID := c.GetString("company_id")

	// Bind request body
	var input models.CreateUserRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid input")
		return
	}

	// Validasi email & username unik
	var existing models.MstrUser
	if err := config.DB.WithContext(c.Request.Context()).
		Where("username = ? OR email = ?", input.Username, input.Email).
		First(&existing).Error; err == nil {
		utils.JSONError(c, http.StatusBadRequest, "Username or Email already exists")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// Tentukan CompanyID berdasarkan role
	companyCode := userCompanyID
	if role == "super-admin" && input.CompanyID != "" {
		companyCode = input.CompanyID
	}

	//now := time.Now()
	user := models.MstrUser{
		Username:  input.Username,
		Password:  string(hashedPassword),
		FullName:  input.FullName,
		Email:     input.Email,
		Phone:     input.Phone,
		Role:      input.Role,
		IsActive:  input.IsActive,
		CreatedBy: username,
		UpdatedBy: username,
		CompanyID: companyCode,
		//CreatedAt: now,
		//UpdatedAt: now,
	}

	// Insert user ke database dengan context-aware
	if err := config.DB.WithContext(c.Request.Context()).Create(&user).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Failed to create user")
		return
	}

	utils.JSONSuccess(c, "User created", user)
}

func UpdateUserByID(c *gin.Context) {
	id := c.Param("id")
	username := c.GetString("username")
	//updatedBy, _ := usernameCtx.(string)

	// Cari user by ID
	var user models.MstrUser
	if err := config.DB.WithContext(c.Request.Context()).First(&user, id).Error; err != nil {
		utils.JSONError(c, http.StatusNotFound, "User not found")
		return
	}

	// Bind request body
	var input models.UpdateUserRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid input")
		return
	}

	// Cek duplikasi Username & Email (dalam satu blok)
	dupChecks := []struct {
		field string
		value string
		msg   string
	}{
		{"username", input.Username, "Username already exists"},
		{"email", input.Email, "Email already exists"},
	}

	for _, check := range dupChecks {
		if check.value != "" {
			var count int64
			config.DB.WithContext(c.Request.Context()).
				Model(&models.MstrUser{}).
				Where(fmt.Sprintf("%s = ? AND id <> ?", check.field), check.value, id).
				Count(&count)
			if count > 0 {
				utils.JSONError(c, http.StatusBadRequest, check.msg)
				return
			}
		}
	}

	// Update field
	user.Username = input.Username
	user.FullName = input.FullName
	user.Email = input.Email
	user.Phone = input.Phone
	user.Role = input.Role
	user.IsActive = input.IsActive
	//user.UpdatedAt = time.Now()
	user.UpdatedBy = username

	// Update password jika ada
	if input.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			utils.JSONError(c, http.StatusInternalServerError, "Failed to hash password")
			return
		}
		user.Password = string(hashedPassword)
	}

	// Simpan perubahan
	if err := config.DB.WithContext(c.Request.Context()).Save(&user).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Failed to update user")
		//log.Printf("DB error: %v", err)
		return
	}

	utils.JSONSuccess(c, "User updated", user)
}

func DeleteUserByID(c *gin.Context) {
	id := c.Param("id")
	username := c.GetString("username")

	// Cari user yang akan dihapus
	var user models.MstrUser
	if err := config.DB.WithContext(c.Request.Context()).First(&user, id).Error; err != nil {
		utils.JSONError(c, http.StatusNotFound, "User not found")
		return
	}

	// Set kolom deleted_by dan simpan sebelum delete
	if err := config.DB.WithContext(c.Request.Context()).
		Model(&user).
		Update("deleted_by", username).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Failed to set deleted_by")
		return
	}

	/*
		result := config.DB.WithContext(c.Request.Context()).Delete(&models.MstrUser{}, id)
		if result.Error != nil {
			utils.JSONError(c, http.StatusInternalServerError, "Failed to delete user")
			return
		}
	*/

	if err := config.DB.WithContext(c.Request.Context()).
		Delete(&user).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	utils.JSONSuccess(c, "User deleted", nil)
}

func Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		DeviceID string `json:"device_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid input")
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))

	// Ambil user dari database
	var user models.MstrUser
	if err := config.DB.WithContext(c.Request.Context()).
		Where("LOWER(email) = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.JSONError(c, http.StatusUnauthorized, "Invalid credentials")
			return
		}
		utils.JSONError(c, http.StatusInternalServerError, "Failed to fetch user")
		return
	}

	// Cek password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		utils.JSONError(c, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Cek role admin/super-admin — tidak boleh login via mobile app (pakai device_id)
	if (user.Role == "admin" || user.Role == "super-admin") && req.DeviceID != "" {
		utils.JSONError(c, http.StatusBadRequest, "This role is not allowed to login to mobile app")
		return
	}

	// Kalau bukan super-admin
	var company models.MstrCompany
	if user.Role != "super-admin" {

		//Cek apakah company ada
		if err := config.DB.WithContext(c.Request.Context()).
			Where("company_id = ?", user.CompanyID).First(&company).Error; err != nil {
			utils.JSONError(c, http.StatusNotFound, "Company not found")
			return
		}

		//Cek apakah company active
		if !company.IsActive {
			utils.JSONError(c, http.StatusForbidden, "Company is not active")
			return
		}

		// Cek apakah user aktif
		if !user.IsActive {
			utils.JSONError(c, http.StatusForbidden, "User is not active")
			return
		}

		// Kalau role user
		if user.Role == "user" {

			//cek apakah device_id ada
			if req.DeviceID == "" {
				utils.JSONError(c, http.StatusBadRequest, "Device ID is required")
				return
			}

			var device models.MstrDevice
			//Cek device_id dan company_id dari user
			if err := config.DB.WithContext(c.Request.Context()).
				Where("device_id = ? AND company_id = ?", req.DeviceID, user.CompanyID).First(&device).Error; err != nil {
				utils.JSONError(c, http.StatusNotFound, "Device not found")
				return
			}

			//Cek apakah device user active
			if !device.IsActive {
				utils.JSONError(c, http.StatusForbidden, "Device is not active")
				return
			}
		}

	}

	// Generate tokens
	accessToken, expiresAt, err := utils.GenerateAccessToken(
		user.Id, user.Username, user.Email, user.Role, user.CompanyID,
		30*24*time.Hour)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Failed to generate access token")
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(user.Id, 7*24*time.Hour)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Failed to generate refresh token")
		return
	}

	response := gin.H{
		"user_id":       user.Id,
		"username":      user.Username,
		"fullname":      user.FullName,
		"email":         user.Email,
		"company_id":    user.CompanyID,
		"company_name":  company.CompanyName,
		"company_logo":  company.ImageUrl,
		"role":          user.Role,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_at":    expiresAt,
	}

	fmt.Println("[DEBUG] Login successful:", response) // debug

	utils.JSONSuccess(c, "Login successful", response)
}

func Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid input")
		return
	}

	claims, err := utils.ParseToken(req.RefreshToken)
	if err != nil {
		utils.JSONError(c, http.StatusUnauthorized, "Invalid refresh token")
		return
	}

	if claims["scope"] != "refresh" {
		utils.JSONError(c, http.StatusForbidden, "Invalid token scope")
		return
	}

	// Ambil user_id dari refresh token
	userIDFloat, ok := claims["sub"].(float64)
	if !ok {
		utils.JSONError(c, http.StatusUnauthorized, "Invalid subject")
		return
	}
	userID := uint(userIDFloat)

	// Cari user dari DB
	var user models.MstrUser
	if err := config.DB.First(&user, userID).Error; err != nil {
		utils.JSONError(c, http.StatusUnauthorized, "User not found")
		return
	}

	// Buat access token baru
	accessToken, expiresAt, err := utils.GenerateAccessToken(user.Id, user.Username, user.Email, user.Role, user.CompanyID, 15*time.Minute)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Failed to generate new access token")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
		"expires_at":   expiresAt,
	})
}

func ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid input")
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))

	// cari user berdasarkan email
	var user models.MstrUser
	if err := config.DB.Where("LOWER(email) = ?", email).First(&user).Error; err != nil {
		utils.JSONSuccess(c, "If the email exists, reset link has been sent.", nil)
		return
	}

	// generate reset token
	token := utils.GenerateRandomString(40)
	expiresAt := time.Now().Add(60 * time.Minute)

	reset := models.PasswordResetToken{
		UserID:    user.Id,
		Token:     token,
		ExpiresAt: expiresAt,
	}

	config.DB.Create(&reset)

	// buat link reset
	frontendURL := os.Getenv("SERVER_FRONTEND")
	resetLink := fmt.Sprintf("%s/#/reset-password/%s", frontendURL, token)

	// kirim email
	/*
		go utils.SendEmail(user.Email,
			"Reset Password Request",
			fmt.Sprintf("Click link to reset your password: %s", resetLink))
	*/

	go utils.SendEmail(
		user.Email,
		"Reset Password Request",
		utils.BuildResetPasswordEmail(user.FullName, resetLink),
	)

	//fmt.Println("Sending email to:", user.Email)
	//fmt.Println("Reset link:", resetLink)

	utils.JSONSuccess(c, "Reset password link sent to your email.", nil)
}

func ResetPassword(c *gin.Context) {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid input")
		return
	}

	var reset models.PasswordResetToken
	if err := config.DB.
		Where("token = ? AND used = FALSE AND expires_at > NOW()", req.Token).
		First(&reset).Error; err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid or expired token")
		return
	}

	var user models.MstrUser
	config.DB.First(&user, reset.UserID)

	hashed, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	user.Password = string(hashed)
	config.DB.Save(&user)

	reset.Used = true
	config.DB.Save(&reset)

	utils.JSONSuccess(c, "Password successfully updated", nil)
}
