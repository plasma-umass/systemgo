package unit

import (
	"errors"
	"fmt"
)

var ErrNotSet = errors.New("Field not specified")
var ErrNotExist = errors.New("Does not exist")
var ErrNotSupported = errors.New("Not supported")
var ErrUnknownType = errors.New("Unknown type")
var ErrPathNotAbs = errors.New("Path specified is not absolute")
var ErrNotLoaded = errors.New("Unit definition is not loaded properly")
var ErrWrongVal = errors.New("Wrong value received")

type ParseError struct {
	Source string
	Err    error
}

func ParseErr(source string, err error) ParseError {
	return ParseError{
		Source: source,
		Err:    err,
	}
}

func (err ParseError) Error() string {
	return fmt.Sprintf("%s: %s", err.Source, err.Err)
}

type MultiError []error

func (m MultiError) Errors() (errs []string) {
	errs = make([]string, len(m))
	for i, err := range m {
		errs[i] = err.Error()
	}
	return
}

func (m MultiError) Error() string {
	if len(m) == 0 {
		return "No errors"
	}
	return fmt.Sprintf("%d errors encountered, first: %s", len(m), m[0])
}
