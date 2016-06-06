package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/b1101/systemgo/lib/handle"
	"github.com/b1101/systemgo/lib/systemctl"
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
	log.Fatalln(http.ListenAndServe(":28537", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		body, err := ioutil.ReadAll(req.Body)
		defer req.Body.Close()
		if err != nil {
			log.Println(err.Error())
		}
		var msg systemctl.Request
		if err := json.Unmarshal(body, &msg); err != nil {
			log.Println(err.Error())
		}
		switch msg.Cmd {
		case "status":
			for _, u := range msg.Units {
				if st, err := sys.StatusOf(u); err != nil {
					w.Write([]byte(err.Error()))
				} else {
					if b, err := json.Marshal(st); err != nil {
						log.Println(err.Error())
					} else {
						w.Write(b)
					}
				}
			}

		}
	})))
}
