// This file just benchmarks export on a live node.
package cmd

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/tendermint/tendermint/config"
	tmcfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/log"
)

type basePrinter struct {
}

func (b basePrinter) Println(i ...interface{}) {
	fmt.Println(i...)
}

type localAppOpts struct {
	config *config.Config
}

func (a localAppOpts) Get(o string) interface{} {
	if o == flags.FlagHome {
		return a.config.RootDir
	}
	return nil
}

func BenchmarkExport(b *testing.B) {
	config := tmcfg.DefaultConfig()
	// manually adjust this based on server your on
	config.RootDir = "/root/.osmosisd"
	for i := 0; i < b.N; i++ {
		exportLogic(log.NewNopLogger(), basePrinter{}, localAppOpts{config}, config, createOsmosisAppAndExport, -1, []string{})
	}
}
