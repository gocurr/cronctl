package cronctl

import (
	log "github.com/sirupsen/logrus"
)

type DefaultLogger struct {
}

func (l DefaultLogger) Info(msg string, _ ...interface{}) {
	log.Infof("%s", msg)
}

func (l DefaultLogger) Error(err error, msg string, _ ...interface{}) {
	log.Errorf("%s %v", msg, err)
}
