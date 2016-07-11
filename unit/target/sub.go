package target

type Sub int

//go:generate stringer -type=Sub
const (
	Dead Sub = iota
	Active
)
