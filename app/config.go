package app

import (
	"fmt"
	"time"

	dbm "github.com/cosmos/cosmos-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sims "github.com/cosmos/cosmos-sdk/testutil/sims"

	pruningtypes "cosmossdk.io/store/pruning/types"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v28/app/keepers"
)

type OTELConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	ServiceName string `mapstructure:"service-name"`
}

// DefaultConfig returns a default configuration suitable for nearly all
// testing requirements.
func DefaultConfig() network.Config {
	encCfg := MakeEncodingConfig()

	return network.Config{
		Codec:             encCfg.Marshaler,
		TxConfig:          encCfg.TxConfig,
		LegacyAmino:       encCfg.Amino,
		InterfaceRegistry: encCfg.InterfaceRegistry,
		AccountRetriever:  authtypes.AccountRetriever{},
		AppConstructor:    NewAppConstructor("osmosis-code-test"),
		GenesisState:      keepers.AppModuleBasics.DefaultGenesis(encCfg.Marshaler),
		TimeoutCommit:     1 * time.Second / 2,
		ChainID:           "osmosis-code-test",
		NumValidators:     1,
		BondDenom:         sdk.DefaultBondDenom,
		MinGasPrices:      fmt.Sprintf("0.000006%s", sdk.DefaultBondDenom),
		AccountTokens:     sdk.TokensFromConsensusPower(1000, sdk.DefaultPowerReduction),
		StakingTokens:     sdk.TokensFromConsensusPower(500, sdk.DefaultPowerReduction),
		BondedTokens:      sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction),
		PruningStrategy:   pruningtypes.PruningOptionNothing,
		CleanupDir:        true,
		SigningAlgo:       string(hd.Secp256k1Type),
		KeyringOptions:    []keyring.Option{},
	}
}

// NewAppConstructor returns a new Osmosis app given encoding type configs.
func NewAppConstructor(chainId string) network.AppConstructor {
	return func(val network.ValidatorI) servertypes.Application {
		valCtx := val.GetCtx()
		appConfig := val.GetAppConfig()

		return NewOsmosisApp(
			valCtx.Logger, dbm.NewMemDB(), nil, true, make(map[int64]bool), valCtx.Config.RootDir, 0,
			sims.EmptyAppOptions{},
			EmptyWasmOpts,
			baseapp.SetMinGasPrices(appConfig.MinGasPrices),
			baseapp.SetChainID(chainId),
		)
	}
}

// NewOTELConfigFromOptions returns a new DataDog config from the given options.
func NewOTELConfigFromOptions(opts servertypes.AppOptions) OTELConfig {
	isEnabled := osmoutils.ParseBool(opts, "otel", "enabled", false)

	if !isEnabled {
		return OTELConfig{
			Enabled: false,
		}
	}

	serviceName := osmoutils.ParseString(opts, "otel", "service-name")

	return OTELConfig{
		Enabled:     isEnabled,
		ServiceName: serviceName,
	}
}
