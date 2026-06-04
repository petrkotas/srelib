package sdk

import (
	"github.com/hashicorp/go-plugin"
)

// Setup constants for the plugin handshake. These can be hardcoded as they are not
// considered sensitive information. The plugin will reject any connection that does not
// use these exact values. Additionally the plugin handshake is considered usability
// feature not security feature.
const (
	MagicCookieKey   = "SRELIB"
	MagicCookieValue = "5ca2b5decf594c7dc220935342fe31873eb6dd0b875d45079a44dc1a5b8c3be4"
)

// HandshakeConfigs are used to just do a basic handshake between a plugin and host.
// If the handshake fails, a user friendly error is shown.
// This prevents users from executing bad plugins or executing a plugin in the wrong process.
// This is not a security measure, but it is a good sanity check.
var HandshakeConfig = plugin.HandshakeConfig{
	MagicCookieKey:   MagicCookieKey,
	MagicCookieValue: MagicCookieValue,
}
