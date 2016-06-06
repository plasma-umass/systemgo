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

var sys *system.System

var (
	paths = []string{
		// Gentoo-specific
		//"/usr/lib/systemd/system",
		// User overrides
		//"/etc/systemd/system",

		"test",
	}

	handlers = map[string]func(s string) interface{}{
		"status": func(s string) interface{} {
			if st, err := sys.StatusOf(s); err != nil {
				return err
			} else {
				return st
			}
		},
	}
)

func main() {
	var err error

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
		for cmd, handler := range handlers {
			if msg.Cmd == cmd {
				for _, u := range msg.Units {
					x := handler(u)

					b, err := json.Marshal(x)
					if err != nil {
						log.Println(err.Error())
						continue
					}

					if _, err := w.Write(b); err != nil {
						log.Println(err.Error())
					}
				}
				break
			}
		}
	})))
}
