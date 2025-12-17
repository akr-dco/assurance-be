package models

import (
	"time"

	"gorm.io/gorm"
)

// Enum yang dipakai di code: "yesno", "multiple", "essay", "image"

type Questionnaire struct {
	ID          uint           `json:"id" gorm:"primaryKey;autoIncrement;comment:Primary key for questionnaire"`
	Title       string         `json:"title" gorm:"type:varchar(200);not null;comment:Title of the questionnaire"`
	Description string         `json:"description" gorm:"type:text;comment:Detailed description of the questionnaire"`
	Type        string         `json:"type" gorm:"type:varchar(20);comment:Type of questionnaire Pre-Inspection or Post-Inspection"`
	IsActive    bool           `json:"is_active" gorm:"default:true;comment:Whether the questionnaire is active"`
	CompanyID   string         `json:"company_id" gorm:"type:varchar(50);index;comment:Foreign key to Company"`
	CreatedBy   string         `json:"created_by" gorm:"type:varchar(100);comment:User or system that created this questionnaire"`
	UpdatedBy   string         `json:"updated_by" gorm:"type:varchar(100);comment:User or system that last updated this questionnaire"`
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime;comment:Timestamp when questionnaire was created"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"autoUpdateTime;comment:Timestamp when questionnaire was last updated"`
	DeletedBy   string         `json:"deleted_by" gorm:"type:varchar(100);comment:User or system that deleted the record"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index;comment:Timestamp when the record was soft deleted"`

	Questions  []Question   `json:"questions" gorm:"foreignKey:QuestionnaireID;constraint:OnDelete:CASCADE;comment:List of questions belonging to this questionnaire"`
	Groups     []MstrGroup  `gorm:"many2many:mstr_group_questionnaire;comment:List of groups associated with the questionnaire" json:"groups"`
	Mstranswer []MstrAnswer `json:"mstranswer" gorm:"foreignKey:QuestionnaireID;constraint:OnDelete:CASCADE;comment:List of master answer belonging to this questionnaire"`
}

type Question struct {
	ID              uint           `json:"id" gorm:"primaryKey;autoIncrement;comment:Primary key for question"`
	QuestionnaireID uint           `json:"questionnaire_id" gorm:"index;not null;comment:Foreign key to Questionnaire"`
	Text            string         `json:"text" gorm:"type:text;not null;comment:Question text"`
	Type            string         `json:"type" gorm:"type:varchar(20);not null;comment:Question type (yesno|multiple|essay|image)"`
	CreatedBy       string         `json:"created_by" gorm:"type:varchar(100);comment:User or system that created the question"`
	UpdatedBy       string         `json:"updated_by" gorm:"type:varchar(100);comment:User or system that last updated the question"`
	CreatedAt       time.Time      `json:"created_at" gorm:"autoCreateTime;comment:Timestamp when question was created"`
	UpdatedAt       time.Time      `json:"updated_at" gorm:"autoUpdateTime;comment:Timestamp when question was last updated"`
	DeletedBy       string         `json:"deleted_by" gorm:"type:varchar(100);comment:User or system that deleted the record"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at" gorm:"index;comment:Timestamp when the record was soft deleted"`

	// Options are used only for multiple-choice or yes/no or essay or image
	Options          []Option           `json:"options" gorm:"foreignKey:QuestionID;constraint:OnDelete:CASCADE;comment:List of options for multiple-choice questions"`
	Answers          []Answer           `json:"answers" gorm:"foreignKey:QuestionID;constraint:OnDelete:CASCADE;comment:List of options for multiple-choice questions"`
	Mstranswerdetail []MstrAnswerDetail `json:"mstranswerdetail" gorm:"foreignKey:QuestionID;constraint:OnDelete:CASCADE;comment:List of options for multiple-choice questions"`
}

type Option struct {
	ID         uint      `json:"id" gorm:"primaryKey;autoIncrement;comment:Primary key for option"`
	QuestionID uint      `json:"question_id" gorm:"index;not null;comment:Foreign key to Question"`
	Label      string    `json:"label" gorm:"type:varchar(10);not null;comment:Option label like A,B,C,D,E"`
	Text       string    `json:"text" gorm:"type:varchar(200);not null;comment:Text description of the option"`
	IsCorrect  bool      `json:"is_correct" gorm:"default:false;comment:Optional flag indicating if option is correct"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime;comment:Timestamp when option was created"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime;comment:Timestamp when option was last updated"`
	//DeletedBy  string         `json:"deleted_by" gorm:"type:varchar(100);comment:User or system that deleted the record"`
	//DeletedAt  gorm.DeletedAt `json:"deleted_at" gorm:"index;comment:Timestamp when the record was soft deleted"`
}

type Answer struct {
	ID         uint           `json:"id" gorm:"primaryKey;autoIncrement;comment:Primary key for answer"`
	QuestionID uint           `json:"question_id" gorm:"index;not null;comment:Foreign key to Question"`
	UserID     uint           `json:"user_id" gorm:"comment:Foreign key to User who answered"`
	AnswerText string         `json:"answer_text" gorm:"type:text;comment:User's answer in text format (Yes/No, A..E, essay)"`
	AnswerFile string         `json:"answer_file" gorm:"type:varchar(255);comment:Relative path to uploaded image for image-type question"`
	CreatedBy  string         `json:"created_by" gorm:"type:varchar(100);comment:User or system that created this answer record"`
	UpdatedBy  string         `json:"updated_by" gorm:"type:varchar(100);comment:User or system that last updated this answer record"`
	CreatedAt  time.Time      `json:"created_at" gorm:"autoCreateTime;comment:Timestamp when answer was created"`
	UpdatedAt  time.Time      `json:"updated_at" gorm:"autoUpdateTime;comment:Timestamp when answer was last updated"`
	DeletedBy  string         `json:"deleted_by" gorm:"type:varchar(100);comment:User or system that deleted the record"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at" gorm:"index;comment:Timestamp when the record was soft deleted"`
}

// MasterAnswer = 1 kali submit questionnaire
type MstrAnswer struct {
	ID              uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	QuestionnaireID uint           `json:"questionnaire_id" gorm:"index;not null;comment:Foreign key to Questionnaire"`
	UserID          uint           `json:"user_id" gorm:"not null;index"`
	DeviceID        string         `json:"device_id" gorm:"type:varchar(50);not null;comment:device identifier"`
	CompanyID       string         `json:"company_id" gorm:"type:varchar(50);comment:reference to Company (MstrCompany.CompanyID)"`
	ChainingID      uint           `json:"chaining_id" gorm:"comment:chaining_id"`
	CreatedBy       string         `json:"created_by" gorm:"type:varchar(100);comment:User or system that created this answer record"`
	UpdatedBy       string         `json:"updated_by" gorm:"type:varchar(100);comment:User or system that last updated this answer record"`
	CreatedAt       time.Time      `json:"created_at" gorm:"autoCreateTime;comment:Timestamp when answer was created"`
	UpdatedAt       time.Time      `json:"updated_at" gorm:"autoUpdateTime;comment:Timestamp when answer was last updated"`
	DeletedBy       string         `json:"deleted_by" gorm:"type:varchar(100);comment:User or system that deleted the record"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at" gorm:"index;comment:Timestamp when the record was soft deleted"`

	Details []MstrAnswerDetail `json:"details" gorm:"foreignKey:MasterAnswerID;constraint:OnDelete:CASCADE"`
}

func (MstrAnswer) TableName() string {
	return "mstr_answer"
}

// MasterAnswerDetail = jawaban per question
type MstrAnswerDetail struct {
	ID             uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	MasterAnswerID uint           `json:"master_answer_id" gorm:"index;not null;comment:Foreign key to Master Answer"`
	QuestionID     uint           `json:"question_id" gorm:"index;not null;comment:Foreign key to Question"`
	AnswerText     string         `json:"answer_text" gorm:"type:text;comment:User's answer in text format (Yes/No, A..E, essay)"`
	AnswerFile     string         `json:"answer_file" gorm:"type:varchar(255);comment:Relative path to uploaded image for image-type question"`
	CreatedBy      string         `json:"created_by" gorm:"type:varchar(100);comment:User or system that created this answer record"`
	UpdatedBy      string         `json:"updated_by" gorm:"type:varchar(100);comment:User or system that last updated this answer record"`
	CreatedAt      time.Time      `json:"created_at" gorm:"autoCreateTime;comment:Timestamp when answer was created"`
	UpdatedAt      time.Time      `json:"updated_at" gorm:"autoUpdateTime;comment:Timestamp when answer was last updated"`
	DeletedBy      string         `json:"deleted_by" gorm:"type:varchar(100);comment:User or system that deleted the record"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at" gorm:"index;comment:Timestamp when the record was soft deleted"`
}

func (MstrAnswerDetail) TableName() string {
	return "mstr_answer_detail"
}
