package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
)

var (
	// Configuration of the default genesis state for the module.
	DefaultTokenPairArbRoutes = []TokenPairArbRoutes{}
	// Configure the initial base denoms used for cyclic route building. The order of the list of base
	// denoms is the order in which routes will be prioritized i.e. routes will be built and simulated in a
	// first come first serve basis that is based on the order of the base denoms.
	DefaultBaseDenoms = []BaseDenom{
		{
			Denom:    OsmosisDenomination,
			StepSize: osmomath.NewInt(1_000_000),
		},
	}
	DefaultPoolTypeInfo = InfoByPoolType{
		Balancer: BalancerPoolInfo{
			Weight: 2, // it takes around 2 ms to simulate and execute a balancer swap
		},
		Stable: StablePoolInfo{
			Weight: 5, // it takes around 5 ms to simulate and execute a stable swap
		},
		Concentrated: ConcentratedPoolInfo{
			Weight:          7, // it takes around 7 ms to simulate and execute a concentrated swap
			MaxTicksCrossed: 5,
		},
		Cosmwasm: CosmwasmPoolInfo{},
	}
	DefaultDaysSinceModuleGenesis    = uint64(0)
	DefaultDeveloperFees             = []sdk.Coin{}
	DefaultLatestBlockHeight         = uint64(0)
	DefaultDeveloperAddress          = ""
	DefaultMaxPoolPointsPerBlock     = uint64(100)
	DefaultMaxPoolPointsPerTx        = uint64(18)
	DefaultPoolPointsConsumedInBlock = uint64(0)
	DefaultProfits                   = []sdk.Coin{}
	DefaultCyclicArbTracker          = CyclicArbTracker{
		CyclicArb:                  sdk.Coins(nil),
		HeightAccountingStartsFrom: 0,
	}
)

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:                 DefaultParams(),
		TokenPairArbRoutes:     DefaultTokenPairArbRoutes,
		BaseDenoms:             DefaultBaseDenoms,
		InfoByPoolType:         DefaultPoolTypeInfo,
		DaysSinceModuleGenesis: DefaultDaysSinceModuleGenesis,
		DeveloperFees:          DefaultDeveloperFees,
		DeveloperAddress:       DefaultDeveloperAddress,
		LatestBlockHeight:      DefaultLatestBlockHeight,
		MaxPoolPointsPerBlock:  DefaultMaxPoolPointsPerBlock,
		MaxPoolPointsPerTx:     DefaultMaxPoolPointsPerTx,
		PointCountForBlock:     DefaultPoolPointsConsumedInBlock,
		Profits:                DefaultProfits,
		CyclicArbTracker:       &DefaultCyclicArbTracker,
	}
}

// Validate performs basic genesis state validation returning an error upon any failure.
func (gs GenesisState) Validate() error {
	// Validate the token pair arb routes
	if err := ValidateTokenPairArbRoutes(gs.TokenPairArbRoutes); err != nil {
		return err
	}

	// Validate the base denoms
	if err := ValidateBaseDenoms(gs.BaseDenoms); err != nil {
		return err
	}

	// Validate the pool type information
	if err := gs.InfoByPoolType.Validate(); err != nil {
		return err
	}

	// Validate the developer fees
	if err := ValidateDeveloperFees(gs.DeveloperFees); err != nil {
		return err
	}

	// Validate the developer address if it is set
	if gs.DeveloperAddress != "" {
		if _, err := sdk.AccAddressFromBech32(gs.DeveloperAddress); err != nil {
			return err
		}
	}

	// Validate the max pool points per block
	if err := ValidateMaxPoolPointsPerBlock(gs.MaxPoolPointsPerBlock); err != nil {
		return err
	}

	// Validate the max pool points per tx
	if err := ValidateMaxPoolPointsPerTx(gs.MaxPoolPointsPerTx); err != nil {
		return err
	}

	return gs.Params.Validate()
}

func init() {
	// no-op
}
