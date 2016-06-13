package system

type State int

//go:generate stringer -type=State state.go  //TODO: find out why this breaks `go generate`
const (
	Something State = iota // TODO: find all possible states
	Degraded
)
