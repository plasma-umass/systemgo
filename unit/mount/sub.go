package mount

// Status of mount units -- https://goo.gl/vg6p7Q
type Sub int

//go:generate stringer -type=Sub
const (
	Dead Sub = iota
	Mounting
	MountingDone
	Mounted
	Remounting
	Unmounting
	MountingSigterm
	MountingSigkill
	RemountingSigterm
	RemountingSigkill
	UnmountingSigterm
	UnmountingSigkill
	Failed
)
