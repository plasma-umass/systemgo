package system

import (
	"errors"
	"fmt"
	"sync"

	"github.com/apex/log"
	//log "github.com/Sirupsen/logrus"
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

//go:generate stringer -type=jobState job.go
const (
	waiting jobState = iota
	running
	success
	failed
)

type jobType int

//go:generate stringer -type=jobType job.go
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
	}
}

func (j *job) String() string {
	return fmt.Sprintf("%s job for %s", j.typ, j.unit.Name())
}

type Runner interface {
	Run() error
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
