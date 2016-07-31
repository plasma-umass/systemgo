package system

import "errors"

var ErrUnmergeable = errors.New("Unmergeable job types")

type job struct {
	typ  jobType
	unit *Unit

	wants, requires, conflicts         set
	wantedBy, requiredBy, conflictedBy set
}

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

type set map[*job]struct{}

func (s set) Contains(j *job) (ok bool) {
	_, ok = s[j]
	return
}

func (s set) Put(j *job) {
	s[j] = struct{}{}
}

func (s set) Remove(j *job) {
	delete(s, j)
}

type Runner interface {
	Run() error
}

func (j *job) Run() (err error) {
	switch j.typ {
	case start:
		return j.unit.Start()
	case stop:
		return j.unit.Stop()
	case restart:
		if err = j.unit.Stop(); err != nil {
			return
		}
		return j.unit.Start()
	case reload:
		return j.unit.Reload()
	}
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
	if j.typ == other.typ {
		return
	}

	var t jobType
	if t, ok = mergeTable[j.typ][other.typ]; !ok {
		return ErrUnmergeable
	}

	j.typ = t
	return
}
