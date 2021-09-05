crontab

`crontab` is an enhance tool
for [cron](https://github.com/robfig/cron).

It's also written in go, with enhance functions:

# Details
List running jobs in `crontab` container.

# Start
Start or Continue Mapped jobs.

# Suspend
Immediately Suspends all the jobs in `crontab` container.

# Enable
Immediately Readds the job to `crontab` container.

# Disable
Immediately Removes jhe job in `crontab` container.

## httpstart
curl http://localhost:9090/crontab-start?token=xxx

## httpstop
curl http://localhost:9090/crontab-stop?token=xxx

## httpenable
curl http://localhost:9090/crontab-enable?name=demo1\&token=xxx

## httpdisable
curl http://localhost:9090/crontab-disable?name=demo1\&token=xxx

## httpdetails
curl http://localhost:9090/crontab-details?token=xxx
