// This file just benchmarks export on a live node.
package cmd

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"
	tmcfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/log"
)

type basePrinter struct {
}

func (b basePrinter) Println(i ...interface{}) {
	fmt.Println(i...)
}

func BenchmarkExport(b *testing.B) {
	config := tmcfg.DefaultConfig()
	for i := 0; i < b.N; i++ {
		exportLogic(log.NewNopLogger(), basePrinter{}, simapp.EmptyAppOptions{}, config, createOsmosisAppAndExport, -1, []string{})
	}
}
