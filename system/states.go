package system

type State int

//go:generate stringer -type=State
const (
	Something State = iota // TODO: find all possible states
	Degraded
)
