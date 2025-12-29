package repository

import (
	apitypes "ssr_immemory/internal/app/api_types"
	"ssr_immemory/internal/app/ds"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

func (r *Repository) DeletePaintFromRequest(requestId uint, paintId uint) (ds.PaintRequest, error) {
	var request ds.PaintRequest
	err := r.db.Where("id = ?", requestId).First(&request).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.PaintRequest{}, fmt.Errorf("%w: заявка с id %d", ErrNotFound, requestId)
		}
		return ds.PaintRequest{}, err
	}
    
	err = r.db.Where("paint_id = ? and request_id = ?", paintId, requestId).Delete(&ds.RequestsPaint{}).Error
	if err != nil {
		return ds.PaintRequest{}, err
	}
	return request, nil
}

func (r *Repository) ChangeRequestPaint(requestId uint, paintId uint, requestPaintJSON apitypes.RequestsPaintJSON) (ds.RequestsPaint, error) {
	var requestPaint ds.RequestsPaint
	err := r.db.Model(&requestPaint).Where("paint_id = ? and request_id = ?", paintId, requestId).Updates(apitypes.RequestsPaintFromJSON(requestPaintJSON)).First(&requestPaint).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.RequestsPaint{}, fmt.Errorf("%w: краски в заявке", ErrNotFound)
		}
		return ds.RequestsPaint{}, err
	}
	return requestPaint, nil
}
