package controllers

import (
	"bytes"
	"fmt"
	"go-api/config"
	"go-api/models"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type E2Config struct {
	Endpoint   string
	Region     string
	BucketName string
	AccessKey  string
	SecretKey  string
}

func GetCompanyE2Config(c *gin.Context) (*models.MstrCompany, error) {
	companyID := c.GetString("company_id")
	if companyID == "" {
		return nil, fmt.Errorf("company_id missing from token context")
	}

	var company models.MstrCompany
	err := config.DB.WithContext(c.Request.Context()).
		Where("company_id = ?", companyID).
		First(&company).Error

	if err != nil {
		return nil, fmt.Errorf("company not found")
	}

	fmt.Println("[DEBUG] E2 Company Config |",
		"CompanyID="+company.CompanyID,
		"Name="+company.CompanyName,
		"Endpoint="+company.E2Endpoint,
		"Region="+company.E2Region,
		"Bucket="+company.E2BucketName,
		"AccessKey="+company.E2AccessKey,
		"SecretKey="+company.E2SecretKey,
	)

	return &company, nil
}

// UploadFileToE2 uploads the file to E2 private bucket and returns the object key
func UploadFileToE2(c *gin.Context, file multipart.File, fileName string, contentType string, folder string, override *E2Config) (string, error) {

	var cfg E2Config

	// Jika super-admin memberikan config manual → pakai config itu
	if override != nil {
		cfg = *override
	} else {
		// Default: ambil dari company_id
		company, err := GetCompanyE2Config(c)
		if err != nil {
			return "", err
		}

		cfg = E2Config{
			Endpoint:   company.E2Endpoint,
			Region:     company.E2Region,
			BucketName: company.E2BucketName,
			AccessKey:  company.E2AccessKey,
			SecretKey:  company.E2SecretKey,
		}
	}

	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(cfg.Region),
		Credentials:      credentials.NewStaticCredentials(cfg.AccessKey, cfg.SecretKey, ""),
		Endpoint:         aws.String(cfg.Endpoint),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create AWS session: %v", err)
	}

	s3Client := s3.New(sess)

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(file); err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	objectKey := fmt.Sprintf("%s/%s", folder, fileName)

	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(cfg.BucketName),
		Key:         aws.String(objectKey),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String(contentType),
		ACL:         aws.String("private"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to E2: %v", err)
	}

	return objectKey, nil
}

// GeneratePresignedURL generates a signed URL valid for 15 minutes
func GeneratePresignedURL(c *gin.Context, objectKey string) (string, error) {
	// Get credential per-company
	company, err := GetCompanyE2Config(c)
	if err != nil {
		return "", err
	}

	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(company.E2Region),
		Credentials:      credentials.NewStaticCredentials(company.E2AccessKey, company.E2SecretKey, ""),
		Endpoint:         aws.String(company.E2Endpoint),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		return "", err
	}

	s3Client := s3.New(sess)

	req, _ := s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(company.E2BucketName),
		Key:    aws.String(objectKey),
	})

	presignedURL, err := req.Presign(15 * time.Minute) // URL berlaku 15 menit
	if err != nil {
		return "", err
	}

	return presignedURL, nil
}

// Endpoint untuk Superset / frontend mengakses E2 IDrive
func GetSignedFileURL(c *gin.Context) {
	//fmt.Println("[DEBUG] company_id from context:", c.GetString("company_id"))
	objectKey := c.Param("objectKey")
	objectKey = strings.TrimPrefix(objectKey, "/")

	if objectKey == "" {
		c.String(http.StatusBadRequest, "objectKey is required")
		return
	}

	// Generate signed URL
	signedURL, err := GeneratePresignedURL(c, objectKey)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to generate signed URL: %s", err.Error())
		return
	}

	// Get file from S3 / E2
	resp, err := http.Get(signedURL)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to fetch file from E2: %s", err.Error())
		return
	}
	defer resp.Body.Close()

	// Forward original content type (image/png, image/jpeg, etc)
	c.Header("Content-Type", resp.Header.Get("Content-Type"))
	c.Header("Cache-Control", "no-cache")

	// Stream the file bytes
	io.Copy(c.Writer, resp.Body)
}

func GenerateE2ObjectKey(c *gin.Context, module, originalFilename string) string {
	modulePrefix := strings.ToLower(module)

	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	day := now.Format("02")
	timestamp := now.Unix()

	uniqueID := uuid.New().String()

	// Ambil extension asli (.jpg, .png, .pdf, dll)
	ext := filepath.Ext(originalFilename)

	return fmt.Sprintf(
		"%s/%s/%s/%s/%d-%s%s",
		modulePrefix, // folder
		year,
		month,
		day,
		timestamp,
		uniqueID,
		ext, // tambahkan extension
	)
}

func GeneratePresignedURLWithCompanyID(c *gin.Context, companyID, objectKey string) (string, error) {
	// Ambil credential per-company
	var company models.MstrCompany
	err := config.DB.WithContext(c.Request.Context()).
		Where("company_id = ?", companyID).
		First(&company).Error
	if err != nil {
		return "", fmt.Errorf("company not found")
	}

	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(company.E2Region),
		Credentials:      credentials.NewStaticCredentials(company.E2AccessKey, company.E2SecretKey, ""),
		Endpoint:         aws.String(company.E2Endpoint),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		return "", err
	}

	s3Client := s3.New(sess)
	req, _ := s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(company.E2BucketName),
		Key:    aws.String(objectKey),
	})

	return req.Presign(15 * time.Minute)
}

func GetSignedFileURLWithCompany(c *gin.Context) {
	companyID := c.Param("companyID")
	if companyID == "" {
		c.String(http.StatusBadRequest, "companyID is required")
		return
	}

	objectKey := c.Param("objectKey")
	objectKey = strings.TrimPrefix(objectKey, "/")
	if objectKey == "" {
		c.String(http.StatusBadRequest, "objectKey is required")
		return
	}

	// Generate signed URL
	signedURL, err := GeneratePresignedURLWithCompanyID(c, companyID, objectKey)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to generate signed URL: %s", err.Error())
		return
	}

	// Fetch file dari E2/S3
	resp, err := http.Get(signedURL)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to fetch file from E2: %s", err.Error())
		return
	}
	defer resp.Body.Close()

	c.Header("Content-Type", resp.Header.Get("Content-Type"))
	c.Header("Cache-Control", "no-cache")
	io.Copy(c.Writer, resp.Body)
}
