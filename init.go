package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/b1101/systemgo/system"
)

// TODO: introduce proper configuration
type config struct {
	// Default target
	Target string

	// Paths to search for unit files
	Paths []string

	// Method of communication with systemctl
	Method string

	// Http connection configuration
	*httpConfig
}

type httpConfig struct {
	// Port for system daemon to listen on
	Port
}

type Port int

func (p Port) String() string {
	return fmt.Sprintf(":%v", int(p))
}

var (
	// Instance of a system
	sys system.Daemon

	// Default configuration
	conf *config = &config{
		Target: "default.target",
		Method: "http",
		httpConfig: &httpConfig{
			Port: 28537,
		},
		Paths: []string{
			// Gentoo-specific
			//"/usr/lib/systemd/system",
			// User overrides
			//"/etc/systemd/system",

			"test",
		},
	}
)

func main() {
	// Initialize system
	sys = system.New()
	sys.SetPaths(conf.Paths...)

	// Start the default target
	if err := sys.Start(conf.Target); err != nil {
		//sys.Log.Printf("Error starting default target %s: %s", conf.Target, err)
		log.Printf("Error starting default target %s: %s", conf.Target, err)
		if err = sys.Start("rescue.target"); err != nil {
			//sys.Log.Printf("Error starting rescue target %s: %s", "rescue.target", err)
			log.Printf("Error starting rescue target %s: %s", "rescue.target", err)
		}
	}

	// Listen for systemctl requests
	switch conf.Method {
	case "http":
		if err := listenHTTP(conf.Port.String()); err != nil {
			log.Fatalf("Error starting server on %s: %s", conf.Port, err)
		}
	}
}

// Handle systemctl requests using HTTP
func listenHTTP(addr string) (err error) {
	server := http.NewServeMux()

	for name, handler := range sys.Handlers() {
		func(handler system.Handler) {
			server.HandleFunc("/"+name, func(w http.ResponseWriter, req *http.Request) {
				//msg := []systemctl.Response{}
				v := req.URL.Query()

				names, ok := v["unit"]
				if !ok {
					names = []string{}
				}

				var resp []byte
				if resp, err = json.Marshal(handler(names...)); err != nil {
					log.Printf("json.Marshal(result): %s", err)
					return
				}

				if _, err = w.Write(resp); err != nil {
					log.Printf("Write(resp): %s", err)
				}
			})
		}(handler)
	}

	return http.ListenAndServe(addr, server)
}
