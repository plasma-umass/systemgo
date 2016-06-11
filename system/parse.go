package system

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/b1101/systemgo/lib/handle"
	"github.com/b1101/systemgo/unit"
	"github.com/b1101/systemgo/unit/service"
	"github.com/b1101/systemgo/unit/target"
)

var (
	constructors = map[string]func(io.Reader) (Supervisable, error){
		".service": func(r io.Reader) (Supervisable, error) {
			return service.New(r)
		},
		".target": func(r io.Reader) (Supervisable, error) {
			return target.New(r)
		},
	}
)

// ParseAll searches for all specifications in given paths and returns a map of *Unit's parsed
func ParseAll(paths ...string) (map[string]*Unit, error) {
	units := map[string]*Unit{}
	for _, path := range paths {
		if err := filepath.Walk(path, func(fpath string, finfo os.FileInfo, _err error) error {
			switch {
			case _err != nil:
				return _err
			case fpath == path, finfo.IsDir() && !strings.HasSuffix(fpath, ".wants"):
				return nil
			}

			var b bytes.Buffer
			u := NewUnit(&b) // TODO: use config
			u.name = finfo.Name()
			units[finfo.Name()] = u

			file, err := os.Open(fpath)
			defer file.Close()
			if err != nil {
				u.Log.Println(err.Error())
				return nil
			}

			if sup, err := matchAndCreate(finfo.Name(), file); err != nil {
				u.Log.Println(err.Error())
				u.stats.loaded = unit.Error
			} else {
				u.Supervisable = sup
			}
			u.stats.path = fpath
			return nil

		}); err != nil {
			handle.Err(err)
			continue
		}
	}
	return units, nil
}

//
// TODO: Come back when reload() is ready
//

//// ParseOne searches for specification of unit name in given paths, parses and returns a Supervisable
//func ParseOne(name string, paths ...string) (Supervisable, error) {
//	for _, path := range paths {
//		file, err := os.Open(path + "/" + name)
//		defer file.Close()
//		if err != nil {
//			if err != os.ErrNotExist {
//				handle.Err(err)
//			}
//			continue
//		}
//
//		u, err := matchAndCreate(name, file)
//
//		return &Unit{}
//	}
//	return nil, ErrNotExist // TODO: replace by not found, when lib is ready
//}

// matchAndCreate determines the unit type by name, creates and returns a Supervisable of that type
func matchAndCreate(name string, definition io.Reader) (Supervisable, error) {
	for suffix, constructor := range constructors {
		if strings.HasSuffix(name, suffix) {
			return constructor(definition)
		}
	}
	return nil, errors.New(name + " does not match any known unit type")
}
