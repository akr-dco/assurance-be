package models

import (
	"time"

	"gorm.io/gorm"
)

type MstrInspection struct {
	Id             uint           `json:"id" gorm:"primaryKey;autoIncrement;comment:Primary key (auto increment) for unique inspection"`
	NameInspection string         `json:"name_inspection" gorm:"type:varchar(200);not null;comment:Name of the inspection"`
	ImageUrl       string         `json:"image_url" gorm:"type:varchar(500);not null;comment:URL of the inspection image"`
	CompanyID      string         `json:"company_id" gorm:"type:varchar(50);not null;comment:Foreign key reference to Company (MstrCompany.CompanyID)"`
	CreatedAt      time.Time      `json:"created_at" gorm:"autoCreateTime;comment:Timestamp when inspection was first created"`
	UpdatedAt      time.Time      `json:"updated_at" gorm:"autoUpdateTime;comment:Timestamp when inspection was last updated"`
	CreatedBy      string         `json:"created_by" gorm:"type:varchar(100);comment:User or system that created the inspection record"`
	UpdatedBy      string         `json:"updated_by" gorm:"type:varchar(100);comment:User or system that last updated the inspection record"`
	DeletedBy      string         `json:"deleted_by" gorm:"type:varchar(100);comment:User or system that deleted the record"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at" gorm:"index;comment:Timestamp when the record was soft deleted"`

	Details []MstrInspectionDetail `gorm:"foreignKey:IdMstrInspection;constraint:OnDelete:CASCADE;comment:List of coordinates/details for this inspection"`
	Groups  []MstrGroup            `gorm:"many2many:mstr_group_inspection;constraint:OnDelete:CASCADE;comment:Groups assigned to this inspection" json:"groups"`
}

func (MstrInspection) TableName() string {
	return "mstr_inspection"
}

type MstrInspectionDetail struct {
	Id                 uint             `json:"id" gorm:"primaryKey;autoIncrement;comment:Primary key (auto increment) for inspection detail"`
	IdMstrInspection   uint             `json:"id_mstr_inspection" gorm:"not null;comment:Foreign key to MstrInspection"`
	NameCoordinate     string           `json:"name_coordinate" gorm:"type:varchar(200);not null;comment:Name/label of the coordinate point"`
	X                  float64          `json:"x" gorm:"comment:X coordinate value"`
	Y                  float64          `json:"y" gorm:"comment:Y coordinate value"`
	TutorialCoordinate string           `json:"tutorial_coordinate" gorm:"type:varchar(500);comment:Instruction or tutorial for this coordinate"`
	RequiredCoordinate bool             `json:"required_coordinate" gorm:"default:false;comment:Whether this coordinate is mandatory"`
	SendNow            bool             `json:"send_now" gorm:"default:false;comment:Send now for SAM"`
	TypeTriggerID      *uint            `json:"type_trigger_id" gorm:"column:type_trigger_id"`
	Types              *MstrTypeTrigger `json:"types" gorm:"foreignKey:TypeTriggerID;references:Id"`
	CreatedAt          time.Time        `json:"created_at" gorm:"autoCreateTime;comment:Timestamp when detail was created"`
	UpdatedAt          time.Time        `json:"updated_at" gorm:"autoUpdateTime;comment:Timestamp when detail was last updated"`
	CreatedBy          string           `json:"created_by" gorm:"type:varchar(100);comment:User or system that created the detail record"`
	UpdatedBy          string           `json:"updated_by" gorm:"type:varchar(100);comment:User or system that last updated the detail record"`
	DeletedBy          string           `json:"deleted_by" gorm:"type:varchar(100);comment:User or system that deleted the record"`
	DeletedAt          gorm.DeletedAt   `json:"deleted_at" gorm:"index;comment:Timestamp when the record was soft deleted"`

	Questions []MstrInspectionQuestion `json:"questions" gorm:"foreignKey:InspectionDetailID;constraint:OnDelete:CASCADE;comment:List of questions belonging to this Inspection Detail"`
}

func (MstrInspectionDetail) TableName() string {
	return "mstr_inspection_detail"
}

type MstrInspectionQuestion struct {
	ID                 uint           `json:"id" gorm:"primaryKey;autoIncrement;comment:Primary key for inspection question"`
	InspectionDetailID uint           `json:"inspection_detail_id" gorm:"index;not null;comment:Foreign key to inspection detail"`
	Text               string         `json:"text" gorm:"type:text;not null;comment:Question text"`
	Type               string         `json:"type" gorm:"type:varchar(20);not null;comment:Question type (yesno|multiple|essay|image)"`
	CreatedBy          string         `json:"created_by" gorm:"type:varchar(100);comment:User or system that created the question"`
	UpdatedBy          string         `json:"updated_by" gorm:"type:varchar(100);comment:User or system that last updated the question"`
	CreatedAt          time.Time      `json:"created_at" gorm:"autoCreateTime;comment:Timestamp when question was created"`
	UpdatedAt          time.Time      `json:"updated_at" gorm:"autoUpdateTime;comment:Timestamp when question was last updated"`
	DeletedBy          string         `json:"deleted_by" gorm:"type:varchar(100);comment:User or system that deleted the record"`
	DeletedAt          gorm.DeletedAt `json:"deleted_at" gorm:"index;comment:Timestamp when the record was soft deleted"`

	// Options are used only for multiple-choice or yes/no or essay or image
	Options []MstrInspectionQuestionOption `json:"options" gorm:"foreignKey:InspectionQuestionID;constraint:OnDelete:CASCADE;comment:List of options for multiple-choice inspection questions"`
}

func (MstrInspectionQuestion) TableName() string {
	return "mstr_inspection_question"
}

type MstrInspectionQuestionOption struct {
	ID                   uint      `json:"id" gorm:"primaryKey;autoIncrement;comment:Primary key for inspection question option"`
	InspectionQuestionID uint      `json:"inspection_question_id" gorm:"index;not null;comment:Foreign key to Inspection Question"`
	Label                string    `json:"label" gorm:"type:varchar(10);not null;comment:Option label like A,B,C,D,E"`
	Text                 string    `json:"text" gorm:"type:varchar(200);not null;comment:Text description of the option"`
	IsCorrect            bool      `json:"is_correct" gorm:"default:false;comment:Optional flag indicating if option is correct"`
	CreatedBy            string    `json:"created_by" gorm:"type:varchar(100);comment:User or system that created the question"`
	UpdatedBy            string    `json:"updated_by" gorm:"type:varchar(100);comment:User or system that last updated the question"`
	CreatedAt            time.Time `json:"created_at" gorm:"autoCreateTime;comment:Timestamp when option was created"`
	UpdatedAt            time.Time `json:"updated_at" gorm:"autoUpdateTime;comment:Timestamp when option was last updated"`
	//DeletedBy            string         `json:"deleted_by" gorm:"type:varchar(100);comment:User or system that deleted the record"`
	//DeletedAt            gorm.DeletedAt `json:"deleted_at" gorm:"index;comment:Timestamp when the record was soft deleted"`
}

func (MstrInspectionQuestionOption) TableName() string {
	return "mstr_inspection_question_option"
}

type MstrTypeTrigger struct {
	Id          uint           `json:"id" gorm:"primaryKey;autoIncrement;comment:Primary key for Master Type Trigger"`
	TypeName    string         `json:"type_name" gorm:"type:varchar(100);not null;comment:Name of the type trigger (e.g., banjir, gempa)"`
	Description string         `json:"description" gorm:"type:varchar(255);comment:Description of the type trigger"`
	IsActive    bool           `json:"is_active" gorm:"default:true;comment:Type trigger status"`
	CompanyID   string         `json:"company_id" gorm:"type:varchar(50);not null;comment:Foreign key to Company (MstrCompany.CompanyID)"`
	CreatedBy   string         `json:"created_by" gorm:"type:varchar(100)"`
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedBy   string         `json:"updated_by" gorm:"type:varchar(100)"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedBy   string         `json:"deleted_by" gorm:"type:varchar(100);comment:User or system that deleted the record"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (MstrTypeTrigger) TableName() string {
	return "mstr_type_triggers"
}
