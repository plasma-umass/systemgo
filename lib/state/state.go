package state

// values in these packages correspond to the enums found in
// src/basic/unit-name.h in the systemd library.

// Activation status of a unit -- https://goo.gl/XHBVuC
type Active int

//go:generate stringer -type=Active
const (
	UnitActive Active = iota
	UnitReloading
	UnitInactive
	UnitFailed
	UnitActivating
	UnitDeactivating
)

// Load status of a unit definition file -- https://goo.gl/NRBCVK
type Load int

//go:generate stringer -type=Load
const (
	UnitStub Load = iota
	UnitLoaded
	UnitNotFound
	UnitError
	UnitMerged
	UnitMasked
)

// Status of service units -- https://goo.gl/eg9PS3
type Service int

//go:generate stringer -type=Service
const (
	ServiceDead Service = iota
	ServiceStartPre
	ServiceStart
	ServiceStartPost
	ServiceRunning
	ServiceExited // not running anymore, but RemainAfterExit true for this unit
	ServiceReload
	ServiceStop
	ServiceStopSigabrt // watchdog timeout
	ServiceStopSigterm
	ServiceStopSigkill
	ServiceStopPost
	ServiceFinalSigterm
	ServiceFinalSigkill
	ServiceFailed
	ServiceAutoRestart
)

// Status of mount units -- https://goo.gl/vg6p7Q
type Mount int

//go:generate stringer -type=Mount
const (
	MountDead Mount = iota
	MountMounting
	MountMountingDone
	MountMounted
	MountRemounting
	MountUnmounting
	MountMountingSigterm
	MountMountingSigkill
	MountRemountingSigterm
	MountRemountingSigkill
	MountUnmountingSigterm
	MountUnmountingSigkill
	MountFailed
)

type System int

//go:generate stringer -type=System
const (
	Something System = iota // TODO: find all possible states
	Degraded
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
