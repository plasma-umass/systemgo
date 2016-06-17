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
type definition struct {
	Unit struct {
		Description                               string
		Documentation                             string
		Wants, Requires, Conflicts, Before, After []string
	}
	Install struct {
		WantedBy, RequiredBy []string
	}
}

// Description returns a string as found in definition
func (def definition) Description() string {
	return def.Unit.Description
}

// Documentation returns a string as found in definition
func (def definition) Documentation() string {
	return def.Unit.Documentation
}

// Wants returns a slice of unit names as found in definition
func (def definition) Wants() []string {
	return def.Unit.Wants
}

// Requires returns a slice of unit names as found in definition
func (def definition) Requires() []string {
	return def.Unit.Requires
}

// Conflicts returns a slice of unit names as found in definition
func (def definition) Conflicts() []string {
	return def.Unit.Conflicts
}

// After returns a slice of unit names as found in definition
func (def definition) After() []string {
	return def.Unit.After
}

// After returns a slice of unit names as found in definition
func (def definition) Before() []string {
	return def.Unit.After
}

// RequiredBy returns a slice of unit names as found in definition
func (def definition) RequiredBy() []string {
	return def.Install.RequiredBy
}

// WantedBy returns a slice of unit names as found in definition
func (def definition) WantedBy() []string {
	return def.Install.WantedBy
}

// ParseDefinition parses the data in Systemd unit-file format and stores the result in value pointed by definition
func parseDefinition(contents io.Reader, definition interface{}) (err error) { // TODO: find a better name for io.Reader parameter
	// Access the underlying value of the pointer
	def := reflect.ValueOf(definition).Elem()
	if !def.IsValid() || !def.CanSet() {
		return errors.New("Wrong value received")
	}

	// Deserialized options
	var opts []*unit.UnitOption
	if opts, err = unit.Deserialize(contents); err != nil {
		return
	}

	// Loop over deserialized options trying to match them to the ones as found in definition
	for _, opt := range opts {
		if v := def.FieldByName(opt.Section); v.IsValid() && v.CanSet() {
			if v := v.FieldByName(opt.Name); v.IsValid() && v.CanSet() {
				// reflect.Kind of field in definition
				switch v.Kind() {

				// string
				case reflect.String:
					v.SetString(opt.Value)

				// bool
				case reflect.Bool:
					if opt.Value == "yes" {
						v.SetBool(true)
					} else if opt.Value != "no" {
						return ParseErr(opt.Name, errors.New(`Value should be "yes" or "no"`))
					}

				// []string
				case reflect.SliceOf(reflect.TypeOf(reflect.String)).Kind():
					v.Set(reflect.ValueOf(strings.Split(opt.Value, " ")))

				// []int
				case reflect.SliceOf(reflect.TypeOf(reflect.Int)).Kind():
					ints := []int{}
					for _, val := range strings.Split(opt.Value, " ") {
						if converted, err := strconv.Atoi(val); err == nil {
							ints = append(ints, converted)
						} else {
							return ParseErr(opt.Name, err)
						}
					}
					v.Set(reflect.ValueOf(ints))

				// unknown
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
