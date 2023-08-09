package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mangohow/easygin"
	"github.com/mangohow/imchat/cmd/authserver/internal/log"
	"github.com/mangohow/imchat/cmd/authserver/internal/resultcode"
	"github.com/mangohow/imchat/cmd/authserver/internal/service"
	"github.com/sirupsen/logrus"
)

/*

	FriendController 用户联系人controller
*/
type FriendController struct {
	friendService *service.FriendService
	logger         *logrus.Logger
}

func NewFriendController() *FriendController {
	return &FriendController{
		friendService: service.NewContactFriendService(),
		logger:         log.Logger(),
	}
}

// GetAllFriendsInfo 获取所有联系人的信息
// GET /api/auth/friends
func (c *FriendController) GetAllFriendsInfo(ctx *gin.Context) *easygin.Result {
	value, exists := ctx.Get("id")
	if !exists {
		return easygin.Error(http.StatusUnauthorized, -1)
	}
	id := value.(int64)
	friends, err := c.friendService.GetAllFriends(id)
	if err != nil {
		return easygin.Fail(resultcode.QueryFailed)
	}

	return easygin.Ok(friends)
}

// GetOnlineFriends 查询用户在线状态, 返回在线的用户ID
func (c *FriendController) GetOnlineFriends(ctx *gin.Context) *easygin.Result {
	value, exists := ctx.Get("id")
	if !exists {
		return easygin.Error(http.StatusUnauthorized, -1)
	}
	id := value.(int64)
	onlineFriends := c.friendService.GetOnlineFriends(id)
	return easygin.Ok(onlineFriends)
}
