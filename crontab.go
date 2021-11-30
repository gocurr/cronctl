package cronctl

import (
	"errors"
	"fmt"
	"github.com/robfig/cron/v3"
	"sync"
)

var (
	errNone      = errors.New("crontab is not set")
	errInactive  = errors.New("crontab is inactive")
	errRunning   = errors.New("crontab is running")
	errSuspended = errors.New("crontab is suspended")
)

type Crontab struct {
	c         *cron.Cron
	logger    cron.Logger
	jobinfos  map[string]jobinfo
	jobinfos_ map[string]jobinfo // back up jobinfos when crontab is afire
	done      chan struct{}
	running   bool
	afire     bool
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

func jobInfos(jobs map[string]Job) map[string]jobinfo {
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

func Create(jobs map[string]Job, logger cron.Logger) (*Crontab, error) {
	if len(jobs) == 0 {
		return nil, errNone
	}

	// get jobinfos with jobs
	jobinfos := jobInfos(jobs)

	c := cron.New(cron.WithParser(cron.NewParser(
		cron.SecondOptional|cron.Minute|cron.Hour|cron.Dom|cron.Month|cron.Dow|cron.Descriptor,
	)), cron.WithChain(cron.Recover(logger)))

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
		logger:    logger,
		jobinfos:  jobinfos,
		done:      make(chan struct{}),
		cronLock:  &sync.RWMutex{},
		started:   make(chan struct{}),
		suspended: make(chan struct{}),
	}

	crontab.jobinfos_ = make(map[string]jobinfo)
	// backup jobinfos
	for k, v := range jobinfos {
		crontab.jobinfos_[k] = v
	}
	return crontab, nil
}

func (crontab *Crontab) Startup() error {
	crontab.cronLock.Lock()
	defer crontab.cronLock.Unlock()

	if crontab.running {
		return errRunning
	}

	go crontab.doStart()

	select {
	case <-crontab.started:
		crontab.logger.Info("crontab is running...")
	}
	return nil
}

func (crontab *Crontab) doStart() {
	crontab.c.Start()
	defer crontab.c.Stop()

	crontab.running = true
	crontab.afire = true
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
		return errSuspended
	}

	crontab.done <- struct{}{}

	// wait for notification
	select {
	case <-crontab.suspended:
		crontab.logger.Info("crontab is suspended...")
	}
	return nil
}

func (crontab *Crontab) Disable(name string) error {
	crontab.cronLock.Lock()
	defer crontab.cronLock.Unlock()

	if !crontab.afire {
		return errInactive
	}

	job, ok := crontab.jobinfos[name]
	if !ok {
		return fmt.Errorf("name: %s not found", name)
	}

	crontab.c.Remove(job.Id)
	delete(crontab.jobinfos, job.Name)
	crontab.logger.Info(fmt.Sprintf("job: %s has been disabled", name))
	return nil
}

func (crontab *Crontab) Enable(name string) error {
	crontab.cronLock.Lock()
	defer crontab.cronLock.Unlock()

	if !crontab.afire {
		return errInactive
	}

	jobinfo, ok := crontab.jobinfos[name]
	if ok {
		return fmt.Errorf("job: %s already enabled", name)
	}

	jobinfo, ok = crontab.jobinfos_[name]
	if !ok {
		return fmt.Errorf("name: %s not found", name)
	}

	newId, err := crontab.c.AddFunc(jobinfo.Spec, jobinfo.Fn)
	if err != nil {
		return err
	}
	jobinfo.Id = newId

	crontab.jobinfos[name] = jobinfo
	crontab.logger.Info(fmt.Sprintf("job: %s has been enabled", name))
	return nil
}

func (crontab *Crontab) Details() (map[string]map[string]jobinfo, error) {
	crontab.cronLock.RLock()
	defer crontab.cronLock.RUnlock()

	if !crontab.afire {
		return nil, errInactive
	}

	return map[string]map[string]jobinfo{
		"current":  crontab.jobinfos,
		"original": crontab.jobinfos_,
	}, nil
}
