package repository

import(
	"fmt"
	"errors"
	"mime/multipart"
	"context"
	"ssr_immemory/internal/app/api_types"
	"ssr_immemory/internal/app/ds"
	"ssr_immemory/internal/app/minioClient"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)


func (r *Repository) GetPaints() ([]ds.Paint, error) {
	var paints []ds.Paint
	err := r.db.Order("id").Where("is_delete = false").Find(&paints).Error
	if err != nil {
		return nil, err
	}
	if len(paints) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}

	return paints, nil
}

func (r *Repository) GetPaint(id int) (*ds.Paint, error) {
	paint := ds.Paint{}
	err := r.db.Order("id").Where("id = ? and is_delete = ?", id, false).First(&paint).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: краска с id %d", ErrNotFound, id)
		}
		return &ds.Paint{}, err
	}
	return &paint, nil
}
func (r *Repository) GetPaintsByTitle(title string) ([]ds.Paint, error) {
	var paints []ds.Paint
	err := r.db.Order("id").Where("title ILIKE ? and is_delete = ?", "%"+title+"%", false).Find(&paints).Error
	if err != nil {
		return nil, err
	}
	return paints, nil
}

func (r *Repository) CreatePaint(paintJSON apitypes.PaintJSON) (ds.Paint, error) {
	paint := apitypes.PaintFromJSON(paintJSON)
	if paint.HidingPower <= 0 {
		return ds.Paint{}, errors.New("неправильная укрывистость")
	}
	if paint.Title == "" {
		return ds.Paint{}, errors.New("название не может быть пустым")
	}
	err := r.db.Create(&paint).First(&paint).Error
	if err != nil {
		return ds.Paint{}, err
	}
	return paint, nil
}
func (r *Repository) ChangePaint(id int, paintJSON apitypes.PaintJSON) (ds.Paint, error) {
	paint := ds.Paint{}
	if id < 0 {
		return ds.Paint{}, errors.New("id должно быть >= 0")
	}
	err := r.db.Where("id = ? and is_delete = ?", id, false).First(&paint).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.Paint{}, fmt.Errorf("%w: краска с id %d", ErrNotFound, id)
		}
		return ds.Paint{}, err
	}
	if paintJSON.HidingPower <= 0 {
		return ds.Paint{}, errors.New("неправильная укрывистость")
	}
	err = r.db.Model(&paint).Updates(apitypes.PaintFromJSON(paintJSON)).Error
	if err != nil {
		return ds.Paint{}, err
	}
	return paint, nil
}

func (r *Repository) DeletePaint(id int) error {
	paint := ds.Paint{}
	if id < 0 {
		return errors.New("id должно быть >= 0")
	}

	err := r.db.Where("id = ? and is_delete = ?", id, false).First(&paint).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: краска с id %d", ErrNotFound, id)
		}
		return err
	}
	if paint.Photo != "" {
		err = minioClient.DeleteObject(context.Background(), r.mc, minioClient.GetImgBucket(), paint.Photo)
		if err != nil {
			return err
		}
	}

	err = r.db.Model(&ds.Paint{}).Where("id = ?", id).Update("is_delete", true).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) AddPaintToRequest(requestId int, paintId int, area float64, layers int) error {
	var paint ds.Paint
	if err := r.db.First(&paint, paintId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: краска с id %d", ErrNotFound, paintId)
		}
		return err
	}

	var request ds.PaintRequest
	if err := r.db.First(&request, requestId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: заявка с id %d", ErrNotFound, requestId)
		}
		return err
	}
	
	requestPaint := ds.RequestsPaint{}
	result := r.db.Where("paint_id = ? and request_id = ?", paintId, requestId).Find(&requestPaint)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 0 {
		return fmt.Errorf("%w: краска %d уже в заявке %d", ErrAlreadyExists, paintId, requestId)
	}
	
	return r.AddPaintToRequest(requestId, paintId, area, layers)
}

func (r *Repository) GetModeratorAndCreatorLogin(request ds.PaintRequest) (string, string, error) {
	var creator ds.User
	var moderator ds.User

	err := r.db.Where("id = ?", request.CreatorID).First(&creator).Error
	if err != nil {
		return "", "", err
	}

	var moderatorLogin string
	if request.ModeratorID.Valid {
		err = r.db.Where("id = ?", request.ModeratorID.Int64).First(&moderator).Error
		if err != nil {
			return "", "", err
		}
		moderatorLogin = moderator.Login
	}
	
	return creator.Login, moderatorLogin, nil
}

func (r *Repository) UploadImage(ctx *gin.Context, paintId int, file *multipart.FileHeader) (ds.Paint, error) {
	paint_, err := r.GetPaint(paintId)
	if err != nil {
		return ds.Paint{}, err
	}
	fileName, err := minioClient.UploadPaintImage(ctx, r.mc, minioClient.GetImgBucket(), file, *paint_)
	if err != nil {
		return ds.Paint{}, err
	}

	paint, err := r.GetPaint(paintId)
	if err != nil {
		return ds.Paint{}, err
	}
	paint.Photo = fileName
	err = r.db.Save(&paint).Error
	if err != nil {
		return ds.Paint{}, err
	}
	return *paint, nil
}

