package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
	apitypes "ssr_immemory/internal/app/api_types"
	"ssr_immemory/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

// CreateUser godoc
// @Summary Регистрация пользователя
// @Description Регистрирует нового пользователя
// @Tags users
// @Accept json
// @Produce json
// @Param user body apitypes.UserJSON true "Данные пользователя"
// @Success 201 {object} apitypes.UserJSON
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /users/register [post]
func (h *Handler) CreateUser(ctx *gin.Context) {
	var userJSON apitypes.UserJSON
	if err := ctx.BindJSON(&userJSON); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Валидация обязательных полей
	if userJSON.Login == "" {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("field 'login' is required"))
		return
	}
	if userJSON.Password == "" {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("field 'password' is required"))
		return
	}

	user, err := h.Repository.CreateUser(userJSON)
	if err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			h.errorHandler(ctx, http.StatusConflict, err)
		} else if errors.Is(err, repository.ErrNotAllowed) {
			h.errorHandler(ctx, http.StatusForbidden, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.Header("Location", fmt.Sprintf("/users/%v", user.ID))
	ctx.JSON(http.StatusCreated, apitypes.UserToJSON(user))
}

// SignIn godoc
// @Summary Вход в систему
// @Description Аутентификация пользователя и получение JWT токена
// @Tags users
// @Accept json
// @Produce json
// @Param credentials body apitypes.UserJSON true "Логин и пароль"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /users/signin [post]
func (h *Handler) SignIn(ctx *gin.Context) {
	var userJSON apitypes.UserJSON
	if err := ctx.BindJSON(&userJSON); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if userJSON.Login == "" {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("field 'login' is required"))
		return
	}
	if userJSON.Password == "" {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("field 'password' is required"))
		return
	}

	token, err := h.Repository.SignIn(userJSON)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("invalid login or password"))
		} else if err.Error() == "wrong password" {
			h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("invalid login or password"))
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

// GetProfile godoc
// @Summary Получить профиль
// @Description Получить данные текущего пользователя
// @Tags users
// @Produce json
// @Success 200 {object} apitypes.UserJSON
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security ApiKeyAuth
// @Router /users/profile [get]
func (h *Handler) GetProfile(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}

	user, err := h.Repository.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	// Не возвращаем пароль
	user.Password = ""
	ctx.JSON(http.StatusOK, apitypes.UserToJSON(user))
}

// ChangeProfile godoc
// @Summary Изменить профиль
// @Description Обновить данные текущего пользователя
// @Tags users
// @Accept json
// @Produce json
// @Param user body apitypes.UserJSON true "Новые данные пользователя"
// @Success 200 {object} apitypes.UserJSON
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security ApiKeyAuth
// @Router /users/profile [put]
func (h *Handler) ChangeProfile(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}

	var userJSON apitypes.UserJSON
	if err := ctx.BindJSON(&userJSON); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	user, err := h.Repository.UpdateUserProfile(userID, userJSON)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	// Не возвращаем пароль
	user.Password = ""
	ctx.JSON(http.StatusOK, apitypes.UserToJSON(user))
}

// SignOut godoc
// @Summary Выход из системы
// @Description Добавляет токен в черный список
// @Tags users
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /users/signout [post]
func (h *Handler) SignOut(ctx *gin.Context) {
	tokenString := extractTokenFromHeader(ctx.Request)
	if tokenString == "" {
		h.errorHandler(ctx, http.StatusUnauthorized, errors.New("no token provided"))
		return
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_KEY")), nil
	})

	if err != nil || token == nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("invalid token claims"))
		return
	}

	ttl, err := getTokenTTLFromClaims(claims)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"status": "signed_out"})
		return
	}

	err = h.Repository.AddTokenToBlacklist(context.Background(), tokenString, ttl)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "signed_out"})
}


func getUserIDFromContext(ctx *gin.Context) (uuid.UUID, error) {
	userIDStr, exists := ctx.Get("user_id")
	if !exists {
		return uuid.Nil, errors.New("user_id not found in context")
	}
	
	userIDStrValue, ok := userIDStr.(string)
	if !ok {
		return uuid.Nil, errors.New("user_id is not a string")
	}
	
	userID, err := uuid.Parse(userIDStrValue)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user_id format: %v", err)
	}
	
	return userID, nil
}

func getTokenTTLFromClaims(claims jwt.MapClaims) (time.Duration, error) {
	expVal, ok := claims["exp"]
	if !ok {
		return 0, errors.New("exp not present in token")
	}

	var expUnix int64
	switch v := expVal.(type) {
	case float64:
		expUnix = int64(v)
	case int64:
		expUnix = v
	case json.Number:
		i, err := v.Int64()
		if err != nil {
			return 0, err
		}
		expUnix = i
	default:
		return 0, fmt.Errorf("unsupported exp type %T", v)
	}

	expTime := time.Unix(expUnix, 0)
	ttl := time.Until(expTime)
	if ttl < 0 {
		return 0, errors.New("token already expired")
	}
	return ttl, nil
}
