package types

import (
	"fmt"

	"github.com/spf13/cast"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// If Options are not set in a config somewhere,
// use defaults to preserve functionality with old node software

// DefaultMinGasPriceForArbitrageTx represents minimum gas price
// for arbitrage transactions.
var DefaultMinGasPriceForArbitrageTx = sdk.ZeroDec()

var (
	DefaultMinGasPriceForHighGasTx = sdk.ZeroDec()
	DefaultMaxGasWantedPerTx       = uint64(25 * 1000 * 1000)
	DefaultHighGasTxThreshold      = uint64(1 * 1000 * 1000)
)

type MempoolFeeOptions struct {
	MaxGasWantedPerTx         uint64
	MinGasPriceForArbitrageTx sdk.Dec
	HighGasTxThreshold        uint64
	MinGasPriceForHighGasTx   sdk.Dec
}

func NewDefaultMempoolFeeOptions() MempoolFeeOptions {
	return MempoolFeeOptions{
		MaxGasWantedPerTx:         DefaultMaxGasWantedPerTx,
		MinGasPriceForArbitrageTx: DefaultMinGasPriceForArbitrageTx.Clone(),
		HighGasTxThreshold:        DefaultHighGasTxThreshold,
		MinGasPriceForHighGasTx:   DefaultMinGasPriceForHighGasTx.Clone(),
	}
}

func NewMempoolFeeOptions(opts servertypes.AppOptions) MempoolFeeOptions {
	return MempoolFeeOptions{
		MaxGasWantedPerTx:         parseMaxGasWantedPerTx(opts),
		MinGasPriceForArbitrageTx: parseMinGasPriceForArbitrageTx(opts),
		HighGasTxThreshold:        DefaultHighGasTxThreshold,
		MinGasPriceForHighGasTx:   parseMinGasPriceForHighGasTx(opts),
	}
}

func parseMaxGasWantedPerTx(opts servertypes.AppOptions) uint64 {
	valueInterface := opts.Get("osmosis-mempool.max-gas-wanted-per-tx")
	if valueInterface == nil {
		return DefaultMaxGasWantedPerTx
	}
	value, err := cast.ToUint64E(valueInterface)
	if err != nil {
		panic("invalidly configured osmosis-mempool.max-gas-wanted-per-tx")
	}
	return value
}

func parseMinGasPriceForArbitrageTx(opts servertypes.AppOptions) sdk.Dec {
	return parseDecFromConfig(opts, "arbitrage-min-gas-fee", DefaultMinGasPriceForArbitrageTx.Clone())
}

func parseMinGasPriceForHighGasTx(opts servertypes.AppOptions) sdk.Dec {
	return parseDecFromConfig(opts, "min-gas-price-for-high-gas-tx", DefaultMinGasPriceForHighGasTx.Clone())
}

func parseDecFromConfig(opts servertypes.AppOptions, optName string, defaultValue sdk.Dec) sdk.Dec {
	valueInterface := opts.Get("osmosis-mempool." + optName)
	value := defaultValue
	if valueInterface != nil {
		valueStr, ok := valueInterface.(string)
		if !ok {
			panic("invalidly configured osmosis-mempool." + optName)
		}
		var err error
		// pre-pend 0 to allow the config to start with a decimal, e.g. ".01"
		value, err = sdk.NewDecFromStr("0" + valueStr)
		if err != nil {
			panic(fmt.Errorf("invalidly configured osmosis-mempool.%v, err= %v", optName, err))
		}
	}
	return value
}
