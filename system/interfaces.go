package system

import "github.com/b1101/systemgo/lib/state"

type Supervisable interface {
	StartStopper

	Description() string

	Active() state.Active

	Wants() []string
	Requires() []string
	Conflicts() []string
	After() []string
}

type Reloader interface {
	Reload()
}

type StartStopper interface {
	Starter
	Stopper
}
type Starter interface {
	Start() error
}
type Stopper interface {
	Stop() error
}
