package apitypes

import (
	"ssr_immemory/internal/app/ds"
	"github.com/google/uuid"
)

type UserJSON struct {
	ID          uuid.UUID `json:"id"`
	Login       string    `json:"login"`
	Password    string    `json:"password"`
	IsModerator bool      `json:"is_moderator"`
}

func UserToJSON(user ds.User) UserJSON {
	return UserJSON{
		ID:          user.ID,
		Login:       user.Login,
		Password:    user.Password,
		IsModerator: user.IsModerator,
	}
}

func UserFromJSON(userJSON UserJSON) ds.User {
	return ds.User{
		Login:       userJSON.Login,
		Password:    userJSON.Password,
		IsModerator: userJSON.IsModerator,
	}
}