package init

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/rvolosatovs/systemgo/system"
	"github.com/rvolosatovs/systemgo/systemctl"
)

// TODO: introduce proper configuration
type Configuration struct {
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
	System *system.Daemon

	// Default configuration
	Conf = &Configuration{
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

// boot initializes the system, sets the default paths, as specified in configuration and attempts to start the default target, falls back to "rescue.target", if it fails
func Boot() {
	// Initialize system
	System = system.New()

	System.SetPaths(Conf.Paths...)

	// Start the default target
	if err := System.Start(Conf.Target); err != nil {
		//sys.Log.Printf("Error starting default target %s: %s", Conf.Target, err)
		log.Printf("Error starting default target %s: %s", Conf.Target, err)
		if err = System.Start("rescue.target"); err != nil {
			//sys.Log.Printf("Error starting rescue target %s: %s", "rescue.target", err)
			log.Printf("Error starting rescue target %s: %s", "rescue.target", err)
		}
	}
}

// Listen for systemctl requests
func Serve() {
	switch Conf.Method {
	case "http":
		if err := listenHTTP(Conf.Port.String()); err != nil {
			log.Fatalf("Error starting server on %s: %s", Conf.Port, err)
		}
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

	return http.ListenAndServe(addr, server)
}
