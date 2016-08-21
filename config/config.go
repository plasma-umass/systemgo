package config

import (
	"fmt"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/plasma-umass/systemgo/system"
	"github.com/spf13/viper"
)

const (
	DEFAULT_PORT   = 8008
	DEFAULT_TARGET = "default.target"
	RESCUE_TARGET  = "rescue.target"
)

var (
	// Default target
	Target string

	// Paths to search for unit files
	Paths []string

	// Port for system daemon to listen on
	Port port

	// Retry specifies the period(in seconds) to wait before
	// restarting the http service if it fails
	Retry time.Duration

	// Wheter to show debugging statements
	Debug bool
)

type port int

func (p port) String() string {
	return fmt.Sprintf(":%v", int(p))
}

func init() {
	viper.SetDefault("port", DEFAULT_PORT)
	viper.SetDefault("target", DEFAULT_TARGET)
	viper.SetDefault("paths", system.DEFAULT_PATHS)
	viper.SetDefault("retry", 1)
	viper.SetDefault("debug", false)

	viper.SetEnvPrefix("systemgo")
	viper.AutomaticEnv()

	viper.SetConfigName("systemgo")
	viper.SetConfigType("yaml")

	viper.AddConfigPath(".")
	if os.Getenv("XDG_CONFIG_HOME") != "" {
		viper.AddConfigPath("$XDG_CONFIG_HOME/systemgo")
	}
	viper.AddConfigPath("/etc/systemgo")

	if err := viper.ReadInConfig(); err != nil {
		if os.IsNotExist(err) {
			log.Warn("Config file not found, using defaults")
		} else {
			log.WithFields(log.Fields{
				"file": viper.ConfigFileUsed(),
				"err":  err,
			}).Errorf("Parse error, using defaults")
		}
	} else {
		log.WithFields(log.Fields{
			"file": viper.ConfigFileUsed(),
		}).Infof("Found configuration file")
	}

	Target = viper.GetString("target")
	Paths = viper.GetStringSlice("paths")
	Port = port(viper.GetInt("port"))
	Retry = viper.GetDuration("retry") * time.Second
	Debug = viper.GetBool("debug")

	if Debug {
		log.SetLevel(log.DebugLevel)
	}
}
