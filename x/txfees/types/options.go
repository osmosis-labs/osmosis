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
	DefaultMinGasPriceForHighGasTx       = osmomath.ZeroDec()
	DefaultMaxGasWantedPerTx             = uint64(25 * 1000 * 1000)
	DefaultHighGasTxThreshold            = uint64(1 * 1000 * 1000)
	DefaultMempool1559Enabled            = false
	DefaultMempool1559                   = osmomath.MustNewDecFromStr("0.025")
	DefaultMempool1559Min                = osmomath.MustNewDecFromStr("0.025")
	DefaultMempool1559Max                = osmomath.MustNewDecFromStr("10")
	DefaultMempool1559TargetGas          = int64(60_000_000)
	DefaultMempool1559MaxBlockChangeRate = osmomath.NewDec(1).Quo(osmomath.NewDec(16))
	DefaultMempool1559ResetInterval      = int64(1000)
	DefaultRecheckFeeConstant            = int64(4)
)

var (
	GlobalMempool1559Enabled            = false
	GlobalMempool1559                   = osmomath.MustNewDecFromStr("0.025")
	GlobalMempool1559Min                = osmomath.MustNewDecFromStr("0.025")
	GlobalMempool1559Max                = osmomath.MustNewDecFromStr("10")
	GlobalMempool1559MaxBlockChangeRate = osmomath.NewDec(1).Quo(osmomath.NewDec(16))
)

type MempoolFeeOptions struct {
	MaxGasWantedPerTx             uint64
	MinGasPriceForArbitrageTx     osmomath.Dec
	HighGasTxThreshold            uint64
	MinGasPriceForHighGasTx       osmomath.Dec
	Mempool1559Enabled            bool
	Mempool1559Default            osmomath.Dec
	Mempool1559Min                osmomath.Dec
	Mempool1559Max                osmomath.Dec
	Mempool1559TargetGas          int64
	Mempool1559MaxBlockChangeRate osmomath.Dec
	Mempool1559ResetInterval      int64
	Mempool1559RecheckFeeConstant int64
}

func NewDefaultMempoolFeeOptions() MempoolFeeOptions {
	return MempoolFeeOptions{
		MaxGasWantedPerTx:             DefaultMaxGasWantedPerTx,
		MinGasPriceForArbitrageTx:     DefaultMinGasPriceForArbitrageTx.Clone(),
		HighGasTxThreshold:            DefaultHighGasTxThreshold,
		MinGasPriceForHighGasTx:       DefaultMinGasPriceForHighGasTx.Clone(),
		Mempool1559Enabled:            DefaultMempool1559Enabled,
		Mempool1559Default:            DefaultMempool1559.Clone(),
		Mempool1559Min:                DefaultMempool1559Min.Clone(),
		Mempool1559Max:                DefaultMempool1559Max.Clone(),
		Mempool1559TargetGas:          DefaultMempool1559TargetGas,
		Mempool1559MaxBlockChangeRate: DefaultMempool1559MaxBlockChangeRate.Clone(),
		Mempool1559ResetInterval:      DefaultMempool1559ResetInterval,
		Mempool1559RecheckFeeConstant: DefaultRecheckFeeConstant,
	}
}

func NewMempoolFeeOptions(opts servertypes.AppOptions) MempoolFeeOptions {
	return MempoolFeeOptions{
		MaxGasWantedPerTx:             parseMaxGasWantedPerTx(opts),
		MinGasPriceForArbitrageTx:     parseMinGasPriceForArbitrageTx(opts),
		HighGasTxThreshold:            DefaultHighGasTxThreshold,
		MinGasPriceForHighGasTx:       parseMinGasPriceForHighGasTx(opts),
		Mempool1559Enabled:            parseMempool1559(opts),
		Mempool1559Default:            parseMempool1559Default(opts),
		Mempool1559Min:                parseMempool1559Min(opts),
		Mempool1559Max:                parseMempool1559Max(opts),
		Mempool1559TargetGas:          parseMempool1559TargetGas(opts),
		Mempool1559MaxBlockChangeRate: parseMempool1559MaxBlockChangeRate(opts),
		Mempool1559ResetInterval:      parseMempool1559ResetInterval(opts),
		Mempool1559RecheckFeeConstant: parseMempool1559RecheckFeeConstant(opts),
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

func parseMempool1559Default(opts servertypes.AppOptions) osmomath.Dec {
	GlobalMempool1559 = parseDecFromConfig(opts, "adaptive-fee-default", DefaultMempool1559)
	return GlobalMempool1559
}

func parseMempool1559Min(opts servertypes.AppOptions) osmomath.Dec {
	GlobalMempool1559Min = parseDecFromConfig(opts, "adaptive-fee-min", DefaultMempool1559Min)
	GlobalMempool1559 = parseDecFromConfig(opts, "adaptive-fee-default", DefaultMempool1559)
	if GlobalMempool1559.LT(GlobalMempool1559Min) {
		panic(fmt.Errorf("invalidly configured osmosis-mempool.adaptive-fee-default, must be greater than or equal to osmosis-mempool.adaptive-fee-min"))
	}
	return GlobalMempool1559Min
}

func parseMempool1559Max(opts servertypes.AppOptions) osmomath.Dec {
	GlobalMempool1559Max = parseDecFromConfig(opts, "adaptive-fee-max", DefaultMempool1559Max)
	return GlobalMempool1559Max
}

func parseMempool1559TargetGas(opts servertypes.AppOptions) int64 {
	valueInterface := opts.Get("osmosis-mempool.adaptive-fee-target-gas")
	if valueInterface == nil {
		return DefaultMempool1559TargetGas
	}
	value, err := cast.ToInt64E(valueInterface)
	if err != nil {
		panic("invalidly configured osmosis-mempool.adaptive-fee-target-gas")
	}
	return value
}

func parseMempool1559MaxBlockChangeRate(opts servertypes.AppOptions) osmomath.Dec {
	GlobalMempool1559MaxBlockChangeRate = parseDecFromConfig(opts, "adaptive-fee-max-block-change-rate", DefaultMempool1559Max)
	return GlobalMempool1559MaxBlockChangeRate
}

func parseMempool1559ResetInterval(opts servertypes.AppOptions) int64 {
	valueInterface := opts.Get("osmosis-mempool.adaptive-fee-reset-interval")
	if valueInterface == nil {
		return DefaultMempool1559ResetInterval
	}
	value, err := cast.ToInt64E(valueInterface)
	if err != nil {
		panic("invalidly configured osmosis-mempool.adaptive-fee-reset-interval")
	}
	return value
}

func parseMempool1559RecheckFeeConstant(opts servertypes.AppOptions) int64 {
	valueInterface := opts.Get("osmosis-mempool.adaptive-fee-recheck-fee-constant")
	if valueInterface == nil {
		return DefaultRecheckFeeConstant
	}
	value, err := cast.ToInt64E(valueInterface)
	if err != nil {
		panic("invalidly configured osmosis-mempool.adaptive-fee-recheck-fee-constant")
	}
	return value
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
