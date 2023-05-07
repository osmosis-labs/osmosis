package model

import (
	"encoding/json"
)

// String returns the json marshalled string of the pool
func (p CosmWasmPool) String() string {
	out, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}
	return string(out)
}
