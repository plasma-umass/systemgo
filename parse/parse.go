package parse

import (
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/b1101/systemgo/unit"
	"github.com/b1101/systemgo/unit/service"
	systemd "github.com/coreos/go-systemd/unit"
)

var (
	types = map[string]reflect.Type{
		"service": reflect.TypeOf(service.Unit{}),
	}
	re = map[string]string{
		"all":       "*.service|*.target|*.automount|*.device|*.mount|*.path|*.scope|*.slice|*.snapshot|*.socket|*.swap|*.timer",
		"supported": "*.service",
	}
	supported = map[string]map[string]bool{
		"type": {
			"simple":  true,
			"oneshot": true,
		},
		"field": {
			"Description": true,
			"After":       true,
			"Wants":       true,
			"Requires":    true,
			"Conflicts":   true,
			"ExecStart":   true,
		},
		"section": {
			"Service": true,
		},
	}
)

// All searches for all specifications in given paths and returns a map of Supervisables parsed
func All(paths ...string) (map[string]unit.Supervisable, error) {
	units := map[string]unit.Supervisable{}
	for _, path := range paths {
		if err := filepath.Walk(path, func(fpath string, finfo os.FileInfo, err error) error {
			switch {
			case err != nil:
				return err
			case fpath == path, finfo.IsDir() && !is("wants", fpath):
				return nil
			}

			switch u, err := matchAndCreate(finfo.Name()); {
			case err != nil:
				//return err
				handle(err)
				return nil
			default:
				units[finfo.Name()] = u

				if file, err := os.Open(fpath); err != nil {
					file.Close()
					return err
				} else {
					if err := Definition(file, u); err != nil {
						u.Println(err.Error())
						u.SetLoaded(unit.Error)
					}
					file.Close()
				}
			}
			return nil

		}); err != nil {
			handle(err)
			continue
		}
	}
	return units, nil
}

// One searches for specification of unit name in paths, parses and returns a Supervisable
func One(name string, paths ...string) (u unit.Supervisable) {
	switch u, err := matchAndCreate(name); {
	case err != nil:
		handle(err)
		return nil
	default:
		for _, path := range paths {
			switch file, err := os.Open(path + "/" + name); {
			default:
				Definition(file, u)
				file.Close()
			case err != nil:
				if err != os.ErrNotExist {
					handle(err)
				}
				continue
			}
		}
		return u
	}
}

// Definition attempts to parse a specification of a unit into a definition
func Definition(specification io.Reader, unit interface{}) error {
	var err error
	var opts []*systemd.UnitOption
	def := reflect.ValueOf(unit) //.FieldByName("Definition").Elem()
	if !def.CanSet() {
		return errors.New("received a non-pointer value")
	}

	if opts, err = systemd.Deserialize(specification); err != nil {
		return err
	}

	for _, opt := range opts {
		if v := def.FieldByName(opt.Section).Elem(); v.IsValid() && v.CanSet() {
			if v := v.FieldByName(opt.Name); v.IsValid() && v.CanSet() {
				switch v.Kind() {
				case reflect.String:
					v.SetString(opt.Value)
				case reflect.Bool:
					if opt.Value == "yes" {
						v.SetBool(true)
					}
				case reflect.SliceOf(reflect.TypeOf(reflect.String)).Kind():
					v.Set(reflect.ValueOf(strings.Split(opt.Value, " ")))
				case reflect.SliceOf(reflect.TypeOf(reflect.Int)).Kind():
					ints := []int{}
					for _, val := range strings.Split(opt.Value, " ") {
						if converted, err := strconv.Atoi(val); err == nil {
							ints = append(ints, converted)
						} else {
							return errors.New("Error parsing " + opt.Name + ": " + err.Error())
						}
					}
					v.Set(reflect.ValueOf(ints))
				default:
					return errors.New("Can not parse " + opt.Name)
				}
			} else {
				return errors.New("field " + opt.Name + " does not exist")
			}
		} else {
			return errors.New("section " + opt.Section + " does not exist")
		}
	}
	return nil
	//switch {
	//case definition.Service != nil:
	//if definition.Service.ExecStart == "" {
	//return errors.New("'ExecStart' field not set")
	//}

	//if definition.Service.Type == "" {
	//definition.Service.Type = "simple"
	//}

	//if !supported["type"][definition.Service.Type] {
	//return errors.New("service type " + definition.Service.Type + " does not exist or is not supported yet")
	//}
	//}
}

// is checks if filename extension matches given type
func is(typ, name string) bool {
	switch match, err := regexp.MatchString(".*[.]"+typ, name); {
	case err != nil:
		handle(err)
		fallthrough
	case !match:
		return false
	default:
		return true
	}
}

// matchAndCreate determines the unit type by name and returns a Supervisable of that type
func matchAndCreate(name string) (unit.Supervisable, error) {
	for suffix, t := range types {
		if is(suffix, name) {
			switch u, ok := reflect.New(t).Interface().(unit.Supervisable); {
			case !ok:
				return nil, errors.New("unit type is not Supervisable")
			default:
				u.Init()
				return u, nil
			}
		}
	}
	return nil, errors.New(name + " does not match any known unit type")
}

func handle(err error) {
	log.Println(err.Error()) // TODO: fixme
}
