package cronctl

import (
	"errors"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"sync"
)

var (
	jobsNotSetErr           = errors.New("jobs not set")
	nameNotFoundErr         = errors.New("name not found")
	cronNotFiredErr         = errors.New("cron not fired")
	jobAlreadyEnabledErr    = errors.New("job already enabled")
	jobAlreadyRunningErr    = errors.New("job already running")
	cronAlreadySuspendedErr = errors.New("cron already suspended")
)

type Crontab struct {
	c        *cron.Cron
	jobinfos map[string]jobinfo
	// back up jobinfos when fired
	jobinfos_ map[string]jobinfo
	done      chan struct{}
	running   bool
	fired     bool
	cronLock  *sync.RWMutex
	started   chan struct{}
	suspended chan struct{}
}

type jobinfo struct {
	Name string       `json:"name"`
	Spec string       `json:"spec"`
	Fn   func()       `json:"-"`
	Id   cron.EntryID `json:"id"`
}

type Job struct {
	Spec string `json:"spec"`
	Fn   func() `json:"-"`
}

func jobinfos(jobs map[string]Job) map[string]jobinfo {
	var jobinfos = make(map[string]jobinfo)
	for name, job := range jobs {
		fn := job.Fn
		jobinfos[name] = jobinfo{
			Name: name,
			Spec: job.Spec,
			Fn:   fn,
		}
	}

	return jobinfos
}

func Create(jobs map[string]Job) (*Crontab, error) {
	if len(jobs) == 0 {
		return nil, jobsNotSetErr
	}

	// get jobinfos with jobs
	jobinfos := jobinfos(jobs)

	c := cron.New(cron.WithParser(cron.NewParser(
		cron.SecondOptional|cron.Minute|cron.Hour|cron.Dom|cron.Month|cron.Dow|cron.Descriptor,
	)), cron.WithChain(cron.Recover(CronLogger{})))

	// add jobinfos to c
	for name, info := range jobinfos {
		id, err := c.AddFunc(info.Spec, info.Fn)
		if err != nil {
			return nil, err
		}
		jobinfos[name] = jobinfo{
			Name: name,
			Spec: info.Spec,
			Fn:   info.Fn,
			Id:   id,
		}
	}

	crontab := &Crontab{
		c:         c,
		jobinfos:  jobinfos,
		done:      make(chan struct{}),
		cronLock:  &sync.RWMutex{},
		started:   make(chan struct{}),
		suspended: make(chan struct{}),
	}

	crontab.jobinfos_ = make(map[string]jobinfo)
	//backup jobinfos
	for k, v := range jobinfos {
		crontab.jobinfos_[k] = v
	}
	return crontab, nil
}

func (crontab *Crontab) Startup() error {
	crontab.cronLock.Lock()
	defer crontab.cronLock.Unlock()

	if crontab.running {
		return jobAlreadyRunningErr
	}

	go crontab.doStart()

	select {
	case <-crontab.started:
		log.Info("cron has been started")
	}
	return nil
}

func (crontab *Crontab) doStart() {
	crontab.c.Start()
	defer crontab.c.Stop()

	crontab.running = true
	crontab.fired = true
	crontab.started <- struct{}{}

	select {
	case <-crontab.done:
		crontab.running = false
		crontab.suspended <- struct{}{}
	}
}

func (crontab *Crontab) Continue() error {
	return crontab.Startup()
}

func (crontab *Crontab) Suspend() error {
	crontab.cronLock.Lock()
	defer crontab.cronLock.Unlock()

	if !crontab.running {
		return cronAlreadySuspendedErr
	}

	crontab.done <- struct{}{}

	// wait for notification
	select {
	case <-crontab.suspended:
		log.Info("cron has been suspended")
	}
	return nil
}

func (crontab *Crontab) Disable(name string) error {
	crontab.cronLock.Lock()
	defer crontab.cronLock.Unlock()

	if !crontab.fired {
		return cronNotFiredErr
	}

	job, ok := crontab.jobinfos[name]
	if !ok {
		return nameNotFoundErr
	}

	crontab.c.Remove(job.Id)
	delete(crontab.jobinfos, job.Name)
	return nil
}

func (crontab *Crontab) Enable(name string) error {
	crontab.cronLock.Lock()
	defer crontab.cronLock.Unlock()

	if !crontab.fired {
		return cronNotFiredErr
	}

	jobinfo, ok := crontab.jobinfos[name]
	if ok {
		return jobAlreadyEnabledErr
	}

	jobinfo, ok = crontab.jobinfos_[name]
	if !ok {
		return nameNotFoundErr
	}

	newId, err := crontab.c.AddFunc(jobinfo.Spec, jobinfo.Fn)
	if err != nil {
		return err
	}
	jobinfo.Id = newId

	crontab.jobinfos[name] = jobinfo
	return nil
}

func (crontab *Crontab) Details() (map[string]map[string]jobinfo, error) {
	crontab.cronLock.RLock()
	defer crontab.cronLock.RUnlock()

	if !crontab.fired {
		return nil, cronNotFiredErr
	}

	return map[string]map[string]jobinfo{
		"current":  crontab.jobinfos,
		"original": crontab.jobinfos,
	}, nil
}
