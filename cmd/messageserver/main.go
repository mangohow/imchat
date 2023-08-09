package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mangohow/easygin"
	"github.com/mangohow/imchat/cmd/messageserver/internal/conf"
	"github.com/mangohow/imchat/cmd/messageserver/internal/log"
	"github.com/mangohow/imchat/cmd/messageserver/internal/mongodb"
	"github.com/mangohow/imchat/cmd/messageserver/internal/routes"
)

func main() {
	confPath := flag.String("conf", "conf/messageserver.yaml", "specify config path")
	flag.Parse()
	if err := conf.LoadConf(*confPath); err != nil {
		panic(fmt.Errorf("load conf failed, reason:%s", err.Error()))
	}

	if err := log.InitLogger(); err != nil {
		panic(fmt.Errorf("init log error, reason:%v", err))
	}

	if err := mongodb.InitMongoDB(); err != nil {
		panic(fmt.Errorf("init mongodb error:%v", err))
	}


	// 创建gin路由
	easyGin := easygin.NewWithEngine(gin.Default())
	easygin.SetLogOutput(log.Logger().Out)

	// 注册路由
	routes.Register(easyGin)

	err := easyGin.ListenAndServe(fmt.Sprintf("%s:%d", conf.ServerConf.Host, conf.ServerConf.Port))
	if err != nil {
		if err == http.ErrServerClosed {
			fmt.Println("server closed")
		}

		fmt.Println(err)
	}
}
