// Example plugin client, demonstrating how to create the client,
// start the plugin, and call the plugin's methods.
//
// This example shows how to integrate the srelib plugin into a client application.
// The client launches the plugin server as a subprocess, establishes an RPC connection,
// and calls methods on the plugin interface.
//
// Usage:
//
//	srelib-client <cluster-id>
//
// Prerequisites:
//   - OCM credentials configured at ~/.config/ocm/ocm.json
//   - Plugin binary (srelib-plugin) in the same directory as the client
//
// The client will:
//  1. Launch the srelib-plugin server process
//  2. Establish an RPC connection via HashiCorp go-plugin
//  3. Call GetCluster() with the provided cluster ID
//  4. Display cluster information or error messages
//  5. Gracefully terminate the plugin process
package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"

	"github.com/petrkotas/srelib/sdk"
	v1 "github.com/petrkotas/srelib/sdk/v1"
)

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "srelib-client-example",
		Level: hclog.Info,
	})

	clusterID := "test-cluster-id"
	if len(os.Args) > 1 {
		clusterID = os.Args[1]
	}

	fmt.Printf("SRELib Plugin Client Example\n")
	fmt.Printf("=============================\n")
	fmt.Printf("Attempting to fetch cluster: %s\n", clusterID)
	if len(os.Args) == 1 {
		fmt.Println("(Provide cluster ID as argument: ./srelib-client <cluster-id>)")
	}
	fmt.Println()

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: sdk.HandshakeConfig,
		VersionedPlugins: map[int]plugin.PluginSet{
			1: {
				"srelib": &v1.Plugin{},
			},
		},
		Cmd:    exec.Command("./srelib-plugin"),
		Logger: logger,
	})
	defer client.Kill()

	rpcClient, err := client.Client()
	if err != nil {
		logger.Error("Failed to connect to plugin", "error", err)
		fmt.Fprintf(os.Stderr, "Error: Could not start plugin: %v\n", err)
		fmt.Fprintln(os.Stderr, "\nTroubleshooting:")
		fmt.Fprintln(os.Stderr, "  - Ensure the plugin binary (srelib-plugin) is in the same directory")
		fmt.Fprintln(os.Stderr, "  - Ensure OCM credentials exist at ~/.config/ocm/ocm.json")
		fmt.Fprintln(os.Stderr, "  - Login with: ocm login")
		os.Exit(1)
	}

	raw, err := rpcClient.Dispense("srelib")
	if err != nil {
		logger.Error("Failed to dispense plugin", "error", err)
		fmt.Fprintf(os.Stderr, "Error: Plugin handshake failed: %v\n", err)
		fmt.Fprintln(os.Stderr, "\nThis usually indicates a version mismatch between client and plugin.")
		os.Exit(1)
	}

	srelibClient, ok := raw.(v1.Client)
	if !ok {
		logger.Error("Plugin has wrong type", "type", fmt.Sprintf("%T", raw))
		fmt.Fprintf(os.Stderr, "Error: Plugin returned unexpected type: %T\n", raw)
		os.Exit(1)
	}

	logger.Info("Plugin connected successfully", "cluster_id", clusterID)

	cluster, err := srelibClient.GetCluster(clusterID)
	if err != nil {
		logger.Error("Failed to get cluster", "error", err, "cluster_id", clusterID)
		fmt.Fprintf(os.Stderr, "Error: GetCluster failed: %v\n", err)
		fmt.Fprintln(os.Stderr, "\nTroubleshooting:")
		fmt.Fprintln(os.Stderr, "  - Ensure OCM credentials exist at ~/.config/ocm/ocm.json")
		fmt.Fprintln(os.Stderr, "  - Login with: ocm login")
		fmt.Fprintln(os.Stderr, "  - Verify the cluster ID is correct")
		os.Exit(1)
	}

	fmt.Printf("✓ Successfully retrieved cluster\n")
	fmt.Printf("\nCluster Details:\n")
	fmt.Printf("  ID:    %s\n", cluster.ID())
	fmt.Printf("  Name:  %s\n", cluster.Name())
	fmt.Printf("  State: %s\n", cluster.State())

	logger.Info("Client completed successfully")
}
