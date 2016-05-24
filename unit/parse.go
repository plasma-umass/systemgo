package unit

import (
	"errors"
	"io"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/coreos/go-systemd/unit"
)

var supported = map[string]bool{
	"simple": true,
}

// Looks for unit files in paths given and returns a map of units
func ParseDir(paths ...string) (map[string]*Unit, error) {
	units := map[string]*Unit{}
	for _, path := range paths {
		var err error
		var file *os.File
		var files []os.FileInfo

		if file, err = os.Open(path); err != nil {
			return nil, err
		}
		if files, err = file.Readdir(0); err != nil {
			if err == io.EOF {
				log.Println(path + " is empty")
				err = nil
				break
			}
			return nil, err
		}

		for _, f := range files {
			var u *Unit

			fpath := path + "/" + f.Name()

			if f.IsDir() {
				//return ParseDir(fpath)
				//			return nil, errors.New("not implemented yet") //TODO
				log.Println("recursive directory parsing not implemented yet")
				continue
			}
			if file, err = os.Open(fpath); err != nil {
				return nil, err
			}
			if u, err = ParseUnit(file); err != nil {
				//return nil, errors.New("error parsing " + fpath + ": " + err.Error())
				log.Println("error parsing " + fpath + ": " + err.Error())
				continue
			}
			units[f.Name()] = u
		}
	}
	return units, nil
}

// Attempts to parse a specification of a unit
func ParseUnit(specification io.Reader) (*Unit, error) {
	var err error
	var opts []*unit.UnitOption
	definition := &Definition{}
	def := reflect.ValueOf(definition).Elem()

	if opts, err = unit.Deserialize(specification); err != nil {
		return nil, err
	}

	for _, opt := range opts {
		if v := def.FieldByName(opt.Section); v.IsValid() {
			if v := v.FieldByName(opt.Name); v.IsValid() {
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
							return nil, errors.New("Error parsing " + opt.Name + ": " + err.Error())
						}
					}
					v.Set(reflect.ValueOf(ints))
				default:
					return nil, errors.New("Can not parse " + opt.Name)
				}
			} else {
				return nil, errors.New("field " + opt.Name + " does not exist")
			}
		} else {
			return nil, errors.New("section " + opt.Section + " does not exist")
		}
	}

	if definition.Service.ExecStart == "" {
		return nil, errors.New("'ExecStart' field not set")
	}

	if definition.Service.Type == "" {
		definition.Service.Type = "simple"
	}

	if !supported[definition.Service.Type] {
		return nil, errors.New("service type " + definition.Service.Type + " does not exist or is not supported yet")
	}

	return &Unit{Definition: definition}, nil
}
