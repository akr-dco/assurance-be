package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type JSONResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// JSONSuccess response success standar
func JSONSuccess(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, JSONResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

// JSONCreated response sukses untuk create
func JSONCreated(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, JSONResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

// JSONError response error dengan status custom
func JSONError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, JSONResponse{
		Status:  "error",
		Message: message,
		Data:    nil,
	})
}

// JSONValidationError khusus untuk validasi input
func JSONValidationError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{
		"status":  "error",
		"message": "Validation failed",
		"error":   err.Error(),
	})
}
