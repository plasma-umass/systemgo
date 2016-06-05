package state

// Activation status of a unit
type Activation int

//go:generate stringer -type=Activation
const (
	Inactive Activation = iota
	Activating
	Active
	Failed // TODO: check
)

// Enable status of a unit
type Enable int

//go:generate stringer -type=Enable
const (
	Disabled Enable = iota
	Static
	Indirect
	Enabled
)

// Load status of a unit
type Load int

//go:generate stringer -type=Load
const (
	Loaded Load = iota
	Error
)

type Sub int

//go:generate stringer -type=Sub
const (
	Unavailable Sub = iota
	Mounted
	Mounting
	// TODO: add all subs
)

type System int

//go:generate stringer -type=System
const (
	Something System = iota // TODO: find all possible states
	Degraded
)
