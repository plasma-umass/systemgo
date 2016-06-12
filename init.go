package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/b1101/systemgo/lib/errors"
	"github.com/b1101/systemgo/lib/handle"
	"github.com/b1101/systemgo/lib/systemctl"
	"github.com/b1101/systemgo/system"
)

type Handler func(s string) (interface{}, error)

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

	handlers = map[string]Handler{
		"status": func(s string) (interface{}, error) {
			if len(s) == 0 {
				return sys.Status(), nil
			}
			return sys.StatusOf(s)
		},
		"start": func(s string) (interface{}, error) {
			return nil, sys.Start(s)
		},
		"stop": func(s string) (interface{}, error) {
			return nil, sys.Stop(s)
		},
		"": func(s string) (interface{}, error) {
			return nil, errors.WIP
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

	for name, h := range handlers {
		func(handler Handler) {
			http.HandleFunc("/"+name, func(w http.ResponseWriter, req *http.Request) {
				v := req.URL.Query()

				units, ok := v["unit"]
				if !ok {
					units = []string{""}
				}

				for _, u := range units {
					msg := systemctl.Response{}

					result, err := handler(u)
					if err != nil {
						msg.Error = err.Error()
					}

					msg.Yield = result

					resp, err := json.Marshal(msg)
					if err != nil {
						log.Printf("json.Marshal(result): %s", err)
						continue
					}

					if _, err := w.Write(resp); err != nil {
						log.Printf("Write(resp): %s", err)
					}

				}
			})
		}(h)
	}

	err = http.ListenAndServe(host, nil)
	if err != nil {
		log.Fatalf("ListenAndServe: %s", err)
	}
}
