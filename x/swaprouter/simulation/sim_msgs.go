package swaproutersimulation

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	legacysimulationtype "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/osmosis-labs/osmosis/v13/simulation/simtypes"
	"github.com/osmosis-labs/osmosis/v13/x/gamm/pool-models/balancer"
	balancerv2 "github.com/osmosis-labs/osmosis/v13/x/gamm/pool-models/balancer/v2"
	"github.com/osmosis-labs/osmosis/v13/x/swaprouter"
	"github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
)

// simulationKeeper is a wrapper around swaprouter's keeper which makes it easy to wire new keepers
// nolint: unused
type simulationKeeper struct {
	keeper swaprouter.Keeper

	gammKeeper types.GammKeeper
}

// PoolCreationFee denotes how much it costs to create a pool.
var PoolCreationFee = sdk.NewInt64Coin(sdk.DefaultBondDenom, 10_000_000)

// RandomCreatePoolMsg attempts to find an account with two or more distinct denoms and attempts to send a
// create pool message composed of those denoms
func RandomCreateUniV2Msg(k swaprouter.Keeper, sim *simtypes.SimCtx, ctx sdk.Context) (*balancerv2.MsgCreateBalancerPool, error) {
	var poolAssets []balancer.PoolAsset
	// find an address with two or more distinct denoms in their wallet
	sender, senderExists := sim.RandomSimAccountWithConstraint(createPoolRestriction(k, sim, ctx))
	if !senderExists {
		return nil, errors.New("no sender with two different denoms & pool creation fee exists")
	}
	poolCoins, ok := sim.GetRandSubsetOfKDenoms(ctx, sender, 2)
	if !ok {
		return nil, fmt.Errorf("provided sender with requested number of denoms does not exist")
	}
	if poolCoins.Add(PoolCreationFee).IsAnyGT(sim.BankKeeper().SpendableCoins(ctx, sender.Address)) {
		return nil, errors.New("chose an account / creation amount that didn't pass fee bar")
	}

	// TODO: pseudo-randomly generate swap and exit fees
	poolParams := &balancer.PoolParams{
		SwapFee: sdk.NewDecWithPrec(1, 2),
		ExitFee: sdk.ZeroDec(),
	}

	// from the above selected account, determine the token type and respective weight needed to make the pool
	for i := 0; i < len(poolCoins); i++ {
		poolAssets = append(poolAssets, balancer.PoolAsset{
			Weight: sdk.OneInt(),
			Token:  poolCoins[i],
		})
	}

	return &balancerv2.MsgCreateBalancerPool{
		Sender:     sender.Address.String(),
		PoolParams: poolParams,
		PoolAssets: poolAssets,
	}, nil
}

func createPoolRestriction(k swaprouter.Keeper, sim *simtypes.SimCtx, ctx sdk.Context) simtypes.SimAccountConstraint {
	return func(acc legacysimulationtype.Account) bool {
		accCoins := sim.BankKeeper().SpendableCoins(ctx, acc.Address)
		hasTwoCoins := len(accCoins) >= 2
		hasPoolCreationFee := accCoins.AmountOf("stake").GT(PoolCreationFee.Amount)
		return hasTwoCoins && hasPoolCreationFee
	}
}
