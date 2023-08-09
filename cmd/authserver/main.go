package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mangohow/easygin"
	"github.com/mangohow/imchat/cmd/authserver/internal/conf"
	"github.com/mangohow/imchat/cmd/authserver/internal/dao"
	"github.com/mangohow/imchat/cmd/authserver/internal/log"
	"github.com/mangohow/imchat/cmd/authserver/internal/rdsconn"
	"github.com/mangohow/imchat/cmd/authserver/internal/routes"
)

func main() {
	confPath := flag.String("conf", "conf/authserver.yaml", "specify config path")
	flag.Parse()
	if err := conf.LoadConf(*confPath); err != nil {
		panic(fmt.Errorf("load conf failed, reason:%s", err.Error()))
	}

	if err := log.InitLogger(); err != nil {
		panic(fmt.Errorf("init log error, reason:%v", err))
	}

	// 初始化mysql
	if err := dao.InitMysql(); err != nil {
		panic(fmt.Errorf("init mysql failed, reason:%s", err.Error()))
	}

	// 初始化redis
	if err := rdsconn.InitRedis(); err != nil {
		panic(fmt.Errorf("init rdsconn failed, reason:%s", err.Error()))
	}

	// 创建gin路由
	easyGin := easygin.NewWithEngine(gin.Default())
	easygin.SetLogOutput(log.Logger().Out)

	// 注册路由
	routes.Register(easyGin)

	easyGin.SetAfterCloseHandlers(func() {
		dao.CloseMysql()
		rdsconn.CloseRedis()
	})

	err := easyGin.ListenAndServe(fmt.Sprintf("%s:%d", conf.ServerConf.Host, conf.ServerConf.Port))
	if err != nil {
		if err == http.ErrServerClosed {
			fmt.Println("server closed")
		}

		fmt.Println(err)
	}
}


