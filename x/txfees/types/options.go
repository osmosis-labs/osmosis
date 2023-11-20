package types

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cast"

	"github.com/osmosis-labs/osmosis/osmomath"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
)

// If Options are not set in a config somewhere,
// use defaults to preserve functionality with old node software

// DefaultMinGasPriceForArbitrageTx represents minimum gas price
// for arbitrage transactions.
var DefaultMinGasPriceForArbitrageTx = osmomath.ZeroDec()

var (
	DefaultMinGasPriceForHighGasTx = osmomath.ZeroDec()
	DefaultMaxGasWantedPerTx       = uint64(25 * 1000 * 1000)
	DefaultHighGasTxThreshold      = uint64(1 * 1000 * 1000)
	DefaultMempool1559Enabled      = true
)

var GlobalMempool1559Enabled = false

type MempoolFeeOptions struct {
	MaxGasWantedPerTx         uint64
	MinGasPriceForArbitrageTx osmomath.Dec
	HighGasTxThreshold        uint64
	MinGasPriceForHighGasTx   osmomath.Dec
	Mempool1559Enabled        bool
}

func NewDefaultMempoolFeeOptions() MempoolFeeOptions {
	return MempoolFeeOptions{
		MaxGasWantedPerTx:         DefaultMaxGasWantedPerTx,
		MinGasPriceForArbitrageTx: DefaultMinGasPriceForArbitrageTx.Clone(),
		HighGasTxThreshold:        DefaultHighGasTxThreshold,
		MinGasPriceForHighGasTx:   DefaultMinGasPriceForHighGasTx.Clone(),
		Mempool1559Enabled:        DefaultMempool1559Enabled,
	}
}

func NewMempoolFeeOptions(opts servertypes.AppOptions) MempoolFeeOptions {
	return MempoolFeeOptions{
		MaxGasWantedPerTx:         parseMaxGasWantedPerTx(opts),
		MinGasPriceForArbitrageTx: parseMinGasPriceForArbitrageTx(opts),
		HighGasTxThreshold:        DefaultHighGasTxThreshold,
		MinGasPriceForHighGasTx:   parseMinGasPriceForHighGasTx(opts),
		Mempool1559Enabled:        parseMempool1559(opts),
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

func parseMinGasPriceForArbitrageTx(opts servertypes.AppOptions) osmomath.Dec {
	return parseDecFromConfig(opts, "arbitrage-min-gas-fee", DefaultMinGasPriceForArbitrageTx.Clone())
}

func parseMinGasPriceForHighGasTx(opts servertypes.AppOptions) osmomath.Dec {
	return parseDecFromConfig(opts, "min-gas-price-for-high-gas-tx", DefaultMinGasPriceForHighGasTx.Clone())
}

func parseMempool1559(opts servertypes.AppOptions) bool {
	GlobalMempool1559Enabled = parseBoolFromConfig(opts, "adaptive-fee-enabled", DefaultMempool1559Enabled)
	return GlobalMempool1559Enabled
}

func parseDecFromConfig(opts servertypes.AppOptions, optName string, defaultValue osmomath.Dec) osmomath.Dec {
	valueInterface := opts.Get("osmosis-mempool." + optName)
	value := defaultValue
	if valueInterface != nil {
		valueStr, ok := valueInterface.(string)
		if !ok {
			panic("invalidly configured osmosis-mempool." + optName)
		}
		var err error
		// pre-pend 0 to allow the config to start with a decimal, e.g. ".01"
		value, err = osmomath.NewDecFromStr("0" + valueStr)
		if err != nil {
			panic(fmt.Errorf("invalidly configured osmosis-mempool.%v, err= %v", optName, err))
		}
	}
	return value
}

func parseBoolFromConfig(opts servertypes.AppOptions, optName string, defaultValue bool) bool {
	fullOptName := "osmosis-mempool." + optName
	valueInterface := opts.Get(fullOptName)
	value := defaultValue
	if valueInterface != nil {
		valueStr, ok := valueInterface.(string)
		if !ok {
			panic("invalidly configured osmosis-mempool." + optName)
		}
		valueStr = strings.TrimSpace(valueStr)
		v, err := strconv.ParseBool(valueStr)
		if err != nil {
			fmt.Println("error in parsing" + fullOptName + " as bool, setting to false")
			return false
		}
		return v
	}
	return value
}
