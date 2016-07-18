package timer

type Sub int

//go:generate stringer -type=Sub sub.go
const (
	Dead Sub = iota
	Waiting
	Running
	Elapsed
	Failed
)
