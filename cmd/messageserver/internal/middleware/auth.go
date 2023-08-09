package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mangohow/easygin"
	"github.com/mangohow/imchat/cmd/messageserver/internal/log"
	"github.com/mangohow/imchat/pkg/utils"
)

func Authentication() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.GetHeader("authorization")
		if token == "" {
			log.Logger().Warnf("未获得授权, ip:%s", ctx.Request.RemoteAddr)
			ctx.JSON(http.StatusUnauthorized, &easygin.Error(http.StatusUnauthorized, -1).R)
			ctx.Abort()
			return
		}

		if id, username, err := utils.ParseToken(token); err != nil {
			ctx.JSON(http.StatusUnauthorized, &easygin.Error(http.StatusUnauthorized, -1).R)
			ctx.Abort()
			log.Logger().Warnf("未获得授权, ip:%s", ctx.Request.RemoteAddr)
			return
		} else {
			ctx.Set("username", username)
			ctx.Set("id", id)
			ctx.Set("token", token)
		}

		ctx.Next()
	}
}

