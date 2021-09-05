package cronctl

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"testing"
)

func Test_Main(t *testing.T) {
	// create jobs
	var jobs = Jobs{
		{Name: "demo1", Spec: "*/1 * * * * ?", Fn: Counter()},
		{Name: "demo2", Spec: "*/2 * * * * ?", Fn: Counter2()},
	}

	// create a crontab
	crontab, err := Create(jobs)
	if err != nil {
		panic(err)
	}

	// setup http controller
	token := "abc"
	path := "/inn  er-access"
	crontab.HttpControl(path, token)

	// startup crontab
	err = crontab.Startup()
	if err != nil {
		panic(err)
	}

	// listen and serve
	_ = http.ListenAndServe(":9090", nil)
}

// cron demo func
func Counter() func() {
	cnt := 0
	return func() {
		cnt++
		log.Infof("count up %v", cnt)
	}
}

// cron demo func
func Counter2() func() {
	cnt2 := 10000
	return func() {
		cnt2--
		log.Infof("count down %v", cnt2)
	}
}
