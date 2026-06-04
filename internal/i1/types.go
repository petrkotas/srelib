package i1

import (
	"github.com/hashicorp/go-hclog"
)

// Client is the V1 implementation of the v1 Client interface.
type Client struct {
	// common logger used internally by the library
	Logger hclog.Logger
}
