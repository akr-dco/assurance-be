package routes

import (
	"go-api/controllers"
	"go-api/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
	// Group untuk login & refresh (tidak pakai JWT / CheckCompanyActive)
	public := router.Group("/api")
	{
		public.POST("/login", controllers.Login)
		public.POST("/refresh", controllers.Refresh) // opsional kalau pakai refresh token
		public.POST("/mstr-device", controllers.CreateDevice)
		public.POST("/forgot-password", controllers.ForgotPassword)
		public.POST("/reset-password", controllers.ResetPassword)
		public.GET("/e2-signed-company/:companyID/*objectKey", controllers.GetSignedFileURLWithCompany)

	}

	api := router.Group("/api", middleware.APIKeyAuth(), middleware.AuthMiddleware(), middleware.CheckCompanyActive())
	{

		//User
		api.POST("/users", controllers.CreateUser)             //user-mstr
		api.PUT("/users/:id", controllers.UpdateUserByID)      //user-mstr
		api.DELETE("/users/:id", controllers.DeleteUserByID)   //user-mstr
		api.GET("/users/filter", controllers.GetFilteredUsers) //user-mstr

		//MSTR Inspection
		api.POST("/mstr-inspections", controllers.CreateMstrInspection) //assurance-master
		//api.PUT("/mstr-inspections/:id", controllers.UpdateMstrInspectionByID)
		api.PUT("/mstr-inspections/:id/name", controllers.UpdateMstrInspectionByID) //assurance-master/only update assurance_name
		api.DELETE("/mstr-inspections/:id", controllers.DeleteMstrInspectionByID)   //assurance-master
		api.GET("/mstr-inspections/filter", controllers.GetFilteredInspections)     //assurance-master
		api.POST("/mstr-inspections/:id/copy", controllers.CopyMstrInspectionByID)  //assurance-master

		api.DELETE("/mstr-inspection-details/:id", controllers.DeleteMstrInspectionDetailByID) //assurance-master//delete-sam
		api.PUT("/mstr-inspection-details/:id", controllers.UpdateMstrInspectionDetailByID)    //assurance-master//update-sam
		api.PATCH("/mstr-inspection-position-move/:id", controllers.UpdateInspectionPosition)  //assurance-master//move-sam
		api.POST("/mstr-inspection-details", controllers.CreateMstrInspectionDetail)           //assurance-master//copy-sam

		//TRX Inspection
		api.POST("/trx-inspections", controllers.CreateTRXInspection)
		api.PUT("/trx-inspections/:id", controllers.UpdateTRXInspectionByID)
		api.DELETE("/trx-inspections/:id", controllers.DeleteTRXInspectionByID)
		api.GET("/trx-inspections/filter", controllers.GetFilteredTRXInspections)

		//MSTR COMPANY
		api.POST("/mstr-company", controllers.CreateCompany)
		api.PUT("/mstr-company/:id", controllers.UpdateCompany)
		api.DELETE("/mstr-company/:id", controllers.DeleteCompany)
		api.GET("/mstr-company/filter", controllers.GetFilteredCompanies)

		//MSTR Device
		api.PUT("/mstr-device/:id", controllers.UpdateDeviceByID)
		api.DELETE("/mstr-device/:id", controllers.DeleteDeviceByID)
		api.GET("/mstr-device/filter", controllers.GetFilteredDevices)

		//MSTR Group
		api.POST("/mstr-group", controllers.CreateGroup)
		api.PUT("/mstr-group/:id", controllers.UpdateGroupByID)
		api.DELETE("/mstr-group/:id", controllers.DeleteGroupByID)
		api.GET("/mstr-group/filter", controllers.GetFilteredGroups)

		//ASSIGN DEVICE TO GROUP
		api.POST("/groups/:groupId/devices/bulk", controllers.ManageGroupDeviceBulk)
		api.GET("/groups/:id/devices", controllers.GetGroupDevices)

		//ASSIGN CHAINING TO GROUPE
		api.POST("/groups/:groupId/chainings/bulk", controllers.ManageGroupChainingBulk)
		api.GET("/groups/:id/chainings", controllers.GetGroupChainings)

		//GROUP INSPECTION
		api.POST("/groups/:groupId/inspections/bulk", controllers.ManageGroupInspectionBulk)
		api.GET("/groups/:id/inspections", controllers.GetGroupInspections)

		//GROUP Questionnaire
		api.POST("/groups/:groupId/questionnaires/bulk", controllers.ManageGroupQuestionnaireBulk)
		api.GET("/groups/:id/questionnaires", controllers.GetGroupQuestionnaires)

		// Questionnaire CRUD
		api.POST("/questionnaires", controllers.CreateQuestionnaire)
		api.GET("/questionnaires", controllers.ListQuestionnaires)
		api.GET("/questionnaires/:id", controllers.GetQuestionnaire)
		api.PUT("/questionnaires/:id", controllers.UpdateQuestionnaire)
		api.DELETE("/questionnaires/:id", controllers.DeleteQuestionnaire)

		// Question CRUD
		api.POST("/questionnaires/:questionnaireId/questions", controllers.CreateQuestion)
		api.PUT("/questions/:id", controllers.UpdateQuestion)
		api.DELETE("/questions/:id", controllers.DeleteQuestion)

		// Answers
		api.POST("/questions/:id/answers", controllers.SubmitAnswer)
		api.GET("/questions/:id/answers", controllers.ListAnswers)
		//api.POST("/questions/answers-all", controllers.SubmitAllAnswers)
		api.POST("/questions/answers-all", controllers.SubmitAllAnswersWithMaster)

		api.GET("/questionnaires/:id/users/:userId/answers", controllers.GetUserAnswers)
		api.GET("/questionnaires/:id/users/:userId/answersflat", controllers.GetUserAnswersFlat)

		// SUPERSET
		api.GET("/superset/guest-token", controllers.GetSupersetGuestToken)

		//FOR TABLET
		api.GET("/devices/:deviceID/inspections", controllers.GetDeviceInspection)
		api.GET("/devices/:deviceID/preinspections", controllers.GetDevicePreInspection)
		api.GET("/devices/:deviceID/postinspections", controllers.GetDevicePostInspection)
		api.GET("/devices/:deviceID/chaining", controllers.GetChainingByDevice)
		api.GET("/devices/:deviceID/chainingnew", controllers.GetChainingByDeviceNew)

		//CHAINING
		api.GET("/chainings/filter", controllers.GetFilteredChainings)
		api.GET("/chainings/:id", controllers.GetChainingByID)
		api.POST("/chainings/", controllers.CreateChaining)
		api.PUT("/chainings/:id", controllers.UpdateChainingByID)
		api.DELETE("/chainings/:id", controllers.DeleteChainingByID)

		//EVENT
		api.GET("/events/filter", controllers.GetFilteredEvents)
		//api.GET("/events/:id", controllers.GetEventByID)
		api.POST("/events/", controllers.CreateEvent)
		api.PUT("/events/:id", controllers.UpdateEventByID)
		api.DELETE("/events/:id", controllers.DeleteEventByID)

		//Type
		api.GET("/types/filter", controllers.GetFilteredTypes)
		api.POST("/types/", controllers.CreateType)
		api.PUT("/types/:id", controllers.UpdateTypeByID)
		api.DELETE("/types/:id", controllers.DeleteTypeByID)

		//E2 IDrive
		api.GET("/e2-signed/*objectKey", controllers.GetSignedFileURL)

	}
}
