package system

import (
	"os"
	"path/filepath"

	"github.com/b1101/systemgo/unit"
)

// parseFile determines the type of unit specified by the definition found in path,
// creates a new unit of that type and returns a unit.Supervisable and error if any
func parseFile(path string) (u unit.Supervisable, err error) {
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

	var constr unit.Constructor
	if constr, err = unit.ConstructorOfName(path); err != nil {
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
		if unit.SupportedName(name) {
			definitions = append(definitions, filepath.Clean(path+"/"+name))
		}
	}

	return
}
