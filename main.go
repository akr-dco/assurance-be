package main

import (
	"go-api/config"
	"go-api/middleware"
	"go-api/models"
	"go-api/routes"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Cek environment
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "development"
	}

	// Load .env sesuai APP_ENV
	err := godotenv.Load(".env." + appEnv)
	if err != nil {
		log.Println("No .env." + appEnv + " file found, fallback to default .env")
		godotenv.Load(".env") // fallback
	}

	config.ConnectDB()
	config.DB.AutoMigrate(
		&models.MstrCompany{},
		&models.MstrUser{},
		&models.MstrInspection{},
		&models.MstrInspectionDetail{},
		//&models.ChildInspection{},
		//&models.ChildInspectionDetail{},
		&models.MstrChaining{},
		&models.MstrChainingDetail{},
		&models.MstrEventTrigger{},
		&models.MstrTypeTrigger{},
		// &models.InspectionByUser{},
		// &models.MstrCoordinate{},
		&models.TrxInspection{},
		&models.TrxInspectionDetail{},
		&models.TrxInspectionAnswer{},
		&models.MstrDevice{},
		&models.MstrGroup{},
		&models.Questionnaire{},
		&models.Question{},
		&models.Option{},
		&models.Answer{},
		&models.MstrAnswer{},
		&models.MstrAnswerDetail{},
		&models.MstrInspectionQuestion{},
		&models.MstrInspectionQuestionOption{},
		&models.PasswordResetToken{},
	)

	r := gin.Default()

	//Aktifkan middleware CORS sebelum semua route
	r.Use(middleware.CORSMiddleware())
	routes.SetupRoutes(r)

	for _, ri := range r.Routes() {
		println(ri.Method, ri.Path)
	}

	// Jalankan server dengan port dari .env
	r.Run(":" + os.Getenv("SERVER_PORT"))

}
