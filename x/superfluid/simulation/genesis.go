package simulation

import (
	"encoding/json"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
)

// RandomizedGenState generates a random GenesisState for staking
func RandomizedGenState(simState *module.SimulationState) {
	superfluidGenesis := &types.GenesisState{
		Params: types.Params{
			RefreshEpochIdentifier: "second",
			MinimumRiskFactor:      sdk.NewDecWithPrec(5, 2), // 5%
			UnbondingDuration:      time.Second * 10,
		},
		SuperfluidAssets: []types.SuperfluidAsset{},
		TwapPriceRecords: []types.EpochOsmoEquivalentTWAP{},
	}

	bz, err := json.MarshalIndent(&superfluidGenesis.Params, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated superfluid parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(superfluidGenesis)
}
