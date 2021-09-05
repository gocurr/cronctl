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
// create jobs
jobs := map[string]Job{
"demo1": {Spec: "*/1 * * * * ?", Fn: Counter()},
"demo2": {Spec: "*/2 * * * * ?", Fn: Counter2()},
}

// create a crontab
crontab, err := cronctl.Create(jobs)
```

```go
// startup crontab
err := crontab.Startup()
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
err := crontab.Suspend()
```

```go
err := crontab.Continue()
```

```go
err := crontab.Disable("demo")
```

```go
err := crontab.Enable("demo")
```

### http-functions

We also provide http functions to control `crontab` by http call, e.g.
```go
// setup http controller
token := "xxx"
path := "/inner-access"
crontab.HttpControl(path, token)

http.ListenAndServe(":9090", nil)
```

```
httpstart:
curl http://localhost:9090/inner-access/cron-control?token=xxx\&type=start

httpsuspend:
curl http://localhost:9090/inner-access/cron-control?token=xxx\&type=suspend

httpcontinue:
curl http://localhost:9090/inner-access/cron-control?token=xxx\&type=continue

httpenable:
curl http://localhost:9090/inner-access/cron-control?token=xxx\&type=enable\&name=demo1

httpdisable:
http://localhost:9090/inner-access/cron-control?token=xxx\&type=disable\&name=demo1

httpdetails:
http://localhost:9090/inner-access/cron-control?token=xxx\&type=details
```
