package testnetify

import (
	"bytes"

	"github.com/osmosis-labs/osmosis/v8/app"
)

func replaceConsAddrHex(genesis app.GenesisState, fromAddr string, replaceAddr string) {
	// TODO: When optimizing, do single in-place mutation
	replaceAllInGenesis(genesis, fromAddr, replaceAddr)
}

// TODO: Make it possible to only replace all in certain genesis keys
// no reason to pay for iterating over bank every time :)
func replaceAllInGenesis(genesis app.GenesisState, find string, replace string) {
	// To help in debugging
	findBz := []byte(find)
	replaceBz := []byte(replace)
	for k := range genesis {
		value := genesis[k]
		value = bytes.ReplaceAll(value, findBz, replaceBz)
		genesis[k] = value
	}
}
