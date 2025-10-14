package apitypes

import "ssr_immemory/internal/app/ds"

type RequestsPaintJSON struct {
	ID        uint       `gorm:"primaryKey"`
	RequestID uint       `gorm:"not null;uniqueIndex:idx_request_paint"`
	PaintID   uint       `gorm:"not null;uniqueIndex:idx_request_paint"`
	Area      float64    `gorm:"type:decimal(8,2);not null;default:0"`    
	Layers    int        `gorm:"not null;default:1"`                      
	Quantity  float64    `gorm:"type:decimal(10,2);default:0"`            
}

func RequestsPaintToJSON (reqPaint ds.RequestsPaint) RequestsPaintJSON {
	return RequestsPaintJSON{
		ID:             reqPaint.ID,
		RequestID:      reqPaint.RequestID,
		PaintID:        reqPaint.PaintID,
		Area:           reqPaint.Area,
		Layers:         reqPaint.Layers,
		Quantity:       reqPaint.Quantity,		
	}
}

func RequestsPaintFromJSON (reqPaintJSON RequestsPaintJSON) ds.RequestsPaint {
	return ds.RequestsPaint{
		Area:           reqPaintJSON.Area,
		Layers:         reqPaintJSON.Layers,
		Quantity:       reqPaintJSON.Quantity,	
	}
}