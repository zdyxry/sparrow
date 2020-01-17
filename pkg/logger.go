package pkg

import (
	"github.com/sirupsen/logrus"
	"os"
)

type Logger struct {
}

func (logger Logger) Write(data []byte) (count int, error error) {
	return len(data), nil
}

var log *logrus.Logger

func init() {
	log = logrus.New()
	log.SetOutput(os.Stdout)
	log.SetFormatter(&logrus.JSONFormatter{})
}