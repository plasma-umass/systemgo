package errors

import "errors"

var (
	NotFound = errors.New("not found")
	NoReload = errors.New("does not support reloading")
	WIP      = errors.New("WIP")
)
