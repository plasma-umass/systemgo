package unit

import (
	"errors"
	"io"
	"reflect"
	"strconv"
	"strings"

	"github.com/coreos/go-systemd/unit"
)

// Definition of a unit matching the fields found in unit-file
type Definition struct {
	Unit struct {
		Description                               string
		Documentation                             string
		Wants, Requires, Conflicts, Before, After []string
	}
	Install struct {
		WantedBy, RequiredBy []string
	}
}

// Description returns a string as found in Definition
func (def Definition) Description() string {
	return def.Unit.Description
}

// Documentation returns a string as found in Definition
func (def Definition) Documentation() string {
	return def.Unit.Documentation
}

// Wants returns a slice of unit names as found in Definition
func (def Definition) Wants() []string {
	return def.Unit.Wants
}

// Requires returns a slice of unit names as found in Definition
func (def Definition) Requires() []string {
	return def.Unit.Requires
}

// Conflicts returns a slice of unit names as found in Definition
func (def Definition) Conflicts() []string {
	return def.Unit.Conflicts
}

// After returns a slice of unit names as found in Definition
func (def Definition) After() []string {
	return def.Unit.After
}

// Before returns a slice of unit names as found in Definition
func (def Definition) Before() []string {
	return def.Unit.Before
}

// RequiredBy returns a slice of unit names as found in Definition
func (def Definition) RequiredBy() []string {
	return def.Install.RequiredBy
}

// WantedBy returns a slice of unit names as found in Definition
func (def Definition) WantedBy() []string {
	return def.Install.WantedBy
}

// ParseDefinition parses the data in Systemd unit-file format and stores the result in value pointed by Definition
func ParseDefinition(r io.Reader, v interface{}) (err error) {
	// Access the underlying value of the pointer
	def := reflect.ValueOf(v).Elem()

	if !def.IsValid() || !def.CanSet() {
		return ErrWrongVal
	}

	// Deserialized options
	var opts []*unit.UnitOption
	if opts, err = unit.Deserialize(r); err != nil {
		return
	}

	// Loop over deserialized options trying to match them to the ones as found in Definition
	for _, opt := range opts {
		if v := def.FieldByName(opt.Section); v.IsValid() && v.CanSet() {
			if v := v.FieldByName(opt.Name); v.IsValid() && v.CanSet() {
				// reflect.Kind of field in Definition
				switch v.Kind() {

				case reflect.String:
					v.SetString(opt.Value)

				case reflect.Bool:
					if opt.Value == "yes" {
						v.SetBool(true)
					} else if opt.Value != "no" {
						return ParseErr(opt.Name, errors.New(`Value should be "yes" or "no"`))
					}

				case reflect.Slice:
					if _, ok := v.Interface().([]string); ok { // []string
						v.Set(reflect.ValueOf(strings.Fields(opt.Value)))

					} else if _, ok := v.Interface().([]int); ok { // []int
						ints := []int{}
						for _, val := range strings.Fields(opt.Value) {
							if converted, err := strconv.Atoi(val); err == nil {
								ints = append(ints, converted)
							} else {
								return ParseErr(opt.Name, err)
							}
						}
						v.Set(reflect.ValueOf(ints))
					}

				default:
					return ParseErr(opt.Name, ErrUnknownType)
				}
			} else {
				return ParseErr(opt.Name, ErrNotExist)
			}
		} else {
			return ParseErr(opt.Name, ErrNotExist)
		}
	}
	return
}
