package system

import (
	"errors"
	"fmt"
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
	errored
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

func (u *Unit) Wait() {
	u.loading.Wait()
}

type Runner interface {
	Run() error
}

func (j *job) IsRunning() (running bool) {
	return j.loadch != nil
}

func (j *job) Wait() (finished bool) {
	if j.IsRunning() {
		<-j.loadch
	}
	return true
}

func (j *job) Run() (err error) {
	j.loadch = make(chan struct{})

	j.unit.job = j

	switch j.typ {
	case start:
		err = j.unit.start()
	case stop:
		err = j.unit.stop()
	case restart:
		if err = j.unit.stop(); err != nil {
			return
		}
		err = j.unit.start()
	case reload:
		err = j.unit.reload()
	default:
		panic(ErrUnknownType)
	}

	j.err = err

	j.executed = true
	j.unit.job = nil

	close(j.loadch)
	j.loadch = nil

	return
}

func (j *job) State() (st jobState) {
	switch {
	case j.IsRunning():
		return running
	case j.err != nil:
		return errored
	case j.err == nil && j.executed:
		return success
	default:
		return waiting
	}
}
