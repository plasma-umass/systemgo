package system

type State int

//go:generate stringer -type=State state.go
const (
	Something State = iota // TODO: find all possible states
	Degraded
)
