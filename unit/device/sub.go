package busname

type Sub int

//go:generate stringer -type=Sub sub.go
const (
	Dead Sub = iota
	Tentative
	Plugged
)
