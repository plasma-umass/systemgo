package main

import (
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/b1101/systemgo/parse"
	"github.com/b1101/systemgo/system"
)

type Reloader interface {
	Reload()
}

var paths = []string{
	// Gentoo-specific
	"/usr/lib/systemd/system",
	// User overrides
	"/etc/systemd/system",
}

var (
	State = system.State{}
)

var (
	ErrNotFound = errors.New("not found")
	ErrNoReload = errors.New("does not support reloading")
)

func main() {
	var err error

	if State.Units, err = parse.All(paths...); err != nil {
		log.Fatalln(err.Error())
	}

	for n, u := range State.Units {
		fmt.Println(n)
		fmt.Println(u.Status())
		switch reader, ok := u.(io.Reader); {
		case ok:
			b := make([]byte, 100)
			if n, _ := reader.Read(b); n > 0 {
				fmt.Println("\nLog:")
				fmt.Println(string(b))
			}
		default:
			log.Println(n, "not readable")
		}
		fmt.Println()
	}
}

//for _, u := range State.Units {
//u.GS = &State
//}
//State.Loaded = map[*unit.Unit]bool{}
//unit.Units = State.Units
//unit.Loaded = State.Loaded

//var services []byte
//if services, err = ioutil.ReadFile("services.json"); err != nil {
//log.Fatalln(err.Error())
//}

////var enabled map[string]*unit.Unit
//var enabled []string
//if err = json.Unmarshal(services, &enabled); err != nil {
//log.Fatalln("Error reading services.json: ", err.Error())
//}

//for _, name := range enabled {
//if u, ok := State.Units[name]; !ok {
//log.Println("unit " + name + " not found")
//break
//} else {
//go func() {
//if err = u.Start(); err != nil {
//log.Println("Error starting", name, err.Error())
//}
//}()
//}
//}
//for {
//select {
//case err := <-unit.Errs:
//log.Println(err.Error())
//default:
//time.Sleep(time.Second)
//}
//}
//}

func Start(name string) (err error) {
	switch u, ok := State.Units[name]; {
	case !ok:
		err = ErrNotFound
	default:
		u.Start()
	}
	return
}

func Stop(name string) (err error) {
	switch u, ok := State.Units[name]; {
	case !ok:
		err = ErrNotFound
	default:
		u.Stop()
	}
	return
}

func Restart(name string) (err error) {
	switch u, ok := State.Units[name]; {
	case !ok:
		err = ErrNotFound
	default:
		u.Stop()
		u.Start()
	}
	return
}

func Reload(name string) (err error) {
	switch u, ok := State.Units[name]; {
	case !ok:
		err = ErrNotFound
	default:
		switch reloader, ok := u.(Reloader); {
		case !ok:
			err = ErrNoReload
		default:
			reloader.Reload()
		}
	}
	return
}

//func getUnit(name string) (unit.Supervisable, error) {
//if u, ok := State.Units[name]; !ok {
//return nil, errors.New(name + "not found")
//} else {
//return u, nil
//}
//}

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
