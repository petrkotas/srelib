// The server.go file is the entry point for the SRELib plugin server.
// It initializes the plugin and starts serving it using HashiCorp's go-plugin framework.
// The server listens for incoming plugin requests and handles them according to the defined plugin interface.
//
// The logging level can be configured via the SRELIB_LOG_LEVEL environment variable,
// default log level is info. Supported log levels are: trace, debug, info, warn, error.
package main

import (
	"os"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"

	"github.com/petrkotas/srelib/internal/i1"
	"github.com/petrkotas/srelib/sdk"
	v1 "github.com/petrkotas/srelib/sdk/v1"
)

func getLogLevel() hclog.Level {
	levelStr := strings.ToLower(os.Getenv("SRELIB_LOG_LEVEL"))
	switch levelStr {
	case "trace":
		return hclog.Trace
	case "debug":
		return hclog.Debug
	case "info":
		return hclog.Info
	case "warn":
		return hclog.Warn
	case "error":
		return hclog.Error
	default:
		return hclog.Info
	}
}

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "srelib-plugin",
		Level: getLogLevel(),
	})

	srelib := &i1.Client{Logger: logger}

	plugin.Serve(&plugin.ServeConfig{
		Logger:          logger,
		HandshakeConfig: sdk.HandshakeConfig,
		VersionedPlugins: map[int]plugin.PluginSet{
			1: {
				"srelib": &v1.Plugin{Impl: srelib},
			},
		},
	})
}
