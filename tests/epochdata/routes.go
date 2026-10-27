// Package epochdata provides a recorded mainnet routing workload for keeper benchmarks.
package epochdata

import (
	_ "embed"
	"encoding/json"
	"testing"
)

//go:embed epoch_routes.json
var recordedRoutes []byte

type Route struct {
	Base   string `json:"base"`
	Quote  string `json:"quote"`
	PoolID uint64 `json:"pool_id"`
}

type Request struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

type Workload struct {
	Height         int64     `json:"height"`
	SwapHops       int       `json:"swap_hops"`
	Intermediaries []string  `json:"intermediaries"`
	Routes         []Route   `json:"routes"`
	Requests       []Request `json:"requests"`
}

func Load(tb testing.TB) Workload {
	tb.Helper()
	var workload Workload
	if err := json.Unmarshal(recordedRoutes, &workload); err != nil {
		tb.Fatal(err)
	}
	return workload
}
