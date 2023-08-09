package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/mangohow/imchat/cmd/chatserver/internal/chatserver"
	"github.com/mangohow/imchat/cmd/chatserver/internal/log"
	"github.com/mangohow/imchat/cmd/chatserver/internal/rdsconn"
	"github.com/mangohow/imchat/pkg/consts/redisconsts"
	"github.com/mangohow/imchat/pkg/utils"
	"github.com/sirupsen/logrus"
)

type AuthHandler struct {
	logger *logrus.Logger
	redis *redis.Client
	heartBeat time.Duration
	serverId string
}

func NewAuthHandler(heartBeat time.Duration, serverId string) *AuthHandler {
	return &AuthHandler{
		logger: log.Logger(),
		redis: rdsconn.RedisConn(),
		serverId: serverId,
		heartBeat: heartBeat,
	}
}

var UnauthorizedMessage = []byte("unauthorized")

func (h *AuthHandler) Auth(r *http.Request, conn *chatserver.Client) bool {
	h.logger.Debug("auth handler")
	token := r.Header.Get("authorization")
	if token == "" {
		h.logger.Errorf("no authorization, addr:%s", r.RemoteAddr)
		return false
	}

	for {
		id, username, err := utils.ParseToken(token)
		if err != nil {
			break
		}
		// 从redis中查询
		key := redisconsts.TokenKey + strconv.Itoa(int(id))
		result, err := h.redis.Get(context.Background(), key).Result()
		if err != nil {
			if err == redis.Nil {
				break
			}
			h.logger.Errorf("get token error:%v", err)
			break
		}

		if result != token {
			break
		}

		// 验证通过 保存客户端信息到redis
		conn.Set("id", id)
		conn.Set("username", username)

		clientKey := redisconsts.ChatServerClientKey + strconv.Itoa(int(id))
		setVal := h.serverId
		s, err := h.redis.GetSet(context.Background(), clientKey, setVal).Result()
		if err != nil && err != redis.Nil {
			h.logger.Errorf("set client key error:%v", err)
			return false
		}

		// 判断是否重复登录, 如果重复登录，就让另一端下线
		// 如果s != "" 则说明该用户已经登录, 因此不能重复登录
		if s != "" {
			h.logger.Debug("duplicate login:", s)
			h.handleDuplicateLogin(id, s)
		}

		conn.SetAuthed()

		// 添加到manager中
		chatserver.ClientManagerInstance.Add(id, conn)

		return true
	}

	_ = conn.WriteControl(websocket.CloseMessage, UnauthorizedMessage, time.Now().Add(h.heartBeat))
	return false
}

// CheckAuthMiddleware 中间件，权限判断
func (h *AuthHandler) CheckAuthMiddleware(ctx *chatserver.Context) {
	if !ctx.Client.Authed() {
		log.Logger().Info("no auth, close client, ip: ", ctx.RemoteAddr())
		ctx.Client.Close()
		ctx.Abort()
		return
	}
}

// TODO
func (h *AuthHandler) handleDuplicateLogin(uid int64, serverId string) {

}

func (h *AuthHandler) ClientCloseHandler(conn *chatserver.Client) {
	// 将key从redis删除
	val, exist := conn.Get("id")
	if !exist {
		return
	}
	id := val.(int64)
	clientKey := redisconsts.ChatServerClientKey + strconv.Itoa(int(id))
	h.redis.Del(context.Background(), clientKey)

	// 将conn从manager中删除
	chatserver.ClientManagerInstance.Del(id)
}

