package wasmbinding

import (
	"sync"
)

// StargateLayerBindings keeps whitelist and its deterministic
// response binding for stargate queries.
//
// The query can be multi-thread, so we have to use
// thread safe sync.Map instead map[string]bool.
var StargateLayerBindings sync.Map

func init() {

}
