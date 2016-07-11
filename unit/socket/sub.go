package socket

type Sub int

//go:generate stringer -type=Sub
const (
	Dead Sub = iota
	StartPre
	StartChown
	StartPost
	Listening
	Running
	StopPre
	StopPreSigterm
	StopPreSigkill
	StopPost
	FinalSigterm
	FinalSigkill
	Failed
)
