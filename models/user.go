package models

import (
	"time"

	"gorm.io/gorm"
)

type MstrUser struct {
	Id        uint           `json:"id" gorm:"primaryKey;autoIncrement;comment:Primary key (auto increment) for unique user identity"`
	Username  string         `json:"username" gorm:"type:varchar(100);unique;not null;comment:Unique username for login"`
	Password  string         `json:"-" gorm:"type:varchar(255);not null;comment:Hashed password (not exposed in JSON)"`
	FullName  string         `json:"full_name" gorm:"type:varchar(200);comment:Full name of the user"`
	Email     string         `json:"email" gorm:"type:varchar(200);unique;comment:Email address of the user"`
	Phone     string         `json:"phone" gorm:"type:varchar(20);comment:Phone number of the user"`
	Role      string         `json:"role" gorm:"type:varchar(50);not null;comment:Role or permission level of the user"`
	CompanyID string         `json:"company_id" gorm:"type:varchar(50);comment:Foreign key reference to Company (MstrCompany.CompanyID)"`
	IsActive  bool           `json:"is_active" gorm:"default:true;comment:User status (true = active, false = inactive)"`
	LastLogin *time.Time     `json:"last_login" gorm:"comment:Timestamp of user's last login"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime;comment:Timestamp when the user record was first created"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime;comment:Timestamp when the user record was last updated"`
	CreatedBy string         `json:"created_by" gorm:"type:varchar(100);comment:User or system that created this record"`
	UpdatedBy string         `json:"updated_by" gorm:"type:varchar(100);comment:User or system that last updated this record"`
	DeletedBy string         `json:"deleted_by" gorm:"type:varchar(100);comment:User or system that deleted the record"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index;comment:Timestamp when the record was soft deleted"`
}

func (MstrUser) TableName() string {
	return "mstr_user"
}

type CreateUserRequest struct {
	Username  string     `json:"username"`
	Password  string     `json:"password"`
	FullName  string     `json:"full_name"`
	Email     string     `json:"email" gorm:"unique"`
	Phone     string     `json:"phone"`
	Role      string     `json:"role"`
	CompanyID string     `json:"company_id"`
	IsActive  bool       `json:"is_active"`
	LastLogin *time.Time `json:"last_login"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	CreatedBy string     `json:"created_by"`
	UpdatedBy string     `json:"updated_by"`
}

type UpdateUserRequest struct {
	Username  string `json:"username"`
	FullName  string `json:"full_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Role      string `json:"role"`
	IsActive  bool   `json:"is_active"`
	Password  string `json:"password,omitempty"`
	UpdatedBy string `json:"updated_by"`
}

type PasswordResetToken struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	UserID    uint      `gorm:"not null"`
	Token     string    `gorm:"type:text;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	Used      bool      `gorm:"default:false"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}
