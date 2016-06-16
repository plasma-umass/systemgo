package target

import (
	"io"

	"github.com/b1101/systemgo/unit"
)

// Target unit
type Unit struct {
	Definition
}

// Target unit definition
type Definition struct {
	unit.Definition
	// TODO: extend with target-specific fields
}

func New(definition io.Reader) (target *Unit, err error) {
	if err = unit.ParseDefinition(definition, target.Definition); err != nil {
		return
	}

	//switch def := target.Definition; {
	// Check for errors

	// Initialisation

	//default:
	return
	//}
}
