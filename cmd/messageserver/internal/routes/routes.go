package routes

import (
	"github.com/mangohow/easygin"
	"github.com/mangohow/imchat/cmd/messageserver/internal/controller"
	"github.com/mangohow/imchat/cmd/messageserver/internal/middleware"
)

func Register(engine *easygin.EasyGin) {
	engine.Use(middleware.Authentication())
	group := engine.Group("/api/message")
	messageController := controller.NewChatMessageController()
	group.GET("/offline", messageController.PullOfflineMessages)
	group.GET("/history", messageController.GetMessages)
	group.PUT("/status", messageController.UpdateStatus)
}