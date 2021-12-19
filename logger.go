package cronctl

import (
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

var Discard = cron.DiscardLogger

var Logrus defaultLogger

type defaultLogger struct {
}

func (l defaultLogger) Info(msg string, _ ...interface{}) {
	log.Infof("%s", msg)
}

func (l defaultLogger) Error(err error, msg string, _ ...interface{}) {
	log.Errorf("%s %v", msg, err)
}
