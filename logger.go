package cronctl

import (
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

var Logrus DefaultLogger

var Discard = cron.DiscardLogger

type DefaultLogger struct {
}

func (l DefaultLogger) Info(msg string, _ ...interface{}) {
	log.Infof("%s", msg)
}

func (l DefaultLogger) Error(err error, msg string, _ ...interface{}) {
	log.Errorf("%s %v", msg, err)
}
