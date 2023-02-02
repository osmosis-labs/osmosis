package simulation

import (
	"errors"
	"fmt"
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"

	legacysimulationtype "github.com/cosmos/cosmos-sdk/types/simulation"

	osmosimtypes "github.com/osmosis-labs/osmosis/v14/simulation/simtypes"
	clkeeper "github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity"
	clmodeltypes "github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/model"
	cltypes "github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/types"
)

var PoolCreationFee = sdk.NewInt64Coin("stake", 10_000_000)

func RandomMsgCreateConcentratedPool(k clkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (*clmodeltypes.MsgCreateConcentratedPool, error) {
	minExponentAtOneValue := cltypes.ExponentAtPriceOneMin.Int64()
	maxExponentAtOneValue := cltypes.ExponentAtPriceOneMax.Int64()

	// generate random values from -13 to 1 (current accepted range: -12 to -1)
	exponentAtPriceOne := sdk.NewInt((minExponentAtOneValue + 2) + rand.Int63n((maxExponentAtOneValue-1)-(minExponentAtOneValue+2)+1))
	authorizedTickSpacing := cltypes.AuthorizedTickSpacing

	// find an address with two or more distinct denoms in their wallet
	sender, senderExists := sim.RandomSimAccountWithConstraint(createPoolRestriction(k, sim, ctx))
	if !senderExists {
		return nil, fmt.Errorf("no sender with two different denoms & pool creation fee exists")
	}

	// get random 3 coins, use 2 to create pool and 1 for fees (stake denom).
	poolCoins, ok := sim.GetRandSubsetOfKDenoms(ctx, sender, 3)
	if !ok {
		return nil, fmt.Errorf("provided sender with requested number of denoms does not exist")
	}

	// check if the sender has sufficient amount for fees
	if poolCoins.Add(PoolCreationFee).IsAnyGT(sim.BankKeeper().SpendableCoins(ctx, sender.Address)) {
		return nil, errors.New("chose an account / creation amount that didn't pass fee limit")
	}

	if poolCoins[0].Denom == sdk.DefaultBondDenom || poolCoins[1].Denom == sdk.DefaultBondDenom {
		return nil, fmt.Errorf("poolCoins contains denom stake which contains invalid metadata")
	}

	denom0 := poolCoins[0].Denom
	denom1 := poolCoins[1].Denom

	tickSpacing := authorizedTickSpacing[rand.Intn(len(authorizedTickSpacing))]

	return &clmodeltypes.MsgCreateConcentratedPool{
		Sender:                    sender.Address.String(),
		Denom0:                    denom0,
		Denom1:                    denom1,
		TickSpacing:               tickSpacing,
		PrecisionFactorAtPriceOne: exponentAtPriceOne,
		SwapFee:                   sdk.NewDecWithPrec(1, 2),
	}, nil
}
func RandMsgCreatePosition(k clkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (*cltypes.MsgCreatePosition, error) {
	// get random pool
	clPool, poolDenoms, err := getRandCLPool(k, sim, ctx)
	if err != nil {
		return nil, err
	}

	// get random user address with the pool denoms
	sender, tokens, senderExists := sim.SelAddrWithDenoms(ctx, poolDenoms)
	if !senderExists {
		return nil, fmt.Errorf("no sender with denoms %s exists", poolDenoms)
	}

	// ensure that we have atleast 2 tokens
	if len(tokens) < 2 {
		return nil, fmt.Errorf("not enough pool denoms tokens, require 2 got: %d", len(tokens))
	}

	// Randomize tick values from minTick to maxTick
	minTick, maxTick := clkeeper.GetMinAndMaxTicksFromExponentAtPriceOne(clPool.GetPrecisionFactorAtPriceOne())

	// Randomize lowerTick and upperTick to create position
	lowerTick, upperTick, err := getRandomTickPositions(minTick, maxTick, clPool.GetTickSpacing())
	if err != nil {
		return nil, err
	}

	return &cltypes.MsgCreatePosition{
		PoolId:        clPool.GetId(),
		Sender:        sender.Address.String(),
		LowerTick:     lowerTick,
		UpperTick:     upperTick,
		TokenDesired0: tokens[0],
		TokenDesired1: tokens[1],
		// TODO: Randomize TokenMinAmount0 and TokenMinAmount1 in next iteration
		TokenMinAmount0: sdk.NewInt(0),
		TokenMinAmount1: sdk.NewInt(0),
	}, nil
}

func RandMsgWithdrawPosition(k clkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (*cltypes.MsgWithdrawPosition, error) {
	// get random pool
	_, poolDenoms, err := getRandCLPool(k, sim, ctx)
	if err != nil {
		return nil, err
	}

	// get random user address with the pool denoms
	sender, _, senderExists := sim.SelAddrWithDenoms(ctx, poolDenoms)
	if !senderExists {
		return nil, fmt.Errorf("no sender with denoms %s exists", poolDenoms)
	}

	positions, err := k.GetUserPositions(ctx, sender.Address)
	if err != nil {
		return nil, fmt.Errorf("position does not exist")
	}

	if len(positions) == 0 {
		return nil, fmt.Errorf("user does not have any position")
	}

	// pick a random position
	randPosition := positions[rand.Intn(len(positions))]

	withdrawAmount := randPosition.Liquidity.TruncateInt()

	return &cltypes.MsgWithdrawPosition{
		PoolId:          randPosition.PoolId,
		Sender:          sender.Address.String(),
		LowerTick:       randPosition.LowerTick,
		UpperTick:       randPosition.UpperTick,
		LiquidityAmount: withdrawAmount,
	}, nil
}

func RandMsgCollectFees(k clkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (*cltypes.MsgCollectFees, error) {
	// get random pool
	_, poolDenoms, err := getRandCLPool(k, sim, ctx)
	if err != nil {
		return nil, err
	}

	// get random user address with the pool denoms
	sender, _, senderExists := sim.SelAddrWithDenoms(ctx, poolDenoms)
	if !senderExists {
		return nil, fmt.Errorf("no sender with denoms %s exists", poolDenoms)
	}

	positions, err := k.GetUserPositions(ctx, sender.Address)
	if err != nil {
		return nil, fmt.Errorf("position does not exist")
	}

	if len(positions) == 0 {
		return nil, fmt.Errorf("user does not have any position")
	}

	// pick a random position
	randPosition := positions[rand.Intn(len(positions))]

	return &cltypes.MsgCollectFees{
		PoolId:    randPosition.PoolId,
		Sender:    sender.Address.String(),
		LowerTick: randPosition.LowerTick,
		UpperTick: randPosition.UpperTick,
	}, nil
}

// createPoolRestriction creates specific restriction for the creation of a pool.
func createPoolRestriction(k clkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) osmosimtypes.SimAccountConstraint {
	return func(acc legacysimulationtype.Account) bool {
		accCoins := sim.BankKeeper().SpendableCoins(ctx, acc.Address)
		hasTwoCoins := len(accCoins) >= 3
		hasPoolCreationFee := accCoins.AmountOf(PoolCreationFee.Denom).GT(PoolCreationFee.Amount)
		return hasTwoCoins && hasPoolCreationFee
	}
}

// getRandCLPool gets a concnerated liquidity pool with its pool denoms.
func getRandCLPool(k clkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (cltypes.ConcentratedPoolExtension, []string, error) {
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

	concentratedPool, ok := poolI.(cltypes.ConcentratedPoolExtension)
	if !ok {
		return nil, nil, fmt.Errorf("interface conversion failed")
	}

	poolDenoms := []string{concentratedPool.GetToken0(), concentratedPool.GetToken1()}

	return concentratedPool, poolDenoms, err
}

// getRandomTickPositions returns random lowerTick and upperTick divisible by tickSpacing value.
func getRandomTickPositions(minTick, maxTick int64, tickSpacing uint64) (int64, int64, error) {
	// Calculate the number of values between minTick and maxTick that are divisible by tickSpacing
	lowerTickDivisible := ((maxTick - minTick) / int64(tickSpacing)) + 1
	// Return error if no values are found
	if lowerTickDivisible < 1 {
		return 0, 0, fmt.Errorf("lower tick divisible by tickspacing not found")
	}

	// random value between 0 and lowerTickDivisible
	lowerTick := rand.Int63n(lowerTickDivisible)*int64(tickSpacing) + minTick
	// Return error if lowerTick is not divisible by tickSpacing
	if lowerTick%int64(tickSpacing) != 0 {
		return 0, 0, fmt.Errorf("lower tick is not divisible by tickspacing %d", tickSpacing)
	}

	// Calculate the number of values between lowerTick and maxTick that are divisible by tickSpacing
	upperTickDivisible := ((maxTick - lowerTick) / int64(tickSpacing)) + 1
	// Return error if lowerTick is not divisible by tickSpacing
	if upperTickDivisible < 1 {
		return 0, 0, fmt.Errorf("upper tick divisible by tickspacing not found")
	}
	// Calculate upperTick as a random value between 0 and upperTickDivisible
	upperTick := rand.Int63n(upperTickDivisible)*int64(tickSpacing) + lowerTick
	// Return error if upperTick is not divisible by tickSpacing
	if upperTick%int64(tickSpacing) != 0 {
		return 0, 0, fmt.Errorf("lower tick is not divisible by tickspacing %d", tickSpacing)
	}

	return lowerTick, upperTick, nil
}
