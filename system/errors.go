package system

import "errors"

var ErrIsDir = errors.New("Is a directory")
var ErrNotDir = errors.New("Is not a directory")
var ErrNotFound = errors.New("Not found")
var ErrNotImplemented = errors.New("Not implemented yet")
var ErrDepFail = errors.New("Dependency failed to start. See unit log for details.")
var ErrNotLoaded = errors.New("Unit is not loaded.")
var ErrNoReload = errors.New("Unit does not support reloading")
