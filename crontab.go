package cronctl

import (
	"errors"
	"fmt"
	"github.com/robfig/cron/v3"
	"sync"
)

var errInactive = errors.New("crontab is inactive")

type Crontab struct {
	c         *cron.Cron
	logger    cron.Logger
	jobInfos  map[string]jobInfo
	_jobInfos map[string]jobInfo // back up jobInfos when crontab is created
	done      chan struct{}
	running   bool
	afire     bool
	cronLock  *sync.RWMutex
	started   chan struct{}
	suspended chan struct{}
}

type jobInfo struct {
	Name string       `json:"name"`
	Spec string       `json:"spec"`
	Fn   func()       `json:"-"`
	Id   cron.EntryID `json:"id"`
}

type Job struct {
	Spec string `json:"spec"`
	Fn   func() `json:"-"`
}

func convert(jobs map[string]Job) map[string]jobInfo {
	var jobInfos = make(map[string]jobInfo)
	for name, job := range jobs {
		fn := job.Fn
		jobInfos[name] = jobInfo{
			Name: name,
			Spec: job.Spec,
			Fn:   fn,
		}
	}

	return jobInfos
}

func Create(jobs map[string]Job, logger cron.Logger) (*Crontab, error) {
	if len(jobs) == 0 {
		return nil, errors.New("empty jobs")
	}

	// convert jobs to jobInfos
	jobInfos := convert(jobs)

	c := cron.New(cron.WithParser(cron.NewParser(
		cron.SecondOptional|cron.Minute|cron.Hour|cron.Dom|cron.Month|cron.Dow|cron.Descriptor,
	)), cron.WithChain(cron.Recover(logger)))

	// add spec-function to c
	for name, info := range jobInfos {
		id, err := c.AddFunc(info.Spec, info.Fn)
		if err != nil {
			return nil, err
		}
		jobInfos[name] = jobInfo{
			Name: name,
			Spec: info.Spec,
			Fn:   info.Fn,
			Id:   id,
		}
	}

	crontab := &Crontab{
		c:         c,
		logger:    logger,
		jobInfos:  jobInfos,
		done:      make(chan struct{}),
		cronLock:  &sync.RWMutex{},
		started:   make(chan struct{}),
		suspended: make(chan struct{}),
	}

	crontab._jobInfos = make(map[string]jobInfo)
	// backup jobInfos
	for k, v := range jobInfos {
		crontab._jobInfos[k] = v
	}
	return crontab, nil
}

func (crontab *Crontab) Startup() error {
	crontab.cronLock.Lock()
	defer crontab.cronLock.Unlock()

	if crontab.running {
		return errors.New("crontab was started up")
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
		return errors.New("crontab was suspended")
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

	job, ok := crontab.jobInfos[name]
	if !ok {
		return fmt.Errorf("name: %s not found", name)
	}

	crontab.c.Remove(job.Id)
	delete(crontab.jobInfos, job.Name)
	crontab.logger.Info(fmt.Sprintf("job: %s has been disabled", name))
	return nil
}

func (crontab *Crontab) Enable(name string) error {
	crontab.cronLock.Lock()
	defer crontab.cronLock.Unlock()

	if !crontab.afire {
		return errInactive
	}

	job, ok := crontab.jobInfos[name]
	if ok {
		return fmt.Errorf("job: %s already enabled", name)
	}

	job, ok = crontab._jobInfos[name]
	if !ok {
		return fmt.Errorf("name: %s not found", name)
	}

	newId, err := crontab.c.AddFunc(job.Spec, job.Fn)
	if err != nil {
		return err
	}
	job.Id = newId

	crontab.jobInfos[name] = job
	crontab.logger.Info(fmt.Sprintf("job: %s has been enabled", name))
	return nil
}

func (crontab *Crontab) Details() (map[string]map[string]jobInfo, error) {
	crontab.cronLock.RLock()
	defer crontab.cronLock.RUnlock()

	if !crontab.afire {
		return nil, errInactive
	}

	return map[string]map[string]jobInfo{
		"current":  crontab.jobInfos,
		"original": crontab._jobInfos,
	}, nil
}
