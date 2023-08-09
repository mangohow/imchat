package log

import (
	"github.com/mangohow/imchat/cmd/authserver/internal/conf"
	"github.com/mangohow/imchat/pkg/common/xlogger"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

func Logger() *logrus.Logger {
	return logger
}

func InitLogger() error {
	var err error
	if logger, err = xlogger.New(conf.LoggerConf); err != nil {
		return err
	}

	return nil
}
