package cronctl

import (
	log "github.com/sirupsen/logrus"
)

type CronLogger struct {
}

func (c CronLogger) Info(msg string, keysAndValues ...interface{}) {
	log.Infof("%s: %v", msg, keysAndValues)
}

func (c CronLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	log.Errorf("%s %v %v", msg, err, keysAndValues)
}
