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
	RedisConf *xconfig.RedisConfig
	MqConf *xconfig.RabbitMqConfig
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
	initRedisConf()
	initLogConf()
	initMqConf()
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

	}
}

func initRedisConf() {
	RedisConf = &xconfig.RedisConfig{
		Addr:         viper.GetString("redis.addr"),
		Password:     viper.GetString("redis.password"),
		DB:           viper.GetUint32("redis.db"),
		PoolSize:     viper.GetUint32("redis.poolSize"),
		MinIdleConns: viper.GetUint32("redis.minIdleConns"),
	}
}

func initLogConf() {
	LoggerConf = &xconfig.LogConfig{
		Level:       viper.GetString("logger.level"),
		FilePath:    viper.GetString("logger.filePath"),
		FileName:    viper.GetString("logger.fileName"),
		MaxFileSize: viper.GetUint64("logger.maxFileSize"),
		ToFile:      viper.GetBool("logger.toFile"),
		Caller: viper.GetBool("logger.caller"),
	}
}

func initMqConf() {
	MqConf = &xconfig.RabbitMqConfig{
		Host:     viper.GetString("rabbitmq.host"),
		Port:     viper.GetInt("rabbitmq.port"),
		Username: viper.GetString("rabbitmq.username"),
		Password: viper.GetString("rabbitmq.username"),
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