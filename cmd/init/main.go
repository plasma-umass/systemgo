package main

import (
	"encoding/json"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/rvolosatovs/systemgo/config"
	"github.com/rvolosatovs/systemgo/system"
	"github.com/rvolosatovs/systemgo/systemctl"
)

// Initializes the system, sets the default paths, as specified in configuration and attempts to start the default target, falls back to "rescue.target", if it fails
func main() {
	// Initialize system
	System = system.New()
	System.SetPaths(config.Paths...)

	go Serve()

	// Start the default target
	if err := System.Start(config.Target); err != nil {
		log.Errorf("Error starting default target %s: %s", config.Target, err)
		if err = System.Start(config.RESCUE_TARGET); err != nil {
			log.Errorf("Error starting rescue target %s: %s", config.RESCUE_TARGET, err)
		}
	}

	select {}
}

// Instance of a system
var System *system.Daemon

// Listen for systemctl requests
func Serve() {
	for {
		if err := listenHTTP(config.Port.String()); err != nil {
			log.Errorf("Error listening on %v: %s", config.Port, err)
		}
		log.Infof("Retrying in %v seconds", config.Retry)
		time.Sleep(config.Retry)
	}
}

// Handle systemctl requests using HTTP
func listenHTTP(addr string) (err error) {
	server := http.NewServeMux()

	for name, handler := range systemctl.Handlers(System) {
		func(handler systemctl.Handler) {
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

	log.Infof("Listening on http://localhost%s", config.Port)
	return http.ListenAndServe(addr, server)
}
