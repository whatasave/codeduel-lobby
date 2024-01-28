package utils

import (
	"net/http"
	"strconv"

	"github.com/xedom/codeduel-lobby/codeduel"
)

func GetUser(request *http.Request) *codeduel.User {
	cookie, err := request.Cookie("jwt")
	if err != nil {
		return nil
	}
	// TODO: validate jwt calling codeduel-be
	id, err := strconv.Atoi(cookie.Value)
	if err != nil {
		return nil
	}
	return &codeduel.User{
		Id: int32(id),
	}
}
