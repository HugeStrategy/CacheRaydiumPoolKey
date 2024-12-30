package log

import (
	"github.com/sirupsen/logrus"
)

var Logger = logrus.New()

func init() {
	// 配置日志
	Logger.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// 设置日志级别
	Logger.SetLevel(logrus.InfoLevel)
}
