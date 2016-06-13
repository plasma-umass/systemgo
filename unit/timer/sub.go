package timer

type Sub int

//go:generate stringer -type=Sub
const (
	Dead Sub = iota
	Waiting
	Running
	Elapsed
	Failed
)
