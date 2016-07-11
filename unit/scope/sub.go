package scope

type Sub int

//go:generate stringer -type=Sub
const (
	Dead Sub = iota
	Running
	Abandoned
	StopSigterm
	StopSigkill
	Failed
)
