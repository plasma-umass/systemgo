package system

type jobState int

//go:generate stringer -type=jobState job_generate.go
const (
	waiting jobState = iota
	running
	success
	failed
)

type jobType int

//go:generate stringer -type=jobType job_generate.go
const (
	start jobType = iota
	stop
	reload
	restart
)
