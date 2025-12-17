package models

import (
	"time"

	"gorm.io/gorm"
)

type MstrDevice struct {
	Id         uint           `json:"id" gorm:"primaryKey;autoIncrement;comment:Primary key (auto increment) for unique device identity"`
	DeviceName string         `json:"device_name" gorm:"type:varchar(200);comment:Official device name"`
	DeviceID   string         `json:"device_id" gorm:"type:varchar(50);unique;not null;comment:Unique device identifier (used as reference in transactions)"`
	CompanyID  string         `json:"company_id" gorm:"type:varchar(50);comment:Foreign key reference to Company (MstrCompany.CompanyID)"`
	IsActive   bool           `json:"is_active" gorm:"default:false;comment:Device status (true = active, false = inactive)"`
	CreatedBy  string         `json:"created_by" gorm:"type:varchar(100);comment:User or system that created the device record"`
	CreatedAt  time.Time      `json:"created_at" gorm:"autoCreateTime;comment:Timestamp when the device record was first created"`
	UpdatedBy  string         `json:"updated_by" gorm:"type:varchar(100);comment:User or system that last updated the device record"`
	UpdatedAt  time.Time      `json:"updated_at" gorm:"autoUpdateTime;comment:Timestamp when the device record was last updated"`
	DeletedBy  string         `json:"deleted_by" gorm:"type:varchar(100);comment:User or system that deleted the record"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at" gorm:"index;comment:Timestamp when the record was soft deleted"`

	Groups []MstrGroup `gorm:"many2many:mstr_group_device;comment:List of groups associated with the device" json:"groups"`
}

func (MstrDevice) TableName() string {
	return "mstr_device"
}
