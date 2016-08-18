package main

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"
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

	sys.SetPaths(config.Paths...)

	// Start the default target
	if err := sys.Start(config.Target); err != nil {
		log.Errorf("Error starting default target %s: %s", config.Target, err)
		if err = sys.Start(config.RESCUE_TARGET); err != nil {
			log.Errorf("Error starting rescue target %s: %s", config.RESCUE_TARGET, err)
		}
	}

	if log.GetLevel() == log.DebugLevel {
		go func() {
			for range time.Tick(5 * time.Second) {
				for _, u := range sys.Units() {
					st, err := sys.StatusOf(u.Name())
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
	if err := sys.Isolate("shutdown.target"); err != nil {
		log.Fatalf("Error shutting down: %s", err)
	}

	wg := &sync.WaitGroup{}
	for _, u := range sys.Units() {
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
var sys = system.New()

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
	daemonRPC := systemctl.NewServer(sys)
	rpc.Register(daemonRPC)
	rpc.HandleHTTP()

	e := log.WithField("port", config.Port)

	l, err := net.Listen("tcp", config.Port.String())
	if err != nil {
		e.Fatalf("Listen error: %s", err)
	}

	log.Infof("Listening on http://localhost%s", config.Port)
	return http.Serve(l, nil)
}
