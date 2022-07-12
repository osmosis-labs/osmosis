package chain

import (
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"

	"github.com/osmosis-labs/osmosis/v7/tests/e2e/initialization"
)

type NodeConfig struct {
	initialization.Node

	OperatorAddress string
	rpcClient       *rpchttp.HTTP
}
