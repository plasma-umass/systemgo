package system

import (
	"errors"
	"fmt"
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
}

const JOB_TYPE_COUNT = 4

type jobState int

//go:generate stringer -type=jobState
const (
	waiting jobState = iota
	running
	success
	failed
)

type jobType int

//go:generate stringer -type=jobType
const (
	start jobType = iota
	stop
	reload
	restart
)

func newJob(typ jobType, u *Unit) (j *job) {
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
	}
}

func (j *job) String() string {
	return fmt.Sprintf("%s job for %s", j.typ, j.unit.Name())
}

func (j *job) IsRunning() bool {
	return j.waitch != nil
}

func (j *job) Success() bool {
	return j.State() == success
}

func (j *job) Failed() bool {
	return j.State() == failed
}

func (j *job) Wait() (finished bool) {
	if j.IsRunning() {
		<-j.waitch
	}
	return true
}

func (j *job) isOrphan() bool {
	return len(j.wantedBy) == 0 && len(j.requiredBy) == 0 && len(j.conflictedBy) == 0
}

func (j *job) State() (st jobState) {
	switch {
	case j.IsRunning():
		return running
	case !j.executed:
		return waiting
	case j.err == nil:
		return success
	default:
		return failed
	}
}

func (j *job) Run() (err error) {
	j.waitch = make(chan struct{})

	j.unit.job = j

	wg := &sync.WaitGroup{}
	for dep := range j.requires {
		wg.Add(1)
		go func(dep *job) {
			defer wg.Done()
			log.Debugf("%v waiting for %v", j, dep)
			dep.Wait()
			if !dep.Success() {
				j.unit.Log.Printf("%v failed", j)
				err = ErrDepFail
			}
		}(dep)
	}

	wg.Wait()

	j.err = j.execute()

	j.unit.job = nil

	close(j.waitch)
	j.waitch = nil

	return
}

func (j *job) execute() (err error) {
	wg := &sync.WaitGroup{}
	for dep := range j.requires {
		if !dep.Success() {
			wg.Add(1)
			go func() {
				dep.Wait()
				if !dep.Success() {
					err = dep.err
				}
				wg.Done()
			}()
		}
	}

	wg.Wait()
	if err != nil {
		return
	}

	switch j.typ {
	case start:
		err = j.unit.start()
	case stop:
		err = j.unit.stop()
	case restart:
		if err = j.unit.stop(); err != nil {
			break
		}
		err = j.unit.start()
	case reload:
		err = j.unit.reload()
	default:
		panic(ErrUnknownType)
	}
	j.executed = true

	return
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
