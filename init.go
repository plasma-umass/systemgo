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

const (
	host = "127.0.0.1:28537"
)

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

func handleCtlRequests(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("ioutil.ReadAll(body): %s", err)
		return
	}
	defer req.Body.Close()

	var msg systemctl.Request
	if err := json.Unmarshal(body, &msg); err != nil {
		log.Printf("json.Unmarshal: %s", err)
		return
	}

	handler, ok := handlers[msg.Cmd]
	if !ok {
		log.Printf("unhandled command request: %s", msg.Cmd)
		return
	}

	handled := false

	for _, u := range msg.Units {
		result := handler(u)

		resp, err := json.Marshal(result)
		if err != nil {
			log.Printf("json.Marshal(result): %s", err)
			continue
		}

		if _, err := w.Write(resp); err != nil {
			log.Printf("Write(resp): %s", err)
		}

		handled = true
	}

	if !handled {
		// some messages work globally and don't specify units
		// (like a bare `$ systemctl`), and others are
		// potentially in error, forgetting to list a unit.
		log.Printf("TODO: handle messages that don't specify units")
	}
}

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

	err = http.ListenAndServe(host, http.HandlerFunc(handleCtlRequests))
	if err != nil {
		log.Fatalf("ListenAndServe: %s", err)
	}
}
