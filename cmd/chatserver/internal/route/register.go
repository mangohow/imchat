package route

import (
	"fmt"

	"github.com/mangohow/imchat/cmd/chatserver/internal/chatserver"
	"github.com/mangohow/imchat/cmd/chatserver/internal/conf"
	"github.com/mangohow/imchat/cmd/chatserver/internal/handlers"
	"github.com/mangohow/imchat/pkg/consts"
	"github.com/mangohow/imchat/proto/pb"
)

func Register(s *chatserver.ChatServer) {
	if conf.ServerConf.Mode != "test" {
		authHandler := handlers.NewAuthHandler(s.HeartBeat(), s.ServerId())
		// 设置权限验证处理器, 在握手阶段需要在header中传入token
		s.SetAfterHandshakeHandler(authHandler.Auth)

		// 清理数据
		s.SetClientCloseHandler(authHandler.ClientCloseHandler)

		// 权限验证
		s.Use(authHandler.CheckAuthMiddleware)
	}

	s.HandlerAnyFunc(consts.HelloRequest, func(ctx *chatserver.Context, hello *pb.Hello) *pb.Hello {
		fmt.Println(hello.Message)
		ctx.SetRespId(consts.HelloReply)
		return &pb.Hello{Message: "hello"}
	})

	retryHandler := handlers.NewRetryHandler(s.GetCtx(), 8, 5)

	userChatHandler := handlers.NewUserChatHandler(s.GetCtx(), 8, retryHandler)
	s.HandlerAnyFunc(consts.SingleChatMessage, userChatHandler.ForwardMessage)
	s.HandlerAnyFunc(consts.SingleChatAck, userChatHandler.ConfirmMessage)
}
