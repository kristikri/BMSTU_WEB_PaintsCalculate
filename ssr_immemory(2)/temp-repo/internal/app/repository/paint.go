package repository

import(
	"strings"
	"ssr_immemory/internal/app/ds"
)


func (r *Repository) GetPaints() ([]ds.Paint, error) {
	var paints []ds.Paint
	err := r.db.Find(&paints).Error
	if err != nil {
		return nil, err
	}
	return paints, nil
}

func (r *Repository) GetPaint(id int) (ds.Paint, error) {
	var paint ds.Paint
	err := r.db.Where("id = ? AND is_delete = false", id).First(&paint).Error
	if err != nil {
		return ds.Paint{}, err
	}
	return paint, nil
}

func (r *Repository) GetPaintsByTitle(title string) ([]ds.Paint, error) {
	var paints []ds.Paint
	err := r.db.Where("title ILIKE ? AND is_delete = false", "%"+title+"%").Find(&paints).Error
	if err != nil {
		return nil, err
	}
	return paints, nil
}

func (r *Repository) GetOrCreateDraftRequest(userID uint) (ds.PaintRequest, error) {
	var paint_request ds.PaintRequest
	err := r.db.Where("creator_id = ? AND status = ?", userID, "черновик").First(&paint_request).Error
	
	if err != nil {
		paint_request = ds.PaintRequest{
			CreatorID:  userID,
			Status:     "черновик",
		}
		err = r.db.Create(&paint_request).Error
		if err != nil {
			return ds.PaintRequest{}, err
		}
	}
	
	return paint_request, nil
}

func (r *Repository) AddPaintToRequest(paint_requestID uint, paintID uint, area float64, layers int) error {
	var paint ds.Paint
	err := r.db.First(&paint, paintID).Error
	if err != nil {
		return err
	}

	quantity := area * paint.HidingPower * float64(layers) / 1000

	requestPaint := ds.RequestPaint{
		RequestID: paint_requestID,
		PaintID:   paintID,
		Area:      area,
		Layers:    layers,
		Quantity:  quantity,
	}

	err = r.db.Create(&requestPaint).Error
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return r.db.Model(&ds.RequestPaint{}).
				Where("paint_request_id = ? AND paint_id = ?", paint_requestID, paintID).
				Updates(map[string]interface{}{
					"area":     area,
					"layers":   layers,
					"quantity": quantity,
				}).Error
		}
		return err
	}

	return r.UpdateRequestTotals(paint_requestID)
}

func (r *Repository) UpdateRequestTotals(paint_requestID uint) error {
	var totalArea, totalPaints float64
	
	row := r.db.Model(&ds.RequestPaint{}).
		Where("request_id = ?", paint_requestID).
		Select("COALESCE(SUM(area), 0), COALESCE(SUM(quantity), 0)").
		Row()
	
	err := row.Scan(&totalArea, &totalPaints)
	if err != nil {
		return err
	}

	return r.db.Model(&ds.PaintRequest{}).
		Where("id = ?", paint_requestID).
		Updates(map[string]interface{}{
			"total_area":   totalArea,
			"total_paints": totalPaints,
		}).Error
}

func (r *Repository) GetRequestWithPaints(paint_requestID uint) (ds.PaintRequest, error) {
	var request ds.PaintRequest
	err := r.db.Preload("RequestPaints.Paint").First(&request, paint_requestID).Error
	return request, err
}

func (r *Repository) DeleteRequestSQL(paint_requestID uint) error {
    err := r.db.Exec("DELETE FROM request_paints WHERE request_id = ?", paint_requestID).Error
    if err != nil {
        return err
    }
    return r.db.Exec("UPDATE paint_requests SET status = 'удалён', date_finish = NOW() WHERE id = ?", paint_requestID).Error
}

func (r *Repository) GetPaintCount(userID uint) int64 {
	var count int64
	var paint_request ds.PaintRequest
	
	err := r.db.Where("creator_id = ? AND status = ?", userID, "черновик").First(&paint_request).Error
	if err != nil {
		return 0
	}
	
	r.db.Model(&ds.RequestPaint{}).Where("paint_request = ?", paint_request.ID).Count(&count)
	return count
}
func (r *Repository) GetDraftRequest(userID uint) (ds.PaintRequest, error) {
	var paint_request ds.PaintRequest
	err := r.db.Where("creator_id = ? AND status = ?", userID, "черновик").First(&paint_request).Error
	return paint_request, err
}

