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
<<<<<<< HEAD
	rand := sim.GetRand()
	authorizedTickSpacing := cltypes.AuthorizedTickSpacing

	// find an address with two or more distinct denoms in their wallet
	sender, senderExists := sim.RandomSimAccountWithConstraint(createPoolRestriction(k, sim, ctx))
	if !senderExists {
		return nil, fmt.Errorf("no sender with two different denoms & pool creation fee exists")
=======
	poolCreator, coin0, coin1, tickSpacing, exponentAtPriceOne, swapFee, err := RandomPreparePoolFunc(sim, ctx, k)
	if err != nil {
		return nil, err
>>>>>>> a842aaf57 ([CL Incentives] Implement incentives wiring from governance-triggered emissions)
	}

	return &clmodeltypes.MsgCreateConcentratedPool{
<<<<<<< HEAD
		Sender:      sender.Address.String(),
		Denom0:      denom0,
		Denom1:      denom1,
		TickSpacing: tickSpacing,
		SwapFee:     sdk.NewDecWithPrec(1, 2),
=======
		Sender:             poolCreator.String(),
		Denom0:             coin0.Denom,
		Denom1:             coin1.Denom,
		TickSpacing:        tickSpacing,
		ExponentAtPriceOne: exponentAtPriceOne,
		SwapFee:            swapFee,
>>>>>>> a842aaf57 ([CL Incentives] Implement incentives wiring from governance-triggered emissions)
	}, nil
}

func RandMsgCreatePosition(k clkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (*cltypes.MsgCreatePosition, error) {
	// get random pool
	clPool, poolDenoms, err := getRandCLPool(k, sim, ctx)
	if err != nil {
		return nil, err
	}

<<<<<<< HEAD
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

	// Randomize lowerTick and upperTick from max values to create position
	lowerTick, upperTick, err := getRandomTickPositions(sim, cltypes.MinTick, cltypes.MaxTick, clPool.GetTickSpacing())
=======
	positionCreator, tokens, lowerTick, upperTick, err := RandomPrepareCreatePositionFunc(sim, ctx, clPool, poolDenoms)
>>>>>>> a842aaf57 ([CL Incentives] Implement incentives wiring from governance-triggered emissions)
	if err != nil {
		return nil, err
	}

	return &cltypes.MsgCreatePosition{
		PoolId:          clPool.GetId(),
		Sender:          positionCreator.String(),
		LowerTick:       lowerTick,
		UpperTick:       upperTick,
		TokenDesired0:   tokens[0],
		TokenDesired1:   tokens[1],
		TokenMinAmount0: sdk.NewInt(0),
		TokenMinAmount1: sdk.NewInt(0),
	}, nil
}

func RandMsgWithdrawPosition(k clkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (*cltypes.MsgWithdrawPosition, error) {
	rand := sim.GetRand()
	// get random pool
	clPool, poolDenoms, err := getRandCLPool(k, sim, ctx)
	if err != nil {
		return nil, err
	}

	// positionCreator creates the position with pool denoms
	positionCreator, tokens, lowerTick, upperTick, err := RandomPrepareCreatePositionFunc(sim, ctx, clPool, poolDenoms)
	if err != nil {
		return nil, err
	}

	var positionIds []uint64
	remainingToken0Amt := tokens[0].Amount
	remainingToken1Amt := tokens[1].Amount
	// create positions until the funds run out from positionCreator
	for remainingToken0Amt.GT(sdk.ZeroInt()) && remainingToken1Amt.GT(sdk.ZeroInt()) {
		//make sure the user always has tokens to create Position
		randToken0Amt := sim.RandomAmount(remainingToken0Amt)
		randToken1Amt := sim.RandomAmount(remainingToken1Amt)

		positionId, _, _, _, _, err := k.CreatePosition(ctx, clPool.GetId(), positionCreator, randToken0Amt, randToken1Amt, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick, upperTick)
		if err != nil {
			return nil, err
		}

		remainingToken0Amt = remainingToken0Amt.Sub(randToken0Amt)
		remainingToken1Amt = remainingToken1Amt.Sub(randToken1Amt)

		positionIds = append(positionIds, positionId)
	}

	randPositionId := positionIds[rand.Intn(len(positionIds))]
	position, err := k.GetPosition(ctx, randPositionId)
	if err != nil {
		return nil, err
	}

	withdrawAmount := sim.RandomDecAmount(position.Liquidity)

	// check if there is enough position liquidity to withdraw
	availableLiquidity, err := k.GetPositionLiquidity(ctx, randPositionId)
	if err != nil {
		return nil, err
	}

	if withdrawAmount.GT(availableLiquidity) {
		return nil, fmt.Errorf("insufficient liquidity requested to withdraw.")
	}

	return &cltypes.MsgWithdrawPosition{
		PositionId:      randPositionId,
		Sender:          positionCreator.String(),
		LiquidityAmount: withdrawAmount,
	}, nil
}

func RandMsgCollectFeesFullFlow(k clkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (*cltypes.MsgCollectFees, error) {
	// poolCreator creates Pool, randomize pool creation values
	poolCreator, coin0, coin1, tickSpacing, exponentAtPriceOne, swapFee, err := RandomPreparePoolFunc(sim, ctx, k)
	if err != nil {
		return nil, err
	}

	// create Pool
	poolId, err := sim.PoolManagerKeeper().CreatePool(ctx, clmodeltypes.NewMsgCreateConcentratedPool(poolCreator, coin0.Denom, coin1.Denom, tickSpacing, exponentAtPriceOne, swapFee))
	if err != nil {
		return nil, err
	}

	poolI, err := k.GetPool(ctx, poolId)
	if err != nil {
		return nil, err
	}

	clPool, ok := poolI.(cltypes.ConcentratedPoolExtension)
	if !ok {
		return nil, fmt.Errorf("pool is not concentrated liquidity pool")
	}

	poolDenoms := []string{clPool.GetToken0(), clPool.GetToken1()}

	// positionCreator creates the position with pool denoms
	positionCreator, tokens, lowerTick, upperTick, err := RandomPrepareCreatePositionFunc(sim, ctx, clPool, poolDenoms)
	if err != nil {
		return nil, err
	}

	var positionIds []uint64
	remainingToken0Amt := tokens[0].Amount
	remainingToken1Amt := tokens[1].Amount
	// create positions until the funds run out from positionCreator
	for remainingToken0Amt.GT(sdk.ZeroInt()) && remainingToken1Amt.GT(sdk.ZeroInt()) {
		//make sure the user always has tokens to create Position
		randToken0Amt := sim.RandomAmount(remainingToken0Amt)
		randToken1Amt := sim.RandomAmount(remainingToken1Amt)

		positionId, _, _, _, _, err := k.CreatePosition(ctx, poolI.GetId(), positionCreator, randToken0Amt, randToken1Amt, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick, upperTick)
		if err != nil {
			return nil, err
		}

		remainingToken0Amt = remainingToken0Amt.Sub(randToken0Amt)
		remainingToken1Amt = remainingToken1Amt.Sub(randToken1Amt)

		positionIds = append(positionIds, positionId)
	}

	// SwapOwner performs the swap from token0 to token1
	swapOwner, swapOwnerTokens, senderExists := sim.SelAddrWithDenoms(ctx, poolDenoms)
	if !senderExists {
		return nil, fmt.Errorf("no sender with denoms %s exists", poolDenoms)
	}

	// ensure that we always have 2 tokens
	// Note: tokens returns a random subset of poolDenoms, so  had to add this assertion
	if len(swapOwnerTokens) < 2 {
		return nil, fmt.Errorf("user doesnot have pool tokens")
	}

	// perform swap until token 1 runs out
	remainingSwapOwnerToken0Amt := swapOwnerTokens[0].Amount
	remainingSwapOwnerToken1Amt := swapOwnerTokens[1].Amount
	for remainingSwapOwnerToken0Amt.GT(sdk.ZeroInt()) && remainingSwapOwnerToken1Amt.GT(sdk.ZeroInt()) {
		randToken0Amt := sim.RandomAmount(remainingSwapOwnerToken0Amt)
		randToken1Amt := sim.RandomAmount(remainingSwapOwnerToken1Amt)

		// perform swap from token0 to token1 until either token0 or token1 fund runs out
		_, err = k.SwapExactAmountIn(ctx, swapOwner.Address, poolI, sdk.NewCoin(swapOwnerTokens[0].Denom, randToken0Amt), swapOwnerTokens[1].Denom, sdk.OneInt(), swapFee)
		if err != nil {
			return nil, err
		}

		// perform swap from token1 to token0 until either token0 or token1 fund runs out
		_, err = k.SwapExactAmountIn(ctx, swapOwner.Address, poolI, sdk.NewCoin(swapOwnerTokens[1].Denom, randToken1Amt), swapOwnerTokens[0].Denom, sdk.OneInt(), swapFee)
		if err != nil {
			return nil, err
		}

		remainingSwapOwnerToken0Amt = remainingSwapOwnerToken0Amt.Sub(randToken0Amt)
		remainingSwapOwnerToken1Amt = remainingSwapOwnerToken1Amt.Sub(randToken1Amt)
	}

	return &cltypes.MsgCollectFees{
		Sender:      positionCreator.String(),
		PositionIds: positionIds,
	}, nil
}

func RandMsgCollectIncentivesFullFlow(k clkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (*cltypes.MsgCollectIncentives, error) {
	// poolCreator creates Pool, randomize pool creation values
	poolCreator, coin0, coin1, tickSpacing, exponentAtPriceOne, swapFee, err := RandomPreparePoolFunc(sim, ctx, k)
	if err != nil {
		return nil, err
	}

	// CreatePool
	poolId, err := sim.PoolManagerKeeper().CreatePool(ctx, clmodeltypes.NewMsgCreateConcentratedPool(poolCreator, coin0.Denom, coin1.Denom, tickSpacing, exponentAtPriceOne, swapFee))
	if err != nil {
		return nil, err
	}

	poolI, err := k.GetPool(ctx, poolId)
	if err != nil {
		return nil, err
	}

	clPool, ok := poolI.(cltypes.ConcentratedPoolExtension)
	if !ok {
		return nil, fmt.Errorf("pool is not concentrated liquidity pool")
	}

	poolDenoms := []string{clPool.GetToken0(), clPool.GetToken1()}

	// poolCreator creates incentives with same denom values
	_, err = k.CreateIncentive(ctx, poolI.GetId(), poolCreator, coin0.Denom, coin0.Amount, sdk.MustNewDecFromStr("0.008"), ctx.BlockTime(), time.Hour*24*7)
	if err != nil {
		return nil, err
	}

	// positionCreator creates position
	positionCreator, tokens, lowerTick, upperTick, err := RandomPrepareCreatePositionFunc(sim, ctx, clPool, poolDenoms)
	if err != nil {
		return nil, err
	}

	var positionIds []uint64
	remainingToken0Amt := tokens[0].Amount
	remainingToken1Amt := tokens[1].Amount
	// create positions until the funds run out from positionCreator
	for remainingToken0Amt.GT(sdk.ZeroInt()) && remainingToken1Amt.GT(sdk.ZeroInt()) {
		//make sure the user always has tokens to create Position
		randToken0Amt := sim.RandomAmount(remainingToken0Amt)
		randToken1Amt := sim.RandomAmount(remainingToken1Amt)

		positionId, _, _, _, _, err := k.CreatePosition(ctx, poolI.GetId(), positionCreator, randToken0Amt, randToken1Amt, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick, upperTick)
		if err != nil {
			return nil, err
		}

		remainingToken0Amt = remainingToken0Amt.Sub(randToken0Amt)
		remainingToken1Amt = remainingToken1Amt.Sub(randToken1Amt)

		positionIds = append(positionIds, positionId)
	}

	// PositionCreator collects the incentives
	return &cltypes.MsgCollectIncentives{
		PositionIds: positionIds,
		Sender:      positionCreator.String(),
	}, nil
}

func RandMsgCreateIncentives(k clkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (*cltypes.MsgCreateIncentive, error) {
	rand := sim.GetRand()
	// get random pool
	clPool, poolDenoms, err := getRandCLPool(k, sim, ctx)
	if err != nil {
		return nil, err
	}

	// incentiveCreator creates the incentives by supplying tokens
	incentiveCreator, incentivesTokens, senderExists := sim.SelAddrWithDenoms(ctx, poolDenoms)
	if !senderExists {
		return nil, fmt.Errorf("no sender with denoms %s exists", poolDenoms)
	}

	// emission rate is nonzero and nonnegative from 1 to 1million
	randEmissionVal := sim.RandomDecAmount(sdk.MustNewDecFromStr("1000000"))

	startTimeSecs := rand.Intn(1 * 60 * 60 * 24 * 7) // range of 1 week
	startTime := ctx.BlockTime().Add(time.Duration(startTimeSecs) * time.Second)

	durations := []time.Duration{time.Nanosecond, time.Minute, time.Hour, time.Hour * 24, time.Hour * 24 * 7}
	randomDurationIndex := rand.Intn(len(durations))

	// Get the duration value at the random index
	randomDuration := durations[randomDurationIndex]

	return &cltypes.MsgCreateIncentive{
		PoolId:          clPool.GetId(),
		Sender:          incentiveCreator.Address.String(),
		IncentiveDenom:  incentivesTokens[0].Denom,
		IncentiveAmount: incentivesTokens[0].Amount,
		EmissionRate:    randEmissionVal,
		StartTime:       startTime,
		MinUptime:       randomDuration,
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
	clPools, err := k.GetPools(ctx)
	if err != nil {
		return nil, nil, err
	}

	numPools := len(clPools)
	if numPools == 0 {
		return nil, nil, fmt.Errorf("no pools created")
	}

	randPool := clPools[rand.Intn(numPools)]
	randClPool, ok := randPool.(cltypes.ConcentratedPoolExtension)
	if !ok {
		return nil, nil, fmt.Errorf("pool is not concentrated liquidity pool")
	}

	poolDenoms := []string{randClPool.GetToken0(), randClPool.GetToken1()}

	return randClPool, poolDenoms, err
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

func RandomPreparePoolFunc(sim *osmosimtypes.SimCtx, ctx sdk.Context, k clkeeper.Keeper) (sdk.AccAddress, sdk.Coin, sdk.Coin, uint64, sdk.Int, sdk.Dec, error) {
	rand := sim.GetRand()
	minExponentAtOneValue := cltypes.ExponentAtPriceOneMin.Int64()
	maxExponentAtOneValue := cltypes.ExponentAtPriceOneMax.Int64()

	// generate random values from -13 to 1 (current accepted range: -12 to -1)
	exponentAtPriceOne := sdk.NewInt((minExponentAtOneValue + 2) + rand.Int63n((maxExponentAtOneValue-1)-(minExponentAtOneValue+2)+1))
	authorizedTickSpacing := cltypes.AuthorizedTickSpacing

	// find an address with two or more distinct denoms in their wallet
	sender, senderExists := sim.RandomSimAccountWithConstraint(createPoolRestriction(k, sim, ctx))
	if !senderExists {
		return nil, sdk.Coin{}, sdk.Coin{}, 0, sdk.Int{}, sdk.Dec{}, fmt.Errorf("no sender with two different denoms & pool creation fee exists")
	}

	// get random 3 coins, use 2 to create pool and 1 for fees (stake denom).
	poolCoins, ok := sim.GetRandSubsetOfKDenoms(ctx, sender, 3)
	if !ok {
		return nil, sdk.Coin{}, sdk.Coin{}, 0, sdk.Int{}, sdk.Dec{}, fmt.Errorf("provided sender with requested number of denoms does not exist")
	}

	// check if the sender has sufficient amount for fees
	if poolCoins.Add(PoolCreationFee).IsAnyGT(sim.BankKeeper().SpendableCoins(ctx, sender.Address)) {
		return nil, sdk.Coin{}, sdk.Coin{}, 0, sdk.Int{}, sdk.Dec{}, errors.New("chose an account / creation amount that didn't pass fee limit")
	}

	if poolCoins[0].Denom == sdk.DefaultBondDenom || poolCoins[1].Denom == sdk.DefaultBondDenom {
		return nil, sdk.Coin{}, sdk.Coin{}, 0, sdk.Int{}, sdk.Dec{}, fmt.Errorf("poolCoins contains denom stake which contains invalid metadata")
	}

	coin0 := poolCoins[0]
	coin1 := poolCoins[1]
	tickSpacing := authorizedTickSpacing[rand.Intn(len(authorizedTickSpacing))]
	swapFee := sdk.NewDecWithPrec(1, 2)

	return sender.Address, coin0, coin1, tickSpacing, exponentAtPriceOne, swapFee, nil
}

func RandomPrepareCreatePositionFunc(sim *osmosimtypes.SimCtx, ctx sdk.Context, clPool cltypes.ConcentratedPoolExtension, poolDenoms []string) (sdk.AccAddress, sdk.Coins, int64, int64, error) {
	// User C creates Position
	// make sure that the position creator has the poolTokens
	positionCreator, tokens, senderExists := sim.SelAddrWithDenoms(ctx, poolDenoms)
	if !senderExists {
		return nil, sdk.Coins{}, 0, 0, fmt.Errorf("no sender with denoms %s exists", poolDenoms)
	}

	// ensure that we always have 2 tokens
	// Note: tokens returns a random subset of poolDenoms, so  had to add this assertion
	if len(tokens) < 2 {
		return nil, sdk.Coins{}, 0, 0, fmt.Errorf("user doesnot have pool tokens")
	}

	//  Retrieve minTick and maxTick from precision factor
	minTick, maxTick := clkeeper.GetMinAndMaxTicksFromExponentAtPriceOne(clPool.GetExponentAtPriceOne())

	// Randomize lowerTick and upperTick from max values to create position
	lowerTick, upperTick, err := getRandomTickPositions(sim, minTick, maxTick, clPool.GetTickSpacing())
	if err != nil {
		return nil, sdk.Coins{}, 0, 0, err
	}

	return positionCreator.Address, tokens, lowerTick, upperTick, nil
}
