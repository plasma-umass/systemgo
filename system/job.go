package system

import (
	"errors"
	"sync"

	log "github.com/Sirupsen/logrus"
)

var ErrUnmergeable = errors.New("Unmergeable job types")

type job struct {
	typ  jobType
	unit *Unit

	wants, requires, conflicts         set
	wantedBy, requiredBy, conflictedBy set
	after, before                      set

	executed bool

	waitch chan struct{}
	err    error

	mutex sync.Mutex
}

const JOB_TYPE_COUNT = 4

func newJob(typ jobType, u *Unit) (j *job) {
	log.WithFields(log.Fields{
		"typ": typ,
		"u":   u,
	}).Debugf("newJob")

	return &job{
		typ:  typ,
		unit: u,

		requires:  set{},
		wants:     set{},
		conflicts: set{},

		requiredBy:   set{},
		wantedBy:     set{},
		conflictedBy: set{},

		after:  set{},
		before: set{},

		waitch: make(chan struct{}),
	}
}

//func (j *job) String() string {
//return fmt.Sprintf("%s job for %s", j.typ, j.unit.Name())
//}

func (j *job) IsRedundant() bool {
	switch j.typ {
	case stop:
		return j.unit.IsDeactivating() || j.unit.IsDead()
	case start:
		return j.unit.IsActivating() || j.unit.IsActive()
	case reload:
		return j.unit.IsReloading()
	default:
		return false
	}
}

func (j *job) IsRunning() bool {
	return !j.executed
}

func (j *job) Success() bool {
	return j.State() == success
}

func (j *job) Failed() bool {
	return j.State() == failed
}

func (j *job) Wait() (finished bool) {
	<-j.waitch
	return true
}

func (j *job) isOrphan() bool {
	return len(j.wantedBy) == 0 && len(j.requiredBy) == 0 && len(j.conflictedBy) == 0
}

func (j *job) State() (st jobState) {
	switch {
	case j.IsRunning():
		return running
	case j.err == nil:
		return success
	default:
		return failed
	}
}

func (j *job) Run() (err error) {
	e := log.WithFields(log.Fields{
		"unit": j.unit.Name(),
		"job":  j.typ,
	})
	e.Debugf("j.Run()")

	j.unit.job = j
	defer func() {
		j.err = err
		j.finish()
	}()

	wg := &sync.WaitGroup{}
	for dep := range j.requires {
		wg.Add(1)
		go func(dep *job) {
			e := e.WithField("dep", dep.unit.Name())

			e.Debug("dep.Wait")
			dep.Wait()
			e.Debug("dep.Wait returned")

			if !dep.Success() {
				e.Debugf("->!dep.Success: %s", dep.State())
				j.unit.Log.Errorf("%s failed to %s", dep.unit.Name(), dep.typ)
				err = ErrDepFail
			}
			wg.Done()
		}(dep)
	}
	wg.Wait()

	if err != nil {
		e.Debugf("failed: %s", err)
		return
	}

	switch j.typ {
	case start:
		return j.unit.start()
	case stop:
		return j.unit.stop()
	case restart:
		if err = j.unit.stop(); err != nil {
			return err
		}
		return j.unit.start()
	case reload:
		return j.unit.reload()
	default:
		panic(ErrUnknownType)
	}
}

func (j *job) finish() {
	j.executed = true
	close(j.waitch)
}

var mergeTable = map[jobType]map[jobType]jobType{
	start: {
		start: start,
		//verify_active: start,
		reload:  reload, //reload_or_start
		restart: restart,
	},
	reload: {
		start: reload, //reload_or_start
		//verify_active: reload,
		restart: restart,
	},
	restart: {
		start: restart,
		//verify_active: restart,
		reload: restart,
	},
}

func (j *job) mergeWith(other *job) (err error) {
	t, ok := mergeTable[j.typ][other.typ]
	if !ok {
		return ErrUnmergeable
	}

	j.typ = t

	for jSet, oSet := range map[*set]*set{
		&j.wantedBy:     &other.wantedBy,
		&j.requiredBy:   &other.requiredBy,
		&j.conflictedBy: &other.conflictedBy,

		&j.wants:     &other.wants,
		&j.requires:  &other.requires,
		&j.conflicts: &other.conflicts,
	} {
		for oJob := range *oSet {
			jSet.Put(oJob)
		}
	}

	return
}
