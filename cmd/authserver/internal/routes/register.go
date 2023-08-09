package routes

import (
	"github.com/mangohow/easygin"
	"github.com/mangohow/imchat/cmd/authserver/internal/controller"
	"github.com/mangohow/imchat/cmd/authserver/internal/middleware"
)

func Register(router *easygin.EasyGin) {
	group := router.Group("/api")
	userController := controller.NewUserController()
	group.GET("/phoneCode", userController.GetPhoneVerificationCode)
	group.POST("/userRegister", userController.Register)
	group.GET("/login", userController.Login)
	group.GET("/userinfo", userController.GetUserInfo)

	authedGroup := router.Group("/api/auth")
	authedGroup.Use(middleware.Authentication())
	authedGroup.GET("/selfinfo", userController.GetSelfInfo)

	friendController := controller.NewFriendController()
	authedGroup.GET("/friends", friendController.GetAllFriendsInfo)
	authedGroup.GET("/onlineFriends", friendController.GetOnlineFriends)
}

