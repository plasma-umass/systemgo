package automount

type Sub int

//go:generate stringer -type=Sub
const (
	Dead Sub = iota
	Waiting
	Running
	Failed
)
