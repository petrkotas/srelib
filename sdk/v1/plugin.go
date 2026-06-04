package v1

import (
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

// =================================================
// Hashicorp go-plugin implementation
// Struct here is used by the plugin internaly
// =================================================

type Plugin struct {
	Impl Client
}

func (p *Plugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &RPCServer{Impl: p.Impl}, nil
}

func (p *Plugin) Client(_ *plugin.MuxBroker, rpcClient *rpc.Client) (interface{}, error) {
	return &RPCClient{Client: rpcClient}, nil
}
