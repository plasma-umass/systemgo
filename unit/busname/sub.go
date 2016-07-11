package busname

type Sub int

//go:generate stringer -type=Sub sub.go
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
