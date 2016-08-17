package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/rvolosatovs/systemgo/config"
	"github.com/rvolosatovs/systemgo/system"
	"github.com/rvolosatovs/systemgo/systemctl"
)

// Initializes the system, sets the default paths, as specified in configuration and attempts to start the default target, falls back to "rescue.target", if it fails
func main() {
	go Serve()

	// Initialize system
	log.Info("Systemgo starting...")

	System.SetPaths(config.Paths...)

	// Start the default target
	if err := System.Start(config.Target); err != nil {
		log.Errorf("Error starting default target %s: %s", config.Target, err)
		if err = System.Start(config.RESCUE_TARGET); err != nil {
			log.Errorf("Error starting rescue target %s: %s", config.RESCUE_TARGET, err)
		}
	}

	if log.GetLevel() == log.DebugLevel {
		go func() {
			for range time.Tick(5 * time.Second) {
				for _, u := range System.Units() {
					st, err := System.StatusOf(u.Name())
					if err != nil {
						panic(err)
					}
					fmt.Println("********************************************************************************")
					fmt.Println("\t\t", u.Name())
					fmt.Println("********************************************************************************")
					fmt.Printf("->Status:\n%s\n", st)
					fmt.Println("--------------------------------------------------------------------------------")
					fmt.Printf("->Unit:\n%+v\n", u)
					fmt.Println("********************************************************************************")
				}
				fmt.Println("********************************************************************************")
				fmt.Println("********************************************************************************")
				fmt.Println("********************************************************************************")
			}
		}()
	}

	exit := make(chan os.Signal)
	signal.Notify(exit, os.Interrupt, os.Kill)
	<-exit

	log.Infoln("Shutting down...")
	if err := System.Isolate("shutdown.target"); err != nil {
		log.Fatalf("Error shutting down: %s", err)
	}

	wg := &sync.WaitGroup{}
	for _, u := range System.Units() {
		if u.IsActive() {
			log.Infof("Waiting for %s to stop", u.Name())
			wg.Add(1)

			go func(u *system.Unit) {
				defer wg.Done()

				var t time.Duration
				for range time.Tick(time.Second) {
					t += time.Second
					if !u.IsActive() || t == time.Minute {
						return
					}
				}
			}(u)
		}
	}
}

// Instance of a system
var System = system.New()

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
