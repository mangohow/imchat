package log

import (
	"github.com/mangohow/imchat/cmd/chatserver/internal/conf"
	"github.com/mangohow/imchat/pkg/common/xlogger"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

func InitLogger() (err error) {
	logger, err = xlogger.New(conf.LoggerConf)
	if err != nil {
		return err
	}

	return nil
}

func Logger() *logrus.Logger {
	return logger
}
