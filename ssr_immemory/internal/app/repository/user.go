package repository

import (
	"errors"
	"fmt"
	"os"
	"time"
	apitypes "ssr_immemory/internal/app/api_types"
	"ssr_immemory/internal/app/ds"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (r *Repository) GetUserByID(id uuid.UUID) (ds.User, error) {
	user := ds.User{}
	if id == uuid.Nil {
		return ds.User{}, fmt.Errorf("неверный id: не может быть пустым")
	}
	
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.User{}, fmt.Errorf("%w: пользователь с id %s не найден", ErrNotFound, id.String())
		}
		return ds.User{}, err
	}
	return user, nil
}

func (r *Repository) GetUserByLogin(login string) (ds.User, error) {
	user := ds.User{}
	if login == "" {
		return ds.User{}, errors.New("логин не может быть пустым")
	}
	
	err := r.db.Where("login = ?", login).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.User{}, fmt.Errorf("%w: пользователь с логином %s не найден", ErrNotFound, login)
		}
		return ds.User{}, err
	}
	return user, nil
}

func (r *Repository) CreateUser(userJSON apitypes.UserJSON) (ds.User, error) {
	user := apitypes.UserFromJSON(userJSON)
	
	if user.Login == "" {
		return ds.User{}, errors.New("логин обязателен для заполнения")
	}

	if user.Password == "" {
		return ds.User{}, errors.New("пароль обязателен для заполнения")
	}
	_, err := r.GetUserByLogin(user.Login)
	if err == nil {
		return ds.User{}, fmt.Errorf("%w: пользователь с логином %s уже существует", ErrAlreadyExists, user.Login)
	} else if !errors.Is(err, ErrNotFound) {
		return ds.User{}, err
	}
	err = r.db.Create(&user).Error
	if err != nil {
		return ds.User{}, fmt.Errorf("ошибка при создании пользователя: %w", err)
	}
	return user, nil
}

func (r *Repository) SignIn(userJSON apitypes.UserJSON) (string, error) {
	if userJSON.Login == "" {
		return "", errors.New("логин обязателен для заполнения")
	}
	if userJSON.Password == "" {
		return "", errors.New("пароль обязателен для заполнения")
	}

	user, err := r.GetUserByLogin(userJSON.Login)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", errors.New("неверный логин или пароль")
		}
		return "", err
	}

	if user.Password != userJSON.Password {
		return "", errors.New("неверный логин или пароль")
	}
	token, err := GenerateToken(user.ID, user.IsModerator)
	if err != nil {
		return "", err
	}

	return token, nil
}

func GenerateToken(userID uuid.UUID, isModerator bool) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["authorized"] = true
	claims["user_id"] = userID.String() 
	claims["is_moderator"] = isModerator
	claims["exp"] = time.Now().Add(time.Hour * 1).Unix()

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_KEY")))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (r *Repository) UpdateUserProfile(id uuid.UUID, userJSON apitypes.UserJSON) (ds.User, error) {
	if id == uuid.Nil {
		return ds.User{}, fmt.Errorf("неверный id пользователя")
	}
	currentUser, err := r.GetUserByID(id)
	if err != nil {
		return ds.User{}, err
	}
	updates := apitypes.UserFromJSON(userJSON)
	
	if updates.IsModerator && !currentUser.IsModerator {
		updates.IsModerator = false
	}

	err = r.db.Model(&currentUser).Updates(map[string]interface{}{
		"login":        updates.Login,
		"password":     updates.Password,
		"is_moderator": updates.IsModerator,
	}).Error
	
	if err != nil {
		return ds.User{}, fmt.Errorf("ошибка при обновлении профиля: %w", err)
	}

	return r.GetUserByID(id)
}