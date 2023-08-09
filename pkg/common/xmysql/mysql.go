package xmysql

import (
	"github.com/mangohow/imchat/pkg/common/xconfig"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func NewMysqlInstance(conf *xconfig.MysqlConfig) (db *gorm.DB, err error) {
	db, err = gorm.Open(mysql.Open(conf.DataSourceName), &gorm.Config{
		SkipDefaultTransaction: true,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:         "t_",
			SingularTable: true,
		},
	})
	if err != nil {
		return nil, err
	}
	sqlDb, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDb.SetMaxOpenConns(conf.MaxOpenConns)
	sqlDb.SetMaxIdleConns(conf.MaxIdleConns)
	if err = sqlDb.Ping(); err != nil {
		return db, nil
	}

	return db, nil
}