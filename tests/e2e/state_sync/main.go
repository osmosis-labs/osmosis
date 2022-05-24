package main

import (
	"flag"
)

func main() {
	var (
		stateSyncValidatorConfigDir string
		trustHeight                 int64
		trustHash                   string
	)

	flag.StringVar(&stateSyncValidatorConfigDir, "config-dir", "", "validator config dir")
	flag.Int64Var(&trustHeight, "trust-height", 0, "trust Height")
	flag.StringVar(&trustHash, "trust-hash", "", "trust hash")

	if err := configureNodeForStateSync(stateSyncValidatorConfigDir, trustHeight, trustHash); err != nil {
		panic(err)
	}
}
