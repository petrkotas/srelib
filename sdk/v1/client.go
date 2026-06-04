package v1

import (
	"net/rpc"
)

// =================================================
// Client implementation
// =================================================

type RPCClient struct {
	Client *rpc.Client
}

// Function calls
