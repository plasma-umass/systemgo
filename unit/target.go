package unit

import "io"

// Target unit
type Target struct {
	targetDefinition
}

// Target unit definition
type targetDefinition struct {
	definition
	// TODO: extend with target-specific fields
}

func NewTarget(definition io.Reader) (target *Target, err error) {
	target = &Target{}

	if err = parseDefinition(definition, &target.targetDefinition); err != nil {
		return
	}

	//switch def := target.targetDefinition; {
	// Check for errors

	// Initialisation

	//default:
	return
	//}
}
