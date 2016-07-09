package systemctl

import (
	"github.com/b1101/systemgo/system"
	"github.com/b1101/systemgo/unit"
)

type Daemon interface {
	Start(...string) error
	Stop(string) error
	Restart(string) error
	Reload(string) error
	Enable(string) error
	Disable(string) error

	Status() (system.Status, error)
	StatusOf(string) (unit.Status, error)
	IsEnabled(string) (unit.Enable, error)
	IsActive(string) (unit.Activation, error)
}
