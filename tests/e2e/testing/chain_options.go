package e2eTesting

import (
	math "cosmossdk.io/math"
	cmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	mintTypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/osmosis-labs/osmosis/v26/app"
)

// chainConfig is a TestChain config which can be adjusted using options.
type chainConfig struct {
	ValidatorsNum    int
	GenAccountsNum   int
	GenBalanceAmount string
	BondAmount       string
	LoggerEnabled    bool
	DefaultFeeAmt    string
	DummyTestAddr    bool
}

type (
	TestChainConfigOption func(cfg *chainConfig)

	TestChainGenesisOption func(cdc codec.Codec, genesis app.GenesisState)

	TestChainConsensusParamsOption func(params *cmproto.ConsensusParams)
)

// defaultChainConfig builds chain default config.
func defaultChainConfig() chainConfig {
	return chainConfig{
		ValidatorsNum:    1,
		GenAccountsNum:   5,
		GenBalanceAmount: sdk.DefaultPowerReduction.MulRaw(100).String(),
		BondAmount:       sdk.DefaultPowerReduction.MulRaw(1).String(),
		DefaultFeeAmt:    sdk.DefaultPowerReduction.QuoRaw(10).String(), // 0.1
		DummyTestAddr:    false,
	}
}

func WithValidatorsNum(num int) TestChainConfigOption {
	return func(cfg *chainConfig) {
		cfg.ValidatorsNum = num
	}
}

func WithDummyTestAddress() TestChainConfigOption {
	return func(cfg *chainConfig) {
		cfg.DummyTestAddr = true
	}
}

// WithGenAccounts sets the number of genesis accounts
func WithGenAccounts(num int) TestChainConfigOption {
	return func(cfg *chainConfig) {
		cfg.GenAccountsNum = num
	}
}

// WithGenDefaultCoinBalance sets the genesis account balance for the default token (stake).
func WithGenDefaultCoinBalance(amount string) TestChainConfigOption {
	return func(cfg *chainConfig) {
		cfg.GenBalanceAmount = amount
	}
}

// WithDefaultFeeAmount sets the default fee amount which is used for sending Msgs with no fee specified.
func WithDefaultFeeAmount(amount string) TestChainConfigOption {
	return func(cfg *chainConfig) {
		cfg.DefaultFeeAmt = amount
	}
}

// WithBondAmount sets the amount of coins to bond for each validator.
func WithBondAmount(amount string) TestChainConfigOption {
	return func(cfg *chainConfig) {
		cfg.BondAmount = amount
	}
}

// WithLogger enables the app console logger.
func WithLogger() TestChainConfigOption {
	return func(cfg *chainConfig) {
		cfg.LoggerEnabled = true
	}
}

// WithBlockGasLimit sets the block gas limit (not set by default).
func WithBlockGasLimit(gasLimit int64) TestChainConsensusParamsOption {
	return func(params *cmproto.ConsensusParams) {
		params.Block.MaxGas = gasLimit
	}
}

// WithMintParams sets x/mint inflation calculation parameters.
func WithMintParams(inflationMin, inflationMax math.LegacyDec, blocksPerYear uint64) TestChainGenesisOption {
	return func(cdc codec.Codec, genesis app.GenesisState) {
		var mintGenesis mintTypes.GenesisState
		cdc.MustUnmarshalJSON(genesis[mintTypes.ModuleName], &mintGenesis)

		mintGenesis.Params.InflationMin = inflationMin
		mintGenesis.Params.InflationMax = inflationMax
		mintGenesis.Params.BlocksPerYear = blocksPerYear

		genesis[mintTypes.ModuleName] = cdc.MustMarshalJSON(&mintGenesis)
	}
}
