package system

type job struct {
	typ  jobType
	unit *Unit
}

type Runner interface {
	Run() error
}

func (j *job) Run() (err error) {
}

type startJob job

func (j *startJob) Run() (err error) {
	return j.unit.Start()
}

type stopJob job

func (j *stopJob) Run() (err error) {
	return j.unit.Stop()
}

type restartJob job

func (j *restartJob) Run() (err error) {
	if err = j.unit.Stop(); err == nil {
		return j.unit.Start()
	}
	return
}

type reloadJob job

func (j *reloadJob) Run() (err error) {
}
