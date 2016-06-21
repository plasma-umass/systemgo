package system

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/b1101/systemgo/unit"
)

// Load searches for a definition of unit name in configured paths and returns a unit.Supervisable (or nil) and error if any
func (sys *System) load(name string) (u *Unit, err error) {
	if !unit.SupportedName(name) {
		return nil, unit.ErrUnknownType
	}

	for _, path := range sys.paths {
		fpath := filepath.Clean(path + "/" + name)

		sup, err := parsePath(fpath)
		if err == os.ErrNotExist {
			continue
		}

		u = NewUnit()
		u.path = path

		sys.parsedPaths[path] = u
		sys.parsed[name] = u

		if err != nil {
			u.Log.Printf("Error parsing %s: %s", fpath, err)
			u.loaded = unit.Error
			return u, err
		}

		sys.loaded[name] = u
		u.loaded = unit.Loaded

		u.Supervisable = sup
		for ptr, deps := range map[*[]*Unit][]string{
			&u.After:    u.Supervisable.After(),
			&u.Wants:    u.Supervisable.Wants(),
			&u.Before:   u.Supervisable.Before(),
			&u.Requires: u.Supervisable.Requires(),
		} {
			arr := make([]*Unit, len(deps))

			for i, name := range deps {
				if arr[i], err = sys.Unit(name); err != nil {
					if ptr == &u.Requires {
						u.Log.Printf("Error loading dependency %s: %s", name, err)
						u.loaded = unit.Error
						return u, err
					}
				}
			}

			*ptr = arr
		}

		for ptr, suffix := range map[*[]*Unit]string{
			&u.Wants:    ".wants",
			&u.Requires: ".required",
		} {
			dpath := fpath + suffix

			if err = filepath.Walk(dpath, func(fpath string, finfo os.FileInfo, ferr error) error {
				switch {
				case ferr != nil:
					return ferr
				case fpath == dpath:
					if !finfo.IsDir() {
						err = ErrNotDir
					}
					return err
				case !unit.SupportedName(fpath):
					return nil
				}

				if path, err = os.Readlink(fpath); err != nil {
					return fmt.Errorf("Error reading link %s: %s", finfo.Name(), err)
				}

				var dep *Unit
				dep, ok := sys.parsedPaths[path]
				if !ok {
					dep = NewUnit()
					if dep.Supervisable, err = parsePath(path); err != nil {
						return fmt.Errorf("Error parsing definition of %s: %s", path, err)
					}
				}

			}); err != nil {
				u.Log.Printf("Error parsing %s: %s", dpath, err)
				u.loaded = unit.Error
			}
		}
		return u, err
	}
	return nil, ErrNotFound
}

// parsePath determines the type of unit specified by the definition found in path,
// creates a new unit of that type and returns a unit.Supervisable and error if any
func parsePath(path string) (u unit.Supervisable, err error) {
	var file *os.File
	if file, err = os.Open(fpath); err != nil {
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

	return constr(definition)
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
