package unit

import (
	"io"
	"path/filepath"
)

type Constructor func(r io.Reader) (Supervisable, error)

var constructors = map[string]Constructor{
	".service": func(r io.Reader) (Supervisable, error) {
		return NewService(r)
	},
	".target": func(r io.Reader) (Supervisable, error) {
		return NewTarget(r)
	},
}

func ConstructorOf(suffix string) (constr Constructor, err error) {
	if !Supported(suffix) {
		err = ErrUnknownType
	} else {
		constr, _ = constructors[suffix]
	}
	return
}

func ConstructorOfName(filename string) (constr Constructor, err error) {
	return ConstructorOf(filepath.Ext(filename))
}

func Supported(suffix string) (ok bool) {
	_, ok = constructors[suffix]
	return
}

func SupportedName(filename string) (ok bool) {
	return Supported(filepath.Ext(filename))
}
