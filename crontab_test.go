package cronctl_test

import (
	"fmt"
	"github.com/gocurr/cronctl"
	"net/http"
	"testing"
)

func Test_Crontab(t *testing.T) {
	// create jobs
	var jobs = map[string]cronctl.Job{
		"demo1": {Spec: "*/1 * * * * ?", Fn: Counter()},
		"demo2": {Spec: "*/2 * * * * ?", Fn: Counter2()},
	}

	// create a crontab
	crontab, err := cronctl.Create(jobs, cronctl.Discard)
	crontab, err = cronctl.Create(jobs, cronctl.Logrus)
	if err != nil {
		panic(err)
	}

	// setup http controller
	token := "xxx"
	path := "/inner-access"
	ctrlFunc := crontab.HttpCronCtrlFunc(token, true)
	http.HandleFunc(path, ctrlFunc)

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
		fmt.Printf("count up %v\n", cnt)
	}
}

// cron demo func
func Counter2() func() {
	cnt2 := 10000
	return func() {
		cnt2--
		fmt.Printf("count down %v\n", cnt2)
	}
}
