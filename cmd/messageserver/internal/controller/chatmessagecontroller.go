package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mangohow/easygin"
	"github.com/mangohow/imchat/cmd/messageserver/internal/log"
	"github.com/mangohow/imchat/cmd/messageserver/internal/resultcode"
	"github.com/mangohow/imchat/cmd/messageserver/internal/service"
	"github.com/sirupsen/logrus"
)

type ChatMessageController struct {
	logger *logrus.Logger
	chatMessageService *service.ChatMessageService
}

func NewChatMessageController() *ChatMessageController {
	return &ChatMessageController{
		logger: log.Logger(),
		chatMessageService: service.NewChatMessageService(),
	}
}

func getId(ctx *gin.Context) int64 {
	value, exists := ctx.Get("id")
	if !exists {
		return -1
	}
	return value.(int64)
}

// PullOfflineMessages 拉取离线消息
// TODO 目前是拉取所有，处理当数据量非常大时的情况
func (c *ChatMessageController) PullOfflineMessages(ctx *gin.Context) *easygin.Result {
	id := getId(ctx)
	if id == -1 {
		return easygin.Error(http.StatusUnauthorized, resultcode.Unauthorized)
	}

	records, err := c.chatMessageService.GetOfflineMessage(id)
	if err != nil {
		c.logger.Errorf("get offline message error:%v", err)
		return easygin.Fail(resultcode.QueryFailed)
	}

	return easygin.Ok(records)
}

// GetMessages 获取消息, createTime为下一条消息的创建时间, 设置为-1则从最新的开始拉取
func (c *ChatMessageController) GetMessages(ctx *gin.Context, friendId int64, pageSize int, createTime int64) *easygin.Result {
	id := getId(ctx)
	if id == -1 {
		return easygin.Error(http.StatusUnauthorized, resultcode.Unauthorized)
	}
	records, err := c.chatMessageService.GetMessage(id, friendId, pageSize, createTime)
	if err != nil {
		c.logger.Errorf("get message error:%v", err)
		return easygin.Fail(resultcode.QueryFailed)
	}

	return easygin.Ok(records)
}

// UpdateStatus 更新消息状态
func (c *ChatMessageController) UpdateStatus(ctx *gin.Context) *easygin.Result {
	id := getId(ctx)
	if id == -1 {
		return easygin.Error(http.StatusUnauthorized, resultcode.Unauthorized)
	}

	var messageIds []string
	err := ctx.BindJSON(&messageIds)
	if err != nil {
		c.logger.Errorf("bind message ids error:%v", err)
		return easygin.Error(http.StatusInternalServerError, resultcode.QueryFailed)
	}
	if len(messageIds) > 1 {
		err =c.chatMessageService.UpdateMany(id, messageIds)
	} else if len(messageIds) == 1 {
		err = c.chatMessageService.UpdateOne(id, messageIds[0])
	}

	if err != nil {
		return easygin.Fail(resultcode.UpdateMessageFailed)
	}

	return easygin.Ok(nil)
}
