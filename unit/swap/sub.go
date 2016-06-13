package swap

type Sub int

//go:generate stringer -type=Sub
const (
	Dead           Sub = iota
	Activating         // /sbin/swapon is running, but the swap not yet enabled
	ActivatingDone     // /sbin/swapon is running, and the swap is done
	Active
	Deactivating
	ActivatingSigterm
	ActivatingSigkill
	DeactivatingSigterm
	DeactivatingSigkill
	Failed
)
