package logx

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
)

func InitLogx() {
	//logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	level, _ := logrus.ParseLevel(viper.GetString("log.level"))
	logrus.SetLevel(level)
	logrus.SetReportCaller(true)
}
