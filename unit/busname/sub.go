package busname

type Sub int

//go:generate stringer -type=Sub
const (
	Dead Sub = iota
	Making
	Registered
	Listening
	Running
	Sigterm
	Sigkill
	Failed
)
