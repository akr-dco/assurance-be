package models

import (
	"time"

	"gorm.io/gorm"
)

type MstrGroup struct {
	ID        uint           `json:"id" gorm:"primaryKey;autoIncrement;comment:Primary key (auto increment) for unique group identity"`
	GroupName string         `json:"group_name" gorm:"type:varchar(200);not null;comment:Official group name"`
	CompanyID string         `json:"company_id" gorm:"type:varchar(50);not null;comment:Foreign key reference to Company (MstrCompany.CompanyID)"`
	CreatedBy string         `json:"created_by" gorm:"type:varchar(100);comment:User or system that created the group record"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime;comment:Timestamp when the group record was first created"`
	UpdatedBy string         `json:"updated_by" gorm:"type:varchar(100);comment:User or system that last updated the group record"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime;comment:Timestamp when the group record was last updated"`
	DeletedBy string         `json:"deleted_by" gorm:"type:varchar(100);comment:User or system that deleted the record"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index;comment:Timestamp when the record was soft deleted"`

	Devices        []MstrDevice     `json:"devices" gorm:"many2many:mstr_group_device;constraint:OnDelete:CASCADE;comment:List of devices assigned to the group"`
	Inspections    []MstrInspection `json:"inspections" gorm:"many2many:mstr_group_inspection;constraint:OnDelete:CASCADE;comment:List of inspections associated with the group"`
	Questionnaires []Questionnaire  `json:"questionnaires" gorm:"many2many:mstr_group_questionnaire;constraint:OnDelete:CASCADE;comment:List of questionnaires assigned to the group"`
	Chainings      []MstrChaining   `json:"Chainings" gorm:"many2many:mstr_group_chaining;constraint:OnDelete:CASCADE;comment:List of Chainings assigned to the group"`
}

func (MstrGroup) TableName() string {
	return "mstr_group"
}
