package main

import (
	"fmt"
	"io"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	serviceConfigSectionName = "app"
	watsonConfigSectionName  = "watson"
)

const (
	logLevelEntry  = "loglevel"
	logOutputEntry = "logoutput"
	logFormatEntry = "logformat"

	watsonRetrieveTimeoutEntry       = "retrieveTimeout"
	watsonRetrieveRetriesEntry       = "retrieveRetries"
	watsonNativeLangAPIEndpointEntry = "nativeLangAPIEndpoint"
)

// ServiceConfig is a base config for the service.
type ServiceConfig struct {
	LogLevel  log.Level
	LogOutput io.Writer
	LogFormat string
}

const (
	serviceLogLevelDefault  = "info"
	serviceLogOutputDefault = "stdout"
	serviceLogFormatDefault = "json"
)

func (sc *ServiceConfig) setDefaults() {
	viper.SetDefault(fmt.Sprintf("%s.%s", serviceConfigSectionName, logLevelEntry), serviceLogLevelDefault)
	viper.SetDefault(fmt.Sprintf("%s.%s", serviceConfigSectionName, logOutputEntry), serviceLogOutputDefault)
	viper.SetDefault(fmt.Sprintf("%s.%s", serviceConfigSectionName, logFormatEntry), serviceLogFormatDefault)
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

// WatsonConfig is a base config for the IBM Watson API endpoint connection.
type WatsonConfig struct {
	RetrieveTimeout time.Duration
	RetrieveRetries int

	NaturalLangAPIEndpoint string
}

const (
	watsonRetrieveTimeoutDefault       = 15 * time.Second
	watsonRetrieveRetriesDefault       = 2
	watsonNativeLangAPIEndpointDefault = "https://gateway.watsonplatform.net/natural-language-understanding/api/v1/analyze"
)

func (wc *WatsonConfig) setDefaults() {
	viper.SetDefault(fmt.Sprintf("%s.%s", watsonConfigSectionName, watsonRetrieveTimeoutEntry), watsonRetrieveTimeoutDefault)
	viper.SetDefault(fmt.Sprintf("%s.%s", watsonConfigSectionName, watsonRetrieveRetriesEntry), watsonRetrieveRetriesDefault)
	viper.SetDefault(fmt.Sprintf("%s.%s", watsonConfigSectionName, watsonNativeLangAPIEndpointEntry), watsonNativeLangAPIEndpointDefault)
}

func (wc *WatsonConfig) log() {
	log.Infoln(fmt.Sprintf("Service Watson %s:", watsonRetrieveTimeoutEntry), wc.RetrieveTimeout)
	log.Infoln(fmt.Sprintf("Service Watson %s:", watsonRetrieveRetriesEntry), wc.RetrieveRetries)
	log.Infoln(fmt.Sprintf("Service Watson %s:", watsonNativeLangAPIEndpointEntry), wc.NaturalLangAPIEndpoint)
}

// Config is a full config.
type Config struct {
	Service ServiceConfig
	Watson  WatsonConfig
}

// Log logs the settings stored in config.
func (c *Config) Log() {
	c.Service.log()
	c.Watson.log()
}

// complete sets default values on some fields
func (c *Config) setDefaults() {
	c.Service.setDefaults()
	c.Watson.setDefaults()
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
	config := Config{}
	config.setDefaults()

	serviceConfig := ServiceConfig{
		LogLevel:  translateLogLevel(viper.GetString(fmt.Sprintf("%s.%s", serviceConfigSectionName, logLevelEntry))),
		LogOutput: translateLogOutput(viper.GetString(fmt.Sprintf("%s.%s", serviceConfigSectionName, logOutputEntry))),
		LogFormat: viper.GetString(fmt.Sprintf("%s.%s", serviceConfigSectionName, logFormatEntry)),
	}
	watsonConfig := WatsonConfig{
		RetrieveTimeout:        viper.GetDuration(fmt.Sprintf("%s.%s", watsonConfigSectionName, watsonRetrieveTimeoutEntry)),
		RetrieveRetries:        viper.GetInt(fmt.Sprintf("%s.%s", watsonConfigSectionName, watsonRetrieveRetriesEntry)),
		NaturalLangAPIEndpoint: viper.GetString(fmt.Sprintf("%s.%s", watsonConfigSectionName, watsonNativeLangAPIEndpointEntry)),
	}

	config.Service = serviceConfig
	config.Watson = watsonConfig

	return config
}

func getConfig() Config {
	config := buildConfig()
	return config
}
