package service

// Status of service units -- https://goo.gl/eg9PS3
type Sub int

//go:generate stringer -type=Sub
const (
	Dead Sub = iota
	StartPre
	Start
	StartPost
	Running
	Exited // not running anymore, but RemainAfterExit true for this unit
	Reload
	Stop
	StopSigabrt // watchdog timeout
	StopSigterm
	StopSigkill
	StopPost
	FinalSigterm
	FinalSigkill
	Failed
	AutoRestart
)
