// The main.go file is the entry point for the local SRELib binary.
// It allows running the SRE Library implementation directly without the plugin system,
// useful for local development and testing.
//
// The logging level can be configured via the SRELIB_LOG_LEVEL environment variable,
// default log level is info. Supported log levels are: trace, debug, info, warn, error.
//
// Usage:
//   srelib-local --list    List all available methods in the Client interface
//   srelib-local           Run the local binary (interactive mode)
package main

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/hashicorp/go-hclog"

	"github.com/petrkotas/srelib/internal/i1"
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

// listMethods prints all available methods from the v1.Client interface
func listMethods() {
	interfaceType := reflect.TypeOf((*v1.Client)(nil)).Elem()

	fmt.Println("SRELib v1 Client Interface - Available Methods:")
	fmt.Println("===============================================")

	if interfaceType.NumMethod() == 0 {
		fmt.Println("No methods defined yet in the v1.Client interface.")
		fmt.Println("\nThe interface is currently empty and will be extended with methods.")
		return
	}

	for i := 0; i < interfaceType.NumMethod(); i++ {
		method := interfaceType.Method(i)
		fmt.Printf("\n%d. %s\n", i+1, method.Name)

		// Print method signature
		methodType := method.Type

		// Input parameters
		var inputs []string
		for j := 0; j < methodType.NumIn(); j++ {
			inputs = append(inputs, methodType.In(j).String())
		}

		// Output parameters
		var outputs []string
		for j := 0; j < methodType.NumOut(); j++ {
			outputs = append(outputs, methodType.Out(j).String())
		}

		fmt.Printf("   Signature: %s(%s) (%s)\n",
			method.Name,
			strings.Join(inputs, ", "),
			strings.Join(outputs, ", "))
	}

	fmt.Println("\n===============================================")
	fmt.Printf("Total methods: %d\n", interfaceType.NumMethod())
}

func main() {
	listFlag := flag.Bool("list", false, "List all available methods in the Client interface")
	flag.Parse()

	if *listFlag {
		listMethods()
		return
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "srelib-local",
		Level:  getLogLevel(),
		Output: os.Stdout,
	})

	// Initialize the SRELib client directly without plugin layer
	srelib := &i1.Client{Logger: logger}

	logger.Info("SRELib local binary initialized", "client", srelib)

	// Example usage - expand this based on your needs
	fmt.Println("SRELib local binary is running")
	fmt.Println("Use --list to see all available methods")
	fmt.Println("\nThe Client interface will be extended with methods to call directly")
}
