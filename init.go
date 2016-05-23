package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"github.com/b1101/init/unit"
)

type GlobalState struct {
	// Map containing all units found
	Units map[string]*unit.Unit

	// Slice of all loaded units
	Loaded map[*unit.Unit]bool

	// Status of global state
	State state

	// Number of failed units
	Failed failed

	// Number of queued jobs
	Jobs jobs

	// Init time
	Since time.Time

	// Deal with concurrency
	sync.Mutex
}

//func (u units) String() string {
//out := fmt.Sprint("unit\t\tload\tactive\tsub\tdescription\n")
////for _, sv := range u.Units {
////out += string(sv.)
////}
//return out // TODO: draw a fancy table
//}

type failed int

func (f failed) String() string {
	return fmt.Sprintf("%v units", int(f))
}

type jobs int

func (j jobs) String() string {
	return fmt.Sprintf("%v queued", int(j))
}

type state int

const (
	Something = iota // TODO: find all possible states
	Degraded
)

var (
	State = GlobalState{}
)

func main() {
	var err error

	if State.Units, err = unit.ParseDir("lib/systemd/system", "etc/systemd/system"); err != nil {
		log.Fatalln(err.Error())
	}

	State.Loaded = map[*unit.Unit]bool{}
	unit.Units = State.Units
	unit.Loaded = State.Loaded

	var services []byte
	if services, err = ioutil.ReadFile("services.json"); err != nil {
		log.Fatalln(err.Error())
	}

	//var enabled map[string]*unit.Unit
	var enabled []string
	if err = json.Unmarshal(services, &enabled); err != nil {
		log.Fatalln("Error reading services.json: ", err.Error())
	}

	for _, name := range enabled {
		if u, ok := State.Units[name]; !ok {
			log.Println("unit " + name + " not found")
			break
		} else {
			go func() {
				if err = u.Start(); err != nil {
					log.Println("Error starting", name, err.Error())
				}
			}()
		}
	}
	for {
		select {
		case err := <-unit.Errs:
			log.Println(err.Error())
		default:
			time.Sleep(time.Second)
		}
	}
}

//func StartUnit(name string) (err error) {
//u, ok := Units[name]
//if !ok {
//return errors.New("unit "+name+" not found")
//}
//for {
//if u.DepsLoaded() {
//if err = u.Start(); err != nil {
//return
//}
//break
//}
//time.Sleep(time.Second)
//}
//State.Lock()
//Loaded[name] = true
//State.Unlock()
//}
