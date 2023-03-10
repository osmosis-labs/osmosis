package simulation

import (
	"errors"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	legacysimulationtype "github.com/cosmos/cosmos-sdk/types/simulation"

	osmosimtypes "github.com/osmosis-labs/osmosis/v15/simulation/simtypes"
	clkeeper "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity"
	clmodeltypes "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	cltypes "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

var PoolCreationFee = sdk.NewInt64Coin("stake", 10_000_000)

func RandomMsgCreateConcentratedPool(k clkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (*clmodeltypes.MsgCreateConcentratedPool, error) {
	rand := sim.GetRand()
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

	// ensure that we always have 2 tokens
	// Note: tokens returns a random subset of poolDenoms, so  had to add this assertion
	if len(tokens) < 2 {
		return nil, fmt.Errorf("user doesnot have pool tokens")
	}

	//  Retrieve minTick and maxTick from precision factor
	minTick, maxTick := clkeeper.GetMinAndMaxTicksFromExponentAtPriceOne(clPool.GetPrecisionFactorAtPriceOne())

	// Randomize lowerTick and upperTick from max values to create position
	lowerTick, upperTick, err := getRandomTickPositions(sim, minTick, maxTick, clPool.GetTickSpacing())
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
		FreezeDuration:  time.Hour * 24,
	}, nil
}

func RandMsgWithdrawPosition(k clkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (*cltypes.MsgWithdrawPosition, error) {
	rand := sim.GetRand()
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

	positions, err := k.GetUserPositions(ctx, sender.Address, 0)
	if err != nil {
		return nil, fmt.Errorf("position does not exist")
	}

	if len(positions) == 0 {
		return nil, fmt.Errorf("user does not have any position")
	}

	// pick a random position
	randPosition := positions[rand.Intn(len(positions))]

	// check if the position is still frozen
	if randPosition.JoinTime.Add(randPosition.FreezeDuration).After(ctx.BlockTime()) {
		return nil, fmt.Errorf("position is still frozen")
	}

	// get percentage amount from 1 to 100 to withdraw liquidity
	randPerc := sim.RandomDecAmount(sdk.OneDec())

	withdrawAmountInt := randPosition.Liquidity.Mul(randPerc)

	return &cltypes.MsgWithdrawPosition{
		PoolId:          randPosition.PoolId,
		Sender:          sender.Address.String(),
		LowerTick:       randPosition.LowerTick,
		UpperTick:       randPosition.UpperTick,
		LiquidityAmount: withdrawAmountInt,
		JoinTime:        randPosition.JoinTime,
		FreezeDuration:  randPosition.FreezeDuration,
	}, nil
}

func RandMsgCollectFees(k clkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (*cltypes.MsgCollectFees, error) {
	rand := sim.GetRand()
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

	positions, err := k.GetUserPositions(ctx, sender.Address, 0)
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

// getRandCLPool gets a concentrated liquidity pool with its pool denoms.
func getRandCLPool(k clkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (cltypes.ConcentratedPoolExtension, []string, error) {
	rand := sim.GetRand()

	// get all pools
	clPools, err := k.GetAllPools(ctx)
	if err != nil {
		return nil, nil, err
	}

	numPools := len(clPools)
	if numPools == 0 {
		return nil, nil, fmt.Errorf("no pools created")
	}

	randConcentratedPool := clPools[rand.Intn(numPools)]
	poolDenoms := []string{randConcentratedPool.GetToken0(), randConcentratedPool.GetToken1()}

	return randConcentratedPool, poolDenoms, err
}

// getRandomTickPositions returns random lowerTick and upperTick divisible by tickSpacing value.
func getRandomTickPositions(sim *osmosimtypes.SimCtx, minTick, maxTick int64, tickSpacing uint64) (int64, int64, error) {
	lowerTick, err := RandomTickDivisibility(sim, minTick, maxTick, tickSpacing)
	if err != nil {
		return 0, 0, err
	}

	if lowerTick == -1 {
		return 0, 0, fmt.Errorf("random lowertick divisible by tickSpacing not found")
	}

	upperTick, err := RandomTickDivisibility(sim, lowerTick, maxTick, tickSpacing)
	if err != nil {
		return 0, 0, err
	}

	if upperTick == -1 {
		return 0, 0, fmt.Errorf("random lowertick divisible by tickSpacing not found")
	}

	if lowerTick == upperTick {
		return 0, 0, fmt.Errorf("lower tick and upper tick cannot be the same")
	}

	return lowerTick, upperTick, nil
}

// RandomTickDivisibility calculates a random number between minTick - maxTick (inclusive) that is divisible by tickSpacing
func RandomTickDivisibility(sim *osmosimtypes.SimCtx, minTick int64, maxTick int64, tickSpacing uint64) (int64, error) {
	rand := sim.GetRand()

	// Generate a random number in the range [minTick, maxTick]
	randomNumber := rand.Int63n(maxTick-minTick+1) + minTick

	// Find the next multiple of x that is greater than or equal to the random number
	nextMultiple := ((randomNumber + int64(tickSpacing) - 1) / int64(tickSpacing)) * int64(tickSpacing)

	// If the next multiple is within the range [a, b], return it
	if nextMultiple >= minTick && nextMultiple <= maxTick {
		return nextMultiple, nil
	}

	// If the next multiple is not within the range [a, b], return -1
	return int64(-1), nil
}
