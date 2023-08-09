package dao

import (
	"github.com/mangohow/imchat/cmd/authserver/internal/conf"
	"github.com/mangohow/imchat/pkg/common/xmysql"
	"gorm.io/gorm"
)

var mysqlDB *gorm.DB

func InitMysql() (err error) {
	mysqlDB, err = xmysql.NewMysqlInstance(conf.MysqlConf)
	if conf.ServerConf.Mode == "dev" {
		mysqlDB = mysqlDB.Debug()
	}
	return
}

func CloseMysql() error {
	db, err := mysqlDB.DB()
	if err != nil {
		return err
	}
	return db.Close()
}

func MysqlDB() *gorm.DB{
	return mysqlDB
}
