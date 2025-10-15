package repository

import (
	"database/sql"
	"errors"
	"fmt"
	apitypes "ssr_immemory/internal/app/api_types"
	"ssr_immemory/internal/app/ds"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Обновляем методы чтобы принимать userID как параметр
func (r *Repository) GetPaintRequests(from, to time.Time, status string) ([]ds.PaintRequest, error) {
	var requests []ds.PaintRequest
	sub := r.db.Where("status != 'удалён' and status != 'черновик'")
	if !from.IsZero() {
		sub = sub.Where("date_create > ?", from)
	}
	if !to.IsZero() {
		sub = sub.Where("date_create < ?", to.Add(time.Hour*24))
	}
	if status != "" {
		sub = sub.Where("status = ?", status)
	}
	err := sub.Order("id").Find(&requests).Error
	if err != nil {
		return nil, err
	}
	return requests, nil
}

func (r *Repository) GetRequestPaints(requestId uint) ([]ds.RequestsPaint, error) {
    var requestPaints []ds.RequestsPaint
    err := r.db.Where("request_id = ?", requestId).Find(&requestPaints).Error
    if err != nil {
        return nil, err
    }
    return requestPaints, nil
}

func (r *Repository) GetRequestPaintsList(id int) ([]ds.Paint, ds.PaintRequest, error) {
	request, err := r.GetSinglePaintRequest(id)
	if err != nil {
		return []ds.Paint{}, ds.PaintRequest{}, err
	}

	var paints []ds.Paint
	sub := r.db.Table("request_paints").Where("request_id = ?", request.CreatorID)
	err = r.db.Order("id DESC").Where("id IN (?)", sub.Select("paint_id")).Find(&paints).Error

	if err != nil {
		return []ds.Paint{}, ds.PaintRequest{}, err
	}

	return paints, request, nil
}

func (r *Repository) CheckCurrentRequestDraft(creatorID uuid.UUID) (ds.PaintRequest, error) {
	var request ds.PaintRequest
	res := r.db.Where("creator_id = ? AND status = ?", creatorID, "черновик").Limit(1).Find(&request)
	if res.Error != nil {
		return ds.PaintRequest{}, res.Error
	} else if res.RowsAffected == 0 {
		return ds.PaintRequest{}, ErrNoDraft
	}
	return request, nil
}

func (r *Repository) GetRequestDraft(creatorID uuid.UUID) (ds.PaintRequest, bool, error) {
	request, err := r.CheckCurrentRequestDraft(creatorID)
	if errors.Is(err, ErrNoDraft) {
		request = ds.PaintRequest{
			Status:     "черновик",
			CreatorID:  creatorID,
			DateCreate: time.Now(),
		}
		result := r.db.Create(&request)
		if result.Error != nil {
			return ds.PaintRequest{}, false, result.Error
		}
		return request, true, nil
	} else if err != nil {
		return ds.PaintRequest{}, false, err
	}
	return request, true, nil
}

func (r *Repository) GetSinglePaintRequest(id int) (ds.PaintRequest, error) {
	if id < 0 {
		return ds.PaintRequest{}, errors.New("неверное id, должно быть >= 0")
	}

	var request ds.PaintRequest
	err := r.db.Where("id = ?", id).First(&request).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.PaintRequest{}, fmt.Errorf("%w: заявка с id %d", ErrNotFound, id)
		}
		return ds.PaintRequest{}, err
	} else if request.Status == "удалён" {
		return ds.PaintRequest{}, fmt.Errorf("%w: заявка удалена", ErrNotAllowed)
	}
	return request, nil
}

func (r *Repository) FormRequest(requestId int, status string) (ds.PaintRequest, error) {
	request, err := r.GetSinglePaintRequest(requestId)
	if err != nil {
		return ds.PaintRequest{}, err
	}

	if request.Status != "черновик" {
		return ds.PaintRequest{}, fmt.Errorf("эта заявка не может быть %s", status)
	}

	if status != "удалён" {
		if request.MinLayers <= 0 {
			return ds.PaintRequest{}, errors.New("вы не указали минимальное количество слоев")
		}
		requestPaints, _ := r.GetRequestPaints(request.ID)
		for _, requestPaint := range requestPaints {
			if requestPaint.Area == 0 {
				return ds.PaintRequest{}, errors.New("вы не указали площадь для краски")
			}
			if requestPaint.Layers < request.MinLayers {
				return ds.PaintRequest{}, fmt.Errorf("количество слоев для краски должно быть не менее %d", request.MinLayers)
			}
		}
	}

	err = r.db.Model(&request).Updates(ds.PaintRequest{
		Status: status,
		DateForm: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
	}).Error
	if err != nil {
		return ds.PaintRequest{}, err
	}

	return request, nil
}

func (r *Repository) ChangeRequest(id int, requestJSON apitypes.PaintRequestJSON) (ds.PaintRequest, error) {
	request := ds.PaintRequest{}
	if id < 0 {
		return ds.PaintRequest{}, errors.New("неправильное id, должно быть >= 0")
	}
	if requestJSON.MinLayers <= 0 {
		return ds.PaintRequest{}, errors.New("неправильное минимальное количество слоев")
	}
	err := r.db.Where("id = ? and status != 'удалён'", id).First(&request).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.PaintRequest{}, fmt.Errorf("%w: заявка с id %d", ErrNotFound, id)
		}
		return ds.PaintRequest{}, err
	}
	err = r.db.Model(&request).Updates(apitypes.PaintRequestFromJSON(requestJSON)).Error
	if err != nil {
		return ds.PaintRequest{}, err
	}
	return request, nil
}

// Восстанавливаем функцию CalculatePaintQuantity
func CalculatePaintQuantity(hidingPower float64, area float64, layers int) (float64, error) {
	if area <= 0 {
		return 0, errors.New("неправильная площадь")
	}
	if layers <= 0 {
		return 0, errors.New("неправильное количество слоев")
	}
	return area * hidingPower * float64(layers) / 1000, nil
}

func (r *Repository) GetPaintCount(creatorID uuid.UUID) int64 {
	if creatorID == uuid.Nil {
		return 0
	}

	var count int64
	request, err := r.CheckCurrentRequestDraft(creatorID)
	if err != nil {
		return 0
	}
	err = r.db.Model(&ds.RequestsPaint{}).Where("request_id = ?", request.ID).Count(&count).Error
	if err != nil {
		logrus.Println("Error counting records in request_paints:", err)
	}

	return count
}

func (r *Repository) ModerateRequest(id int, status string, moderatorID uuid.UUID) (ds.PaintRequest, error) {
	if status != "завершена" && status != "отклонена" {
		return ds.PaintRequest{}, errors.New("неверный статус")
	}

	user, err := r.GetUserByID(moderatorID)
	if err != nil {
		return ds.PaintRequest{}, err
	}

	if !user.IsModerator {
		return ds.PaintRequest{}, fmt.Errorf("%w: вы не модератор", ErrNotAllowed)
	}

	request, err := r.GetSinglePaintRequest(id)
	if err != nil {
		return ds.PaintRequest{}, err
	} else if request.Status != "сформирована" {
		return ds.PaintRequest{}, fmt.Errorf("эта заявка не может быть %s", status)
	}

	moderatorIDUUID := uuid.NullUUID{
    UUID:  user.ID,
    Valid: true,
}

	err = r.db.Model(&request).Updates(ds.PaintRequest{
		Status: status,
		DateFinish: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		ModeratorID: moderatorIDUUID,
	}).Error
	if err != nil {
		return ds.PaintRequest{}, err
	}

	if status == "завершена" {
		requestPaints, err := r.GetRequestPaints(request.ID)
		if err != nil {
			return ds.PaintRequest{}, err
		}
		for _, requestPaint := range requestPaints {
			paint, err := r.GetPaint(int(requestPaint.PaintID))
			if err != nil {
				return ds.PaintRequest{}, err
			}
			quantity, err := CalculatePaintQuantity(paint.HidingPower, requestPaint.Area, requestPaint.Layers)
			if err != nil {
				return ds.PaintRequest{}, err
			}
			err = r.db.Model(&requestPaint).Updates(ds.RequestsPaint{
				Quantity: quantity,
			}).Error
			if err != nil {
				return ds.PaintRequest{}, err
			}
		}
	}

	return request, nil
}
