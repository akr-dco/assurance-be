package models

import (
	"time"

	"gorm.io/gorm"
)

type MstrChaining struct {
	Id           uint   `json:"id" gorm:"primaryKey;autoIncrement;comment:Primary key for Master Chaining"`
	NameChaining string `json:"name_chaining" gorm:"type:varchar(200);not null;comment:Name of the Chaining"`

	//Bagian untuk trigger
	TriggerDatetime time.Time         `json:"trigger_datetime" gorm:"comment:Datetime when chaining trigger starts"`
	FrequencyValue  *uint             `json:"frequency_value" gorm:"comment:Frequency interval value (e.g. 1, 2, 30)"`
	FrequencyUnit   *string           `json:"frequency_unit" gorm:"type:varchar(20);comment:Unit of frequency (minute, hour, day)"`
	EventTriggerID  *uint             `json:"event_trigger_id" gorm:"column:event_trigger_id"`
	Events          *MstrEventTrigger `json:"events" gorm:"foreignKey:EventTriggerID;references:Id"`

	IsActive  bool           `json:"is_active" gorm:"default:true;comment:Chaining status (true = active, false = inactive)"`
	CompanyID string         `json:"company_id" gorm:"type:varchar(50);not null;comment:Foreign key to Company (MstrCompany.CompanyID)"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime;comment:Timestamp when transaction was created"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime;comment:Timestamp when transaction was last updated"`
	CreatedBy string         `json:"created_by" gorm:"type:varchar(100);not null;comment:User or system that created this record"`
	UpdatedBy string         `json:"updated_by" gorm:"type:varchar(100);comment:User or system that last updated this record"`
	DeletedBy string         `json:"deleted_by" gorm:"type:varchar(100);comment:User or system that deleted the record"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index;comment:Timestamp when the record was soft deleted"`

	Details []MstrChainingDetail `json:"details" gorm:"foreignKey:IdChaining;references:Id;constraint:OnDelete:CASCADE;comment:List of detailed Chaining"`
	//EventTrigger []MstrEventTrigger   `json:"event_trigger" gorm:"foreignKey:EventTriggerID;references:Id"`
}

type MstrChainingDetail struct {
	Id         uint           `json:"id" gorm:"primaryKey;autoIncrement;comment:Primary key for chaining detail"`
	IdChaining uint           `json:"id_chaining" gorm:"not null;comment:Foreign key to MstrChaining"`
	ItemType   string         `json:"item_type" gorm:"type:varchar(50);not null;comment:Type of item (inspection/questionnaire)"`
	ItemID     uint           `json:"item_id" gorm:"not null;comment:ID of inspection or questionnaire based on item_type"`
	Sequence   uint           `json:"sequence" gorm:"not null;comment:Order of the item in the chaining"`
	CreatedAt  time.Time      `json:"created_at" gorm:"autoCreateTime;comment:Timestamp when detail was created"`
	UpdatedAt  time.Time      `json:"updated_at" gorm:"autoUpdateTime;comment:Timestamp when detail was last updated"`
	CreatedBy  string         `json:"created_by" gorm:"type:varchar(100);comment:User or system that created this detail record"`
	UpdatedBy  string         `json:"updated_by" gorm:"type:varchar(100);comment:User or system that last updated this detail record"`
	DeletedBy  string         `json:"deleted_by" gorm:"type:varchar(100);comment:User or system that deleted the record"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at" gorm:"index;comment:Timestamp when the record was soft deleted"`
	// Relasi optional
	//Inspection    *MstrInspection `json:"inspection,omitempty" gorm:"foreignKey:ItemID;references:Id"`
	//Questionnaire *Questionnaire  `json:"questionnaire,omitempty" gorm:"foreignKey:ItemID;references:ID"`

	// Relasi opsional → hilangkan foreignKey
	Inspection    *MstrInspection `json:"inspection,omitempty" gorm:"-"`
	Questionnaire *Questionnaire  `json:"questionnaire,omitempty" gorm:"-"`
}

type MstrEventTrigger struct {
	Id          uint           `json:"id" gorm:"primaryKey;autoIncrement;comment:Primary key for Master Event Trigger"`
	EventName   string         `json:"event_name" gorm:"type:varchar(100);not null;comment:Name of the event trigger (e.g., banjir, gempa)"`
	Trigger     bool           `json:"trigger" gorm:"default:false;comment:Event trigger status"`
	Description string         `json:"description" gorm:"type:varchar(255);comment:Description of the event trigger"`
	IsActive    bool           `json:"is_active" gorm:"default:true;comment:Event trigger status"`
	CompanyID   string         `json:"company_id" gorm:"type:varchar(50);not null;comment:Foreign key to Company (MstrCompany.CompanyID)"`
	CreatedBy   string         `json:"created_by" gorm:"type:varchar(100)"`
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedBy   string         `json:"updated_by" gorm:"type:varchar(100)"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedBy   string         `json:"deleted_by" gorm:"type:varchar(100);comment:User or system that deleted the record"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (MstrEventTrigger) TableName() string {
	return "mstr_event_triggers"
}
