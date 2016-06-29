package system

import (
	"io"
	"os"
	"path/filepath"

	"github.com/b1101/systemgo/unit"
)

type UnitConstructor func(r io.Reader) (Supervisable, error)

var constructors = map[string]UnitConstructor{
	".service": func(r io.Reader) (Supervisable, error) {
		return unit.NewService(r)
	},
	".target": func(r io.Reader) (Supervisable, error) {
		return unit.NewTarget(r)
	},
}

func ConstructorOf(suffix string) (constr UnitConstructor, err error) {
	if !Supported(suffix) {
		err = ErrUnknownType
	} else {
		constr, _ = constructors[suffix]
	}
	return
}

func ConstructorOfName(filename string) (constr UnitConstructor, err error) {
	return ConstructorOf(filepath.Ext(filename))
}

func Supported(suffix string) (ok bool) {
	_, ok = constructors[suffix]
	return
}

func SupportedName(filename string) (ok bool) {
	return Supported(filepath.Ext(filename))
}

// parseFile determines the type of unit specified by the definition found in path,
// creates a new unit of that type and returns a Supervisable and error if any
func parseFile(path string) (u Supervisable, err error) {
	var file *os.File
	if file, err = os.Open(path); err != nil {
		return
	}
	defer file.Close()

	var info os.FileInfo
	if info, err = file.Stat(); err != nil {
		return
	}

	if info.IsDir() {
		return nil, ErrIsDir
	}

	var constr UnitConstructor
	if constr, err = ConstructorOfName(path); err != nil {
		return
	}

	return constr(file)
}

// pathset returns a slice of paths to definitions of supported unit types found in path specified
func pathset(path string) (definitions []string, err error) {
	var file *os.File
	if file, err = os.Open(path); err != nil {
		return
	}
	defer file.Close()

	var info os.FileInfo
	if info, err = file.Stat(); err != nil {
		return
	} else if !info.IsDir() {
		err = ErrNotDir
		return
	}

	var names []string
	if names, err = file.Readdirnames(0); err != nil {
		return
	}

	definitions = make([]string, 0, len(names))

	for _, name := range names {
		if SupportedName(name) {
			definitions = append(definitions, filepath.Clean(path+"/"+name))
		}
	}

	return
}
