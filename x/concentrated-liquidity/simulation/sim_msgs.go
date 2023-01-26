package simulation

import (
	"errors"
	"fmt"
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"

	legacysimulationtype "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/osmosis-labs/osmosis/v14/simulation/simtypes"
	osmosimtypes "github.com/osmosis-labs/osmosis/v14/simulation/simtypes"
	clkeeper "github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity"
	clmodeltypes "github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/types"
	cltypes "github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/types"
)

var PoolCreationFee = sdk.NewInt64Coin("stake", 10_000_000)

func RandomMsgCreateConcentratedPool(k clkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (*clmodeltypes.MsgCreateConcentratedPool, error) {
	// generate random values from -13 to 1 (accepted range: -12 to -1)
	exponentAtPriceOne := sdk.NewInt(rand.Int63n(1+13) - 13)
	authorizedTickSpacing := []uint64{1, 10, 60, 200}

	// find an address with two or more distinct denoms in their wallet
	sender, senderExists := sim.RandomSimAccountWithConstraint(createPoolRestriction(k, sim, ctx))
	if !senderExists {
		return nil, fmt.Errorf("no sender with two different denoms & pool creation fee exists")
	}

	// generate 3 coins, use 2 to create pool and 1 for fees. ["stake" denom - contiains invalid metadata]
	poolCoins, ok := sim.GetRandSubsetOfKDenoms(ctx, sender, 3)
	if !ok {
		return nil, fmt.Errorf("provided sender with requested number of denoms does not exist")
	}

	// check if the sender has sufficient amount for fees
	if poolCoins.Add(PoolCreationFee).IsAnyGT(sim.BankKeeper().SpendableCoins(ctx, sender.Address)) {
		return nil, errors.New("chose an account / creation amount that didn't pass fee bar")
	}

	denom0 := poolCoins[0].Denom
	denom1 := poolCoins[1].Denom

	tickSpacing := authorizedTickSpacing[rand.Intn(len(authorizedTickSpacing))]
	precisionFactorAtPriceOne := exponentAtPriceOne

	return &clmodeltypes.MsgCreateConcentratedPool{
		Sender:                    sender.Address.String(),
		Denom0:                    denom0,
		Denom1:                    denom1,
		TickSpacing:               tickSpacing,
		PrecisionFactorAtPriceOne: precisionFactorAtPriceOne,
		SwapFee:                   sdk.NewDecWithPrec(1, 2),
	}, nil
}

func RandMsgCreatePosition(k clkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (*cltypes.MsgCreatePosition, error) {
	return nil, nil
}

func RandMsgWithdrawPosition(k clkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (*cltypes.MsgWithdrawPosition, error) {
	// PseudoCode:
	// get all pool positions for a specific pool (we can randomize which pool to use)
	// - this will include, position lower_tick, upper_tick, liquidityAmt
	// randomly select 1 pool position
	// get random withdraw liquidity from [0 to existing liqudityAmt]
	return &cltypes.MsgWithdrawPosition{}, nil
}

func RandMsgCollectFees(k clkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (*cltypes.MsgCollectFees, error) {
	return &cltypes.MsgCollectFees{}, nil
}

func createPoolRestriction(k clkeeper.Keeper, sim *simtypes.SimCtx, ctx sdk.Context) simtypes.SimAccountConstraint {
	return func(acc legacysimulationtype.Account) bool {
		accCoins := sim.BankKeeper().SpendableCoins(ctx, acc.Address)
		hasTwoCoins := len(accCoins) >= 2
		hasPoolCreationFee := accCoins.AmountOf(PoolCreationFee.Denom).GT(PoolCreationFee.Amount)
		return hasTwoCoins && hasPoolCreationFee
	}
}

// getRandCLPool gets a concnerated liquidity pool with its pool denoms.
func getRandCLPool(k clkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (types.ConcentratedPoolExtension, []string, error) {
	// get all pools
	clPools, err := k.GetAllPools(ctx)
	if err != nil {
		return nil, nil, err
	}

	numPools := len(clPools)
	if numPools == 0 {
		return nil, nil, fmt.Errorf("no pools created")
	}

	pool_id := clPools[rand.Intn(numPools)].GetId()

	// check if the pool exists
	poolI, err := k.GetPool(ctx, pool_id)
	if err != nil {
		return nil, nil, fmt.Errorf("Pool not found for id %d", pool_id)
	}

	concentratedPool := poolI.(cltypes.ConcentratedPoolExtension)
	poolDenoms := []string{concentratedPool.GetToken0(), concentratedPool.GetToken1()}

	return concentratedPool, poolDenoms, err
}

// getRandMinMaxTicks gets min and max tick range for a specific exponentAtPriceOne.
func getRandMinMaxTicks(exponentAtPriceOne sdk.Int) (minTick, maxTick int64) {
	// Randomize this value
	return clkeeper.GetMinAndMaxTicksFromExponentAtPriceOne(exponentAtPriceOne)
}
