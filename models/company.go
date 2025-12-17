package models

import (
	"time"

	"gorm.io/gorm"
)

type MstrCompany struct {
	Id           uint           `json:"id" gorm:"primaryKey;autoIncrement;comment:Primary key (auto increment) for unique company identity"`
	CompanyName  string         `json:"company_name" gorm:"type:varchar(200);unique;not null;comment:Official company name"`
	CompanyID    string         `json:"company_id" gorm:"type:varchar(50);unique;not null;comment:Unique company code (used as foreign key in other tables)"`
	ImageUrl     string         `json:"image_url" gorm:"type:varchar(500);comment:URL of the company logo"`
	E2Endpoint   string         `json:"e2_endpoint" gorm:"type:varchar(100);comment:E2 Endpoint"`
	E2Region     string         `json:"e2_region" gorm:"type:varchar(100);comment:E2 Region"`
	E2BucketName string         `json:"e2_bucket_name" gorm:"type:varchar(100);comment:E2 Bucket Name"`
	E2AccessKey  string         `json:"e2_access_key" gorm:"type:varchar(100);comment:E2 Access Key"`
	E2SecretKey  string         `json:"e2_secret_key" gorm:"type:varchar(100);comment:E2 Secret Key"`
	IsActive     bool           `json:"is_active" gorm:"default:true;comment:Company status (true = active, false = inactive)"`
	CreatedBy    string         `json:"created_by" gorm:"type:varchar(100);comment:User or system that created the company record"`
	CreatedAt    time.Time      `json:"created_at" gorm:"autoCreateTime;comment:Timestamp when the company record was first created"`
	UpdatedBy    string         `json:"updated_by" gorm:"type:varchar(100);comment:User or system that last updated the company record"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"autoUpdateTime;comment:Timestamp when the company record was last updated"`
	DeletedBy    string         `json:"deleted_by" gorm:"type:varchar(100);comment:User or system that deleted the record"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index;comment:Timestamp when the record was soft deleted"`

	Devices        []MstrDevice     `json:"devices" gorm:"foreignKey:CompanyID;references:CompanyID;comment:List of devices owned by the company"`
	Groups         []MstrGroup      `json:"groups" gorm:"foreignKey:CompanyID;references:CompanyID;comment:List of groups associated with the company"`
	Users          []MstrUser       `json:"users" gorm:"foreignKey:CompanyID;references:CompanyID;comment:List of users registered under the company"`
	Inspections    []MstrInspection `json:"inspections" gorm:"foreignKey:CompanyID;references:CompanyID;comment:List of master inspections belonging to the company"`
	TrxInspections []TrxInspection  `json:"trxinspections" gorm:"foreignKey:CompanyID;references:CompanyID;comment:List of inspection transactions performed by the company"`
	Questionnaires []Questionnaire  `json:"questionnaires" gorm:"foreignKey:CompanyID;references:CompanyID;comment:List of Questionnaires performed by the company"`
}

func (MstrCompany) TableName() string {
	return "mstr_company"
}
