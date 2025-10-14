package repository

import (
	apitypes "ssr_immemory/internal/app/api_types"
	"ssr_immemory/internal/app/ds"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

func (r *Repository) GetUserByID(id uint) (ds.User, error) {
	user := ds.User{}
	if id == 0 {
		return ds.User{}, fmt.Errorf("неверный id: должен быть > 0")
	}
	
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		// if errors.Is(err, gorm.ErrRecordNotFound) {
		// 	return ds.User{}, fmt.Errorf("%w: пользователь с id %d не найден", ErrNotFound, id)
		// }
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
	
	// Валидация
	if user.Login == "" {
		return ds.User{}, errors.New("логин обязателен для заполнения")
	}

	if user.Password == "" {
		return ds.User{}, errors.New("пароль обязателен для заполнения")
	}


	// Проверка существования пользователя
	_, err := r.GetUserByLogin(user.Login)
	if err == nil {
		return ds.User{}, fmt.Errorf("%w: пользователь с логином %s уже существует", ErrAlreadyExists, user.Login)
	} else if !errors.Is(err, ErrNotFound) {
		return ds.User{}, err
	}

	// // Проверка прав для создания модератора
	// if user.IsModerator {
	// 	currentUserID := r.GetUserID()
	// 	if currentUserID == 0 {
	// 		return ds.User{}, fmt.Errorf("%w: требуется аутентификация для создания модератора", ErrNotAllowed)
	// 	}
		
	// 	currentUser, err := r.GetUserByID(currentUserID)
	// 	if err != nil {
	// 		return ds.User{}, err
	// 	}
	// 	if !currentUser.IsModerator {
	// 		return ds.User{}, fmt.Errorf("%w: только модераторы могут создавать учетные записи модераторов", ErrNotAllowed)
	// 	}
	// }

	// Создание пользователя
	err = r.db.Create(&user).Error
	if err != nil {
		return ds.User{}, fmt.Errorf("ошибка при создании пользователя: %w", err)
	}
	return user, nil
}

func (r *Repository) SignIn(userJSON apitypes.UserJSON) (ds.User, error) {
	// Валидация
	if userJSON.Login == "" {
		return ds.User{}, errors.New("логин обязателен для заполнения")
	}
	if userJSON.Password == "" {
		return ds.User{}, errors.New("пароль обязателен для заполнения")
	}

	// Поиск пользователя
	user, err := r.GetUserByLogin(userJSON.Login)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ds.User{}, errors.New("неверный логин или пароль")
		}
		return ds.User{}, err
	}

	// Проверка пароля
	if user.Password != userJSON.Password {
		return ds.User{}, errors.New("неверный логин или пароль")
	}

	// Установка ID пользователя в сессии
	r.SetUserID(user.ID)
	return user, nil
}

func (r *Repository) UpdateUserProfile(id uint, userJSON apitypes.UserJSON) (ds.User, error) {
	if id == 0 {
		return ds.User{}, fmt.Errorf("неверный id пользователя")
	}

	// Получение текущего пользователя
	currentUser, err := r.GetUserByID(id)
	if err != nil {
		return ds.User{}, err
	}

	// Подготовка обновлений
	updates := apitypes.UserFromJSON(userJSON)
	
	// Проверка прав на изменение роли модератора
	if updates.IsModerator && !currentUser.IsModerator {
		// Обычный пользователь не может сделать себя модератором
		updates.IsModerator = false
	}

	// Обновление
	err = r.db.Model(&currentUser).Updates(updates).Error
	if err != nil {
		return ds.User{}, fmt.Errorf("ошибка при обновлении профиля: %w", err)
	}

	// Возвращаем обновленного пользователя
	return r.GetUserByID(id)
}