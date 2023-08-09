package client

import "github.com/mangohow/imchat/pkg/model"

type LoginData struct {
	Token string `json:"token"`
	User model.User `json:"user"`
	Userinfo model.Userinfo `json:"userinfo"`
}

type Response[T any] struct {
	Code uint32 `json:"code"`
	Message string `json:"message"`
	Data T `json:"data"`
}
