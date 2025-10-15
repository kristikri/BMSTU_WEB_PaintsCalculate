package ds

type RequestsPaint struct {
	ID        uint       `gorm:"primaryKey"`
	RequestID uint       `gorm:"not null;uniqueIndex:idx_request_paint"`
	PaintID   uint       `gorm:"not null;uniqueIndex:idx_request_paint"`
	Area      float64    `gorm:"type:decimal(8,2);not null;default:0"`    
	Layers    int        `gorm:"not null;default:1"`                      
	Quantity  float64    `gorm:"type:decimal(10,2);default:0"`            
	Request   PaintRequest `gorm:"foreignKey:RequestID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Paint     Paint        `gorm:"foreignKey:PaintID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}