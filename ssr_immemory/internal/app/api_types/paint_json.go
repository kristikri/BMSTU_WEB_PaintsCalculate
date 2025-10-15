package apitypes

import "ssr_immemory/internal/app/ds"

type PaintJSON struct {
	ID          uint    `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	HidingPower float64 `json:"hiding_power"`
	Photo       string  `json:"photo"`
	IsDelete    bool    `json:"is_delete"`
}

func PaintToJSON(paint ds.Paint) PaintJSON {
	return PaintJSON{
		ID:          paint.ID,
		Title:       paint.Title,
		Description: paint.Description,
		HidingPower: paint.HidingPower,
		Photo:       paint.Photo,
		IsDelete:    paint.IsDelete,
	}
}

func PaintFromJSON(paintJSON PaintJSON) ds.Paint {
	return ds.Paint{
		Title:       paintJSON.Title,
		Description: paintJSON.Description,
		HidingPower: paintJSON.HidingPower,
		Photo:       paintJSON.Photo,
		IsDelete:    paintJSON.IsDelete,
	}
}
