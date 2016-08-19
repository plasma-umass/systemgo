package system

import "errors"

var ErrIsDir = errors.New("Is a directory")
var ErrNotDir = errors.New("Is not a directory")
var ErrNotFound = errors.New("Not found")
var ErrDepFail = errors.New("Dependency failed to start. See unit log for details.")
var ErrDepConflict = errors.New("Error stopping conflicting unit")
var ErrNotLoaded = errors.New("Unit is not loaded.")
var ErrNoReload = errors.New("Unit does not support reloading")
var ErrUnknownType = errors.New("Unknown type")
var ErrNotActive = errors.New("Unit is not active")
var ErrExists = errors.New("Unit already exists")
var ErrNotImplemented = errors.New("Not implemented yet")
var ErrUnmergeable = errors.New("Unmergeable job types")
