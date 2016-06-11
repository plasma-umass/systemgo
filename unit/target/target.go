package target

import (
	"fmt"
	"io"

	"github.com/b1101/systemgo/unit"
)

type Unit struct {
	*unit.Unit
	*Definition
}

type Definition struct {
	*unit.Definition
}

func New(definition io.Reader) (*Unit, error) {
	target := &Unit{}
	target.Unit = unit.New()
	target.Definition = &Definition{target.Unit.Definition}

	if err := unit.Define(definition, target.Definition); err != nil {
		return nil, err
	}

	//switch def := target.Definition; {
	// Check for errors

	// Initialisation

	//default:
	return target, nil
	//}
}

func (u *Unit) Start() (err error) {
	//
	return
}

func (u *Unit) Stop() (err error) {
	//
	return
}

func (u Unit) Active() unit.Activation {
	return unit.Inactive
}

func (u Unit) Sub() string {
	return fmt.Sprint(unit.Inactive)
}
