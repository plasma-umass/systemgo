package unit

type Sub int

const (
	// Not defined to avoid redeclaration TODO: find a better workaround
	// Activating     // /sbin/swapon is running, but the swap not yet enabled
	// Failed
	// Deactivating

	dead Sub = iota
	running
	making
	registered
	listening
	sigterm
	sigkill
	tentative
	plugged
	mounting
	mountingdone
	mounted
	remounting
	unmounting
	mountingsigterm
	mountingsigkill
	remountingsigterm
	remountingsigkill
	unmountingsigterm
	unmountingsigkill
	waiting
	abandoned
	stopsigterm
	stopsigkill
	startpre
	start
	startpost
	exited // not running anymore, but remainafterexit true for this unit
	reload
	stop
	stopsigabrt // watchdog timeout
	stoppost
	autorestart
	startchown
	stoppre
	stoppresigterm
	stoppresigkill
	finalsigterm
	finalsigkill
	activatingdone // /sbin/swapon is running, and the swap is done
	activatingsigterm
	activatingsigkill
	deactivatingsigterm
	deactivatingsigkill
	elapsed
)
