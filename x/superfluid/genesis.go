package superfluid

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)

	// initialize superfluid assets
	for _, asset := range genState.SuperfluidAssets {
		k.SetSuperfluidAsset(ctx, asset)
	}

	// initialize osmo equivalent multipliers
	for _, multiplierRecord := range genState.OsmoEquivalentMultipliers {
		k.SetOsmoEquivalentMultiplier(ctx, multiplierRecord.EpochNumber, multiplierRecord.Denom, multiplierRecord.Multiplier)
	}

	accountsToReplace := make(map[string]types.SuperfluidIntermediaryAccount)

	for _, intermediaryAcc := range genState.IntermediaryAccounts {
		fmt.Printf("INT GAUGEID %v", intermediaryAcc.GaugeId)
		fmt.Printf("INT DENOM %v", intermediaryAcc.Denom)

		if string(intermediaryAcc.ValAddr) == "osmovaloper12smx2wdlyttvyzvzg54y2vnqwq2qjatex7kgq4" {
			// This the account we are replacing
			oldAcc := types.SuperfluidIntermediaryAccount{
				Denom:   intermediaryAcc.Denom,
				ValAddr: "osmovaloper1cyw4vw20el8e7ez8080md0r8psg25n0cq98a9n",
				GaugeId: intermediaryAcc.GaugeId,
			}

			// Key is the old account address and value is the new account
			accountsToReplace[oldAcc.GetAccAddress().String()] = intermediaryAcc
			fmt.Printf("for old account %v caching new acc %v\n", oldAcc, intermediaryAcc)
		}

		k.SetIntermediaryAccount(ctx, intermediaryAcc)
	}

	// initialize lock id and intermediary connections
	for _, connection := range genState.IntemediaryAccountConnections {
		acc, err := sdk.AccAddressFromBech32(connection.IntermediaryAccount)
		if err != nil {
			panic(err)
		}
		intermediaryAcc := k.GetIntermediaryAccount(ctx, acc)

		// We state exported an old address that was created from hash(denom + old address)
		// So we want to replace it with a new address that replaced old
		if intermediaryAcc.Empty() {
			toReplace, ok := accountsToReplace[acc.String()]
			if !ok {
				panic(fmt.Sprintf("did not find %s in accountsToReplace", acc.String()))
			}
			intermediaryAcc = toReplace
		}

		if intermediaryAcc.Denom == "" {
			fmt.Printf("GAUGE ID %v\n", intermediaryAcc.GaugeId)
			fmt.Printf("VAL ADDR %v\n", intermediaryAcc.ValAddr)
			fmt.Printf("intermediaryAcc.GetValAddr()%v\n", intermediaryAcc.GetValAddr())
			fmt.Printf("intermediaryAcc.GetDenom()%v\n", intermediaryAcc.GetValAddr())
			fmt.Printf("intermediaryAcc.GetGaugeId()%v\n", intermediaryAcc.GetValAddr())
			panic("connection to invalid intermediary account found")
		}
		k.SetLockIdIntermediaryAccountConnection(ctx, connection.LockId, intermediaryAcc)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params:                        k.GetParams(ctx),
		SuperfluidAssets:              k.GetAllSuperfluidAssets(ctx),
		OsmoEquivalentMultipliers:     k.GetAllOsmoEquivalentMultipliers(ctx),
		IntermediaryAccounts:          k.GetAllIntermediaryAccounts(ctx),
		IntemediaryAccountConnections: k.GetAllLockIdIntermediaryAccountConnections(ctx),
	}
}
