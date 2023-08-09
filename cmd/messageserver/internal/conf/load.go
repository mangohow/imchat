package conf

import (
	"path/filepath"
	"strings"

	"github.com/mangohow/imchat/pkg/common/xconfig"
	"github.com/spf13/viper"
)

var (
	ServerConf *xconfig.ServerConfig
	LoggerConf *xconfig.LogConfig
	MongoConf *xconfig.MongoConfig
)

func LoadConf(path string) error {
	dir, file := filepath.Split(path)
	split := strings.Split(file, ".")
	name, ext := split[0], split[1]

	viper.SetConfigName(name)
	viper.SetConfigType(ext)
	viper.AddConfigPath(dir)

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	setDefault()

	initServerConf()
	initLogConf()
	initMongoConf()

	return nil
}

func setDefault() {

}

func initServerConf() {
	ServerConf = &xconfig.ServerConfig{
		Host: viper.GetString("server.host"),
		Port: viper.GetInt("server.port"),
		Name: viper.GetString("server.name"),
		Mode: viper.GetString("server.mode"),
		NodeId: viper.GetInt("server.nodeId"),
	}
}

func initLogConf() {
	LoggerConf = &xconfig.LogConfig{
		Level:       viper.GetString("log.level"),
		FilePath:    viper.GetString("log.filePath"),
		FileName:    viper.GetString("log.fileName"),
		MaxFileSize: viper.GetUint64("log.maxFileSize"),
		ToFile:      viper.GetBool("log.toFile"),
		Caller: viper.GetBool("log.caller"),
	}
}

func initMongoConf() {
	MongoConf = &xconfig.MongoConfig{
		Url:         viper.GetString("mongo.url"),
		Db:          viper.GetString("mongo.db"),
		MaxPoolSize: viper.GetInt("mongo.maxPoolSize"),
		MinPoolSize: viper.GetInt("mongo.minPoolSize"),
	}
}
