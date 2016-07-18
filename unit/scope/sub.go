package scope

type Sub int

//go:generate stringer -type=Sub sub.go
const (
	Dead Sub = iota
	Running
	Abandoned
	StopSigterm
	StopSigkill
	Failed
)
