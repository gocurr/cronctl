# cronctl

To download, run:

```bash
go get -u github.com/gocurr/cronctl
```

Import it in your program as:

```go
import "github.com/gocurr/cronctl"
```

It requires Go 1.11 or later due to usage of Go Modules.

- To start a crontab:

```go
package main

import (
	"fmt"
	"github.com/gocurr/cronctl"
)

func main() {
	// create jobs
	jobs := map[string]cronctl.Job{
		"demo1": {Spec: "*/1 * * * * ?", Fn: func() {
			fmt.Println("demo1")
		}},
		"demo2": {Spec: "*/2 * * * * ?", Fn: func() {
			fmt.Println("demo2")
		}},
	}

	// create a crontab
	crontab, _ := cronctl.Create(jobs, cronctl.Logrus)

	// startup crontab
	_ = crontab.Startup()

	// ...
}
```

### The `crontab` has enhanced methods:

- `Details`: List running jobs.

- `Start`: Start Mapped jobs.

- `Continue`: Continue Mapped jobs.

- `Suspend`: Immediately Suspends all the jobs.

- `Enable`: Immediately Readds the job.

- `Disable`: Immediately Removes jhe job.

You can `Start`|`Continue`|`Suspend`|`Enable`|`Disable` the `crontab` anytime.

```go
details, err := crontab.Details()
_ = details

err = crontab.Suspend()

err = crontab.Continue()

err = crontab.Disable("demo")

err = crontab.Enable("demo")
```

### http-functions

We also provide http functions to control `crontab` by http call, e.g.

```go
// setup http controller
token := "xxx"
path := "/inner-access"
ctrlFunc := crontab.HttpCronCtrlFunc(token, true)
http.HandleFunc(path, ctrlFunc)
```

- httpstart:

```bash
curl http://localhost:9090/inner-access?token=xxx\&type=start
```

- httpsuspend:

```bash
curl http://localhost:9090/inner-access?token=xxx\&type=suspend
```

- httpcontinue:

```bash
curl http://localhost:9090/inner-access?token=xxx\&type=continue
```

- httpenable:

```bash
curl http://localhost:9090/inner-access?token=xxx\&type=enable\&name=demo1
```

- httpdisable:

```bash
curl http://localhost:9090/inner-access?token=xxx\&type=disable\&name=demo1
```

- httpdetails:

```bash
curl http://localhost:9090/inner-access?token=xxx\&type=details
```
