package main

import (
	"fmt"
	"io"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	serviceConfigSectionName = "app"
)

const (
	logLevelEntry  = "loglevel"
	logOutputEntry = "logoutput"
	logFormatEntry = "logformat"
)

// ServiceConfig is a base config for the service.
type ServiceConfig struct {
	LogLevel  log.Level
	LogOutput io.Writer
	LogFormat string
}

func (sc *ServiceConfig) log() {
	log.Infoln("Service log level:", sc.LogLevel)
	var output string
	if sc.LogOutput == os.Stderr {
		output = "stderr"
	} else {
		output = "stdout"
	}
	log.Infoln("Service log output:", output)
	log.Infoln("Service log format:", sc.LogFormat)
}

// Log logs the settings stored in config.
func (c *Config) Log() {
	c.Service.log()
}

// Config is a full config.
type Config struct {
	Service ServiceConfig
}

const (
	serviceLogLevelDefault  = "info"
	serviceLogOutputDefault = "stdout"
	serviceLogFormatDefault = "json"
)

func setDefaults() {
	viper.SetDefault(fmt.Sprintf("%s.%s", serviceConfigSectionName, logLevelEntry), serviceLogLevelDefault)
	viper.SetDefault(fmt.Sprintf("%s.%s", serviceConfigSectionName, logOutputEntry), serviceLogOutputDefault)
	viper.SetDefault(fmt.Sprintf("%s.%s", serviceConfigSectionName, logFormatEntry), serviceLogFormatDefault)
}

func translateLogLevel(level string) log.Level {
	lvl, err := log.ParseLevel(level)
	if err != nil {
		log.Warn("Uknown log level set in config. Setting up log level to DEBUG.")
		return log.DebugLevel
	}
	return lvl
}

func translateLogOutput(out string) io.Writer {
	if out == "stderr" {
		return os.Stderr
	}
	return os.Stdout
}

func buildConfig() Config {
	logLevel := translateLogLevel(viper.GetString(fmt.Sprintf("%s.%s", serviceConfigSectionName, logLevelEntry)))
	logOutput := translateLogOutput(viper.GetString(fmt.Sprintf("%s.%s", serviceConfigSectionName, logOutputEntry)))
	return Config{
		Service: ServiceConfig{
			LogLevel:  logLevel,
			LogOutput: logOutput,
			LogFormat: viper.GetString(fmt.Sprintf("%s.%s", serviceConfigSectionName, logFormatEntry)),
		},
	}
}

func getConfig() Config {
	// set defaults first
	setDefaults()

	config := buildConfig()
	return config
}
