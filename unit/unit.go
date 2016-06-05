package unit

import (
	"errors"
	"io"
	"reflect"
	"strconv"
	"strings"

	"github.com/coreos/go-systemd/unit"
)

// Struct representing the unit
type Unit struct {
	*Definition
}

type Definition struct {
	Unit struct {
		Description                               string
		Documentation                             []string
		After, Wants, Requires, Conflicts, Before []string
	}
	Install struct {
		WantedBy string
	}
}

func New() (u *Unit) {
	u = &Unit{}
	u.Definition = &Definition{}

	return
}

func (u Unit) Description() string {
	return u.Definition.Unit.Description
}
func (u Unit) Wants() []string {
	return u.Definition.Unit.Wants
}
func (u Unit) Requires() []string {
	return u.Definition.Unit.Requires
}
func (u Unit) Conflicts() []string {
	return u.Definition.Unit.Conflicts
}
func (u Unit) After() []string {
	return u.Definition.Unit.After
}

// Define attempts to parse a specification of a unit into a definition
func Define(specification io.Reader, definition interface{}) error {
	var err error
	var opts []*unit.UnitOption
	def := reflect.ValueOf(definition).Elem()

	if !def.IsValid() || !def.CanSet() {
		return errors.New("Received a non-pointer value")
	}

	if opts, err = unit.Deserialize(specification); err != nil {
		return err
	}

	for _, opt := range opts {
		if v := def.FieldByName(opt.Section); v.IsValid() && v.CanSet() {
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
				return errors.New("field " + opt.Name + " does not exist or is not supported yet")
			}
		} else {
			return errors.New("section " + opt.Section + " does not exist or is not supported yet")
		}
	}
	return nil
}
