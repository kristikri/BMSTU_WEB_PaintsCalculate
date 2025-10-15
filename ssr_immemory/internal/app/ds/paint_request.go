package ds

import (
	"time"
	"database/sql"
	"github.com/google/uuid"
)

type PaintRequest struct {
	ID           uint            `gorm:"primaryKey;"`
	Status       string          `gorm:"type:varchar(20);not null;default:'черновик'"`
	DateCreate   time.Time       `gorm:"not null;autoCreateTime"`
	DateForm     sql.NullTime    `gorm:"default:null"`     
	DateFinish   sql.NullTime    `gorm:"default:null"`     
	CreatorID    uuid.UUID        `gorm:"not null"`
	ModeratorID  uuid.NullUUID    `gorm:"type:uuid"`
	MinLayers    int              `gorm:"not null;default:1"`
	RequestPaints []RequestsPaint `gorm:"foreignKey:RequestID"`
}