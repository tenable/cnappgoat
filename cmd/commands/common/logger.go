package common

import "github.com/sirupsen/logrus"

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})
}

func SetDebug(debug bool) {
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debug("log level set to debug")
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
}
