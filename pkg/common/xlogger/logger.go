package xlogger

import (
	"bytes"
	"fmt"
	"github.com/mangohow/imchat/pkg/common/xconfig"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
)

const TimeFormat = "2006-01-02 15:04:05"

func New(config *xconfig.LogConfig) (*logrus.Logger, error) {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	formatterName := strings.ToLower(config.Formatter)
	var formatter logrus.Formatter
	switch formatterName {
	case "json":
		formatter = &logrus.JSONFormatter{}
	case "text":
		formatter = &logrus.TextFormatter{}
	default:
		formatter = &LogFormatter{}
	}
	logger.SetFormatter(formatter)

	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		fmt.Printf("parse log level error:%v", err)
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)
	logger.SetReportCaller(config.Caller)

	if config.ToFile {
		w, err := OpenLogFile(config.FilePath, config.FileName)
		if err != nil {
			return nil, fmt.Errorf("open file error:%s", err.Error())
		}
		logger.SetOutput(w)
	}

	return logger, nil
}

func OpenLogFile(filePath, fileName string) (io.Writer, error) {
	if !strings.HasSuffix(fileName, ".log") {
		fileName = fileName + ".log"
	}
	filep := path.Join(filePath, fileName)

	err := os.Mkdir(filePath, 0644)
	if err != nil && !os.IsExist(err) {
		return nil, err
	}
	file, err := os.OpenFile(filep, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return file, nil
}

type LogFormatter struct{}

func (t *LogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	buf := bytes.Buffer{}
	buf.WriteString(entry.Time.Format(TimeFormat))
	buf.WriteString(" ")
	buf.WriteString(strings.ToUpper(entry.Level.String()))
	buf.WriteByte(' ')

	if entry.HasCaller() {
		buf.WriteString(path.Base(entry.Caller.File))
		buf.WriteByte(':')
		buf.WriteString(strconv.Itoa(entry.Caller.Line))
		buf.WriteByte(' ')
	}

	buf.WriteString(entry.Message)
	buf.WriteByte('\n')

	return buf.Bytes(), nil
}
