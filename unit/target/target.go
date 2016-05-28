package target

import "github.com/b1101/systemgo/unit"

type Unit struct {
	*unit.Unit
	*Definition
}

type Definition struct {
	*unit.Definition
}

func New() unit.Supervisable {
	target := &Unit{}
	target.Unit = unit.New()
	target.Definition = &Definition{target.Unit.Definition}
	return target
}
func (u *Unit) Stop() {
	//
}
func (u Unit) Sub() string {
	return "WIP"
}
