package ds

import "github.com/google/uuid"

type User struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Login       string `gorm:"type:varchar(50);unique;not null"`
	Password    string `gorm:"type:varchar(100);not null"`
	IsModerator bool   `gorm:"type:boolean;default:false"`
}