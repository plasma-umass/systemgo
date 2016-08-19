package system

// System state
type State int

//go:generate stringer -type=State state.go

const (
	Initializing State = iota
	Starting
	Running
	Degraded
	Maintenance
	Stopping
)
