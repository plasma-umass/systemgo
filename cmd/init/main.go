package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/rvolosatovs/systemgo/system"
	"github.com/rvolosatovs/systemgo/systemctl"
)

func main() {
	Boot()
	Serve()
}

var Conf Configuration

// TODO: introduce proper configuration
type Configuration struct {
	// Default target
	Target string `yaml:omitempty`

	// Paths to search for unit files
	Paths []string `yaml:omitempty`

	// Debug
	Debug bool `yaml:omitempty`

	// Port for system daemon to listen on
	Port port `yaml:omitempty`
}

type port int

func (p port) String() string {
	return fmt.Sprintf(":%v", int(p))
}

// Instance of a system
var System *system.Daemon

func init() {
	viper.SetConfigName("systemgo")

	viper.AddConfigPath(".")
	if os.Getenv("XDG_CONFIG_HOME") != "" {
		viper.AddConfigPath("$XDG_CONFIG_HOME/systemgo")
	}
	viper.AddConfigPath("/etc/systemgo")

	if err := viper.ReadInConfig(); err != nil {
		log.WithField("path", viper.ConfigFileUsed()).Errorf("Error reading config: %s", err)
	}
	log.WithField("path", viper.ConfigFileUsed()).Infof("Found configuration file")

	viper.SetDefault("port", 28357)
	viper.SetDefault("target", "default.target")
	viper.SetDefault("paths", []string{
		"/etc/systemd/system", "/lib/systemd/system",
	})
	viper.SetDefault("debug", false)

	if err := viper.Unmarshal(&Conf); err != nil {
		log.Errorf("Error parsing %s: %s", viper.ConfigFileUsed(), err)
	}

	if Conf.Debug {
		log.SetLevel(log.DebugLevel)
	}
}

// Boot initializes the system, sets the default paths, as specified in configuration and attempts to start the default target, falls back to "rescue.target", if it fails
func Boot() {
	// Initialize system
	System = system.New()
	System.SetPaths(Conf.Paths...)

	// Start the default target
	if err := System.Start(Conf.Target); err != nil {
		System.Log.Errorf("Error starting default target %s: %s", Conf.Target, err)
		if err = System.Start("rescue.target"); err != nil {
			System.Log.Errorf("Error starting rescue target %s: %s", "rescue.target", err)
		}
	}
}

// Listen for systemctl requests
func Serve() {
	if err := listenHTTP(Conf.Port.String()); err != nil {
		log.Fatalf("Error starting server on %s: %s", Conf.Port, err)
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
