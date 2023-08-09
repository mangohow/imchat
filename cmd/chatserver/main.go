package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/mangohow/easygin"
	"github.com/mangohow/imchat/cmd/chatserver/internal/chatserver"
	"github.com/mangohow/imchat/cmd/chatserver/internal/conf"
	"github.com/mangohow/imchat/cmd/chatserver/internal/consts"
	"github.com/mangohow/imchat/cmd/chatserver/internal/log"
	"github.com/mangohow/imchat/cmd/chatserver/internal/mongodb"
	"github.com/mangohow/imchat/cmd/chatserver/internal/mq"
	"github.com/mangohow/imchat/cmd/chatserver/internal/rdsconn"
	"github.com/mangohow/imchat/cmd/chatserver/internal/route"
	"github.com/mangohow/imchat/pkg/consts/redisconsts"
)

func main() {
	confPath := flag.String("conf", "conf/chatserver.yaml", "specify config path")
	if err := conf.LoadConf(*confPath); err != nil {
		panic(fmt.Errorf("load conf error:%v", err))
	}

	if err := log.InitLogger(); err != nil {
		panic(fmt.Errorf("init logger error:%v", err))
	}

	port := flag.Int("port", 0, "specify server listening port")
	host := flag.String("host", "", "specify server listening host")
	id := flag.Int("id", 0, "specify server id")
	flag.Parse()
	if *port != 0 {
		conf.ServerConf.Port = *port
	}
	if *host != "" {
		conf.ServerConf.Host = *host
	}
	if *id != 0 {
		conf.ServerConf.NodeId = *id
	}

	if err := rdsconn.InitRedis(); err != nil {
		panic(fmt.Errorf("init redis error:%v", err))
	}

	if err := mongodb.InitMongoDB(); err != nil {
		panic(fmt.Errorf("init mongodb error:%v", err))
	}

	server := chatserver.NewServer(&chatserver.Config{
		Addr:      fmt.Sprintf("%s:%d", conf.ServerConf.Host, conf.ServerConf.Port),
		HeartBeat: consts.HeartBeatTime,
	})

	// 初始化消息队列
	if err := mq.InitMQ(redisconsts.ServerConsumerKey + server.ServerId(), redisconsts.ServerConsumerKey + server.ServerId()); err != nil {
		panic(fmt.Errorf("init mq error:%v", err))
	}

	server.RegisterOnShutdown(func() {
		// 关闭server前清空所有连接
		chatserver.ClientManagerInstance.Clear()
	})

	route.Register(server)

	go func() {
		if err := server.Serve(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	// 注册信号, 关闭服务
	easygin.SetupSignal(func() {
		err := server.Shutdown()
		if err != nil {
			log.Logger().Error("server closed", err)
		} else {
			log.Logger().Info("server closed")
		}
	})

}

