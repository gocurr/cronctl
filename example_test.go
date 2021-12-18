package cronctl_test

import (
	"fmt"
	"github.com/gocurr/cronctl"
)

var times = 6
var counter int
var countUp = make(chan int)

// cron demo func
func Count() {
	counter++
	countUp <- counter
	if counter == times {
		close(countUp)
	}
}

func Example_crontab() {
	// create jobs
	var jobs = map[string]cronctl.Job{
		"demo1": {Spec: "*/1 * * * * ?", Fn: Count},
	}

	// create a crontab
	crontab, err := cronctl.Create(jobs, cronctl.Discard)
	if err != nil {
		panic(err)
	}

	if err := crontab.Startup(); err != nil {
		panic(err)
	}

	for v := range countUp {
		fmt.Printf("%v ", v)
	}

	// Output: 1 2 3 4 5 6

}
