package unit

import (
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"reflect"
	"strings"

	"github.com/coreos/go-systemd/unit"
)

const MAXLENGTH = 2048

var (
	// Map containing all units found
	Units map[string]*Unit

	// Slice of all loaded units
	Loaded []*Unit
)

// Struct representing the unit
type Unit struct {
	*Definition
	*exec.Cmd
	Loaded bool
	Status status
}

// Definition as found in the unit specification file
type Definition struct {
	Unit struct {
		Description, After, Wants string
	}
	Service struct {
		Type, ExecStart string
	}
	Install struct {
		// WIP
	}
}

// Execution status of a unit
type status int

const (
	Inactive status = iota
	Active
	Exited
	Failed
)

// Looks for unit files in paths given and adds to 'Units' map
func ParseDir(paths ...string) (err error) {
	Units = map[string]*Unit{}
	for _, path := range paths {
		var file *os.File
		var files []os.FileInfo

		if file, err = os.Open(path); err != nil {
			return
		}
		if files, err = file.Readdir(0); err != nil {
			if err == io.EOF {
				log.Println(path + " is empty")
				err = nil
				break
			}
			return
		}

		for _, f := range files {
			var def *Definition

			fpath := path + "/" + f.Name()

			if f.IsDir() {
				return ParseDir(fpath)
			}
			if file, err = os.Open(fpath); err != nil {
				return
			}
			if def, err = ParseUnit(file); err != nil {
				return errors.New("Error parsing " + fpath + ": " + err.Error())
			}
			Units[f.Name()] = &Unit{Definition: def}
		}
	}
	return
}

// Attempts to parse a specification of a unit
func ParseUnit(file io.Reader) (*Definition, error) {
	var err error
	var opts []*unit.UnitOption
	definition := &Definition{}
	def := reflect.ValueOf(definition).Elem()

	if opts, err = unit.Deserialize(file); err != nil {
		return nil, err
	}

	for _, opt := range opts {
		if v := def.FieldByName(opt.Section); v.IsValid() {
			if v := v.FieldByName(opt.Name); v.IsValid() {
				v.SetString(opt.Value)
			} else {
				return nil, errors.New("field " + opt.Name + " does not exist")
			}
		} else {
			return nil, errors.New("section " + opt.Section + " does not exist")
		}
	}
	return definition, nil
}

// Starts execution of the unit's specified command
func (u *Unit) Start() error {
	cmd := strings.Split(u.Service.ExecStart, " ")
	u.Cmd = exec.Command(cmd[0], strings.Join(cmd[1:], " "))

	u.Loaded = true
	return u.Cmd.Start()
}

// Stops execution of the unit's specified command
func (u *Unit) Stop() error {
	u.Loaded = false
	return u.Process.Kill()
}
