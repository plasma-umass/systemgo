package system

type TargetSub int

//go:generate stringer -type=TargetSub target_sub.go
const (
	Dead TargetSub = iota
	Active
)
