package ds

import (
	
)

type Paint struct {
	ID          uint      `gorm:"primaryKey"`
	Title       string    `gorm:"type:varchar(255) not null"`
	Description string    `gorm:"type:text not null"`
	HidingPower float64   `gorm:"type:decimal(8,2);not null"`
	Photo       string    `gorm:"type:varchar(500)"`
	IsDelete    bool      `gorm:"type:boolean not null;default:false"`
}