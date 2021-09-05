package cronctl

import (
	"errors"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"sync"
)

var (
	jobsNotSetErr         = errors.New("jobs not set")
	crontabNotMappedErr   = errors.New("crontab not Mapped")
	nameNotFoundErr       = errors.New("name not found")
	cronNotFiredErr       = errors.New("cron not fired")
	idNotFoundErr         = errors.New("id not found")
	jobAlreadyEnabledErr  = errors.New("job already enabled")
	jobAlreadyRunningErr  = errors.New("job already running")
	cronAlreadyStoppedErr = errors.New("cron already stopped")
)

type Crontab struct {
	c        *cron.Cron
	jobinfos []*jobinfo
	// back up jobinfos when fired
	jobinfos_ []*jobinfo
	done      chan struct{}
	running   bool
	fired     bool
	cronLock  *sync.RWMutex
	started   chan struct{}
	stopped   chan struct{}
}

type jobinfo struct {
	Name string       `json:"name"`
	Spec string       `json:"spec"`
	Fn   func()       `json:"-"`
	Id   cron.EntryID `json:"id"`
}

func Create(jobs Jobs) (*Crontab, error) {
	if len(jobs) == 0 {
		return nil, jobsNotSetErr
	}

	// get jobinfos with jobs
	jobinfos := jobinfos(jobs)
	if jobinfos == nil {
		return nil, crontabNotMappedErr
	}

	c := cron.New(cron.WithParser(cron.NewParser(
		cron.SecondOptional|cron.Minute|cron.Hour|cron.Dom|cron.Month|cron.Dow|cron.Descriptor,
	)), cron.WithChain(cron.Recover(CronLogger{})))

	// add jobinfos to c
	err := addFuncs(c, jobinfos)
	if err != nil {
		return nil, err
	}

	crontab := &Crontab{
		c:        c,
		jobinfos: jobinfos,
		done:     make(chan struct{}),
		running:  false,
		cronLock: &sync.RWMutex{},
		started:  make(chan struct{}),
		stopped:  make(chan struct{}),
	}

	crontab.jobinfos_ = make([]*jobinfo, len(jobinfos))
	copy(crontab.jobinfos_, jobinfos)
	return crontab, err
}

func addFuncs(c *cron.Cron, jobinfos []*jobinfo) error {
	for _, info := range jobinfos {
		id, err := c.AddFunc(info.Spec, info.Fn)
		if err != nil {
			return err
		}
		info.Id = id
	}
	return nil
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
		crontab.stopped <- struct{}{}
	}
}

func (crontab *Crontab) Suspend() error {
	crontab.cronLock.Lock()
	defer crontab.cronLock.Unlock()

	if !crontab.running {
		return cronAlreadyStoppedErr
	}

	crontab.done <- struct{}{}

	// wait for notification
	select {
	case <-crontab.stopped:
		log.Info("cron has been stopped")
	}
	return nil
}

func (crontab *Crontab) Disable(name string) error {
	crontab.cronLock.Lock()
	defer crontab.cronLock.Unlock()

	id, err := crontab.entryId(true, name)
	if err != nil {
		return err
	}

	if !crontab.fired {
		return cronNotFiredErr
	}

	var tagId = -1
	for idx, info := range crontab.jobinfos {
		if info.Id == *id {
			tagId = idx
			break
		}
	}
	if tagId == -1 {
		return idNotFoundErr
	}

	crontab.c.Remove(*id)
	crontab.jobinfos = append(crontab.jobinfos[:tagId], crontab.jobinfos[tagId+1:]...)
	return nil
}

func (crontab *Crontab) Enable(name string) error {
	crontab.cronLock.Lock()
	defer crontab.cronLock.Unlock()

	id, err := crontab.entryId(true, name)
	if err != nil {
		return err
	}

	if !crontab.fired {
		return cronNotFiredErr
	}

	jobinfos := crontab.jobinfos
	for _, info := range jobinfos {
		if *id == info.Id {
			return jobAlreadyEnabledErr
		}
	}

	jobinfos_ := crontab.jobinfos_
	enabled := false
	for idx, info := range jobinfos_ {
		spec := info.Spec
		entryID := info.Id
		fn := info.Fn
		if entryID == *id {
			newId, err := crontab.c.AddFunc(spec, fn)
			if err != nil {
				return err
			}
			newInfo := jobinfos_[idx]
			newInfo.Id = newId
			crontab.jobinfos = append(jobinfos, newInfo)
			enabled = true
			break
		}
	}

	if enabled {
		return nil
	} else {
		return idNotFoundErr
	}
}

func (crontab *Crontab) Details() (map[string][]*jobinfo, error) {
	crontab.cronLock.RLock()
	defer crontab.cronLock.RUnlock()

	if !crontab.fired {
		return nil, cronNotFiredErr
	}

	var m = make(map[string][]*jobinfo)
	m["current"] = crontab.jobinfos
	m["original"] = crontab.jobinfos_
	return m, nil
}

func (crontab *Crontab) entryId(useBack bool, name string) (*cron.EntryID, error) {
	if !crontab.fired {
		return nil, cronNotFiredErr
	}

	var jobinfos []*jobinfo
	if useBack {
		// look for back data while enabling
		jobinfos = crontab.jobinfos_
	} else {
		// look for current data while disabling
		jobinfos = crontab.jobinfos
	}

	for _, info := range jobinfos {
		if info.Name == name {
			return &info.Id, nil
		}
	}
	return nil, nameNotFoundErr
}
