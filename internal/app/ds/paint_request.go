package ds

import (
	"time"
	"database/sql"
)

type PaintRequest struct {
	ID           uint            `gorm:"primaryKey"`
	Status       string          `gorm:"type:varchar(20);not null;default:'черновик'"`
	DateCreate   time.Time       `gorm:"not null;autoCreateTime"`
	DateForm     sql.NullTime    `gorm:"default:null"`     
	DateFinish   sql.NullTime    `gorm:"default:null"`     
	CreatorID    uint            `gorm:"not null"`
	ModeratorID  sql.NullInt64
	MinLayers    int             `gorm:"-"`//`gorm:"not null;default:1"`
	RequestPaints []RequestPaint `gorm:"foreignKey:RequestID"`
}