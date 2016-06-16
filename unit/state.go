package unit

// values in this package correspond to the enums found in
// src/basic/unit-name.h in the systemd library.

// Activation status of a unit -- https://goo.gl/XHBVuC
type Activation int

//go:generate stringer -type=Activation
const (
	Inactive Activation = iota
	Active
	Reloading
	Failed
	Activating
	Deactivating
)

// Load status of a unit definition file -- https://goo.gl/NRBCVK
type Load int

//go:generate stringer -type=Load
const (
	Stub Load = iota
	Loaded
	NotFound
	Error
	Merged
	Masked
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
