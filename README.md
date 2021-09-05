crontab

`crontab` is an enhance tool for [cron](https://github.com/robfig/cron).

It's also written in go, with enhance functions:

# Details

List running jobs in `crontab` container.

# Start

Start Mapped jobs.

# Continue

Continue Mapped jobs.

# Suspend

Immediately Suspends all the jobs.

# Enable

Immediately Readds the job.

# Disable

Immediately Removes jhe job.

## httpstart

curl http://localhost:9090/inner/cron-control?token=xxx\&type=start

## httpsuspend

curl http://localhost:9090/inner/cron-control?token=xxx\&type=suspend

## httpenable

curl http://localhost:9090/inner/cron-control?token=xxx\&type=enable\&name=demo1

## httpdisable

http://localhost:9090/inner/cron-control?token=xxx\&type=disable\&name=demo1

## httpdetails

http://localhost:9090/inner/cron-control?token=xxx\&type=details
