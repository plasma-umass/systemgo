package main

import (
	"fmt"
	"log"
	"time"

	"github.com/b1101/systemgo/lib/handle"
	"github.com/b1101/systemgo/system"
)

type Reloader interface {
	Reload()
}

var paths = []string{
	// Gentoo-specific
	//"/usr/lib/systemd/system",
	// User overrides
	//"/etc/systemd/system",

	"test",
}

func main() {
	var err error
	var sys *system.System

	if sys, err = system.New(paths...); err != nil {
		log.Fatalln(err.Error())
	}

	var st interface{}

	st = sys.Status()
	fmt.Println(st)

	if err = sys.Start("sv.service"); err != nil {
		handle.Err(err)
	}

	time.Sleep(3 * time.Second)

	st, _ = sys.StatusOf("dep.service")
	fmt.Println(st)

	st, _ = sys.StatusOf("sv.service")
	fmt.Println(st)
}
