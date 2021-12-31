package logger

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	logTypeText = "text"
	logTypeJSON = "json"
)

func InitLog(filename, level, logType string, enableFuncCallDepthRecord bool) error {
	var err error

	err = SetLogLevel(level)
	if err != nil {
		return err
	}

	logrus.SetOutput(&lumberjack.Logger{
		Filename:   filename,
		MaxSize:    500, // 最大的文件500M
		MaxBackups: 10,  // 最多保留10个文件
		MaxAge:     7,   //最长保留7天
	})

	switch logType {
	case logTypeText:
		logrus.SetFormatter(&TextFormatter{TextFormatter: logrus.TextFormatter{DisableColors: true}, EnableFuncCallDepthRecord: enableFuncCallDepthRecord})
	case logTypeJSON:
		fallthrough
	default:
		logrus.SetFormatter(&JSONFormatter{EnableFuncCallDepthRecord: enableFuncCallDepthRecord})
	}

	return nil
}

func SetLogLevel(level string) error {
	l, err := logrus.ParseLevel(level)
	if err == nil {
		logrus.SetLevel(l)
	}

	return err
}
