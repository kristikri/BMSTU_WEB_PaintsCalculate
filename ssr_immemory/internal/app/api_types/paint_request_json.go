package apitypes

import (
	"ssr_immemory/internal/app/ds"
	"time"
)

type PaintRequestJSON struct {
	ID               uint            `gorm:"primaryKey"`
	Status           string          `gorm:"type:varchar(20);not null;default:'черновик'"`
	DateCreate		 *time.Time		 `gorm:"default:null"`
	DateForm         *time.Time      `gorm:"default:null"`     
	DateFinish       *time.Time      `gorm:"default:null"`     
	CreatorLogin	 string		     `json:"creator_login"`
	ModeratorLogin	 *string		 `json:"moderator_login"`
	MinLayers         int            `json:"min_layers" gorm:"default:1"`
}

func PaintRequestToJSON(request ds.PaintRequest, creatorLogin string, moderatorLogin string) PaintRequestJSON {
	var dateForm, dateFinish *time.Time

	if request.DateForm.Valid {
		dateForm = &request.DateForm.Time
	}
	if request.DateFinish.Valid {
		dateFinish = &request.DateFinish.Time
	}
	
	var mLogin *string
	if moderatorLogin != "" {
		mLogin = &moderatorLogin
	}
 	
	return PaintRequestJSON{
		ID:             request.ID,
		Status:         request.Status,
		CreatorLogin:   creatorLogin,
		ModeratorLogin: mLogin,
		DateCreate:		&request.DateCreate,
		DateForm:       dateForm,
		DateFinish:     dateFinish,
		MinLayers:      request.MinLayers,
	}
}

func PaintRequestFromJSON(request PaintRequestJSON) ds.PaintRequest {
	return ds.PaintRequest{
		MinLayers: request.MinLayers,
	}
}


type StatusJSON struct {
	Status string `json:"status"`
}

