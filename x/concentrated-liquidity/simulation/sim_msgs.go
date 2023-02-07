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
	gammtypes "github.com/osmosis-labs/osmosis/v14/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v14/x/poolmanager/types"
)

var PoolCreationFee = sdk.NewInt64Coin("stake", 10_000_000)

func RandomMsgCreateConcentratedPool(k clkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (*clmodeltypes.MsgCreateConcentratedPool, error) {
	rand := sim.GetSeededRand("select random seed")
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
	}, nil
}

func RandMsgWithdrawPosition(k clkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (*cltypes.MsgWithdrawPosition, error) {
	rand := sim.GetSeededRand("select random seed")
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

	// get percentage amount from 1 to 100 to withdraw liquidity
	randPerc := sdk.MustNewDecFromStr(fmt.Sprintf("%.2f", sim.RandomDecAmount(sdk.OneDec())))

	withdrawAmountInt := randPosition.Liquidity.Mul(randPerc).TruncateInt()

	return &cltypes.MsgWithdrawPosition{
		PoolId:          randPosition.PoolId,
		Sender:          sender.Address.String(),
		LowerTick:       randPosition.LowerTick,
		UpperTick:       randPosition.UpperTick,
		LiquidityAmount: withdrawAmountInt,
	}, nil
}

func RandMsgCollectFees(k clkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (*cltypes.MsgCollectFees, error) {
	rand := sim.GetSeededRand("select random seed")
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

// RandomSwapExactAmountIn utilizes a random pool and swaps and exact amount in for minimum of the secondary pool token
func RandomCLSwapExactAmountIn(k clkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (*gammtypes.MsgSwapExactAmountIn, error) {
	// get random pool, randomly select one of the pool denoms to be the coinIn, other is coinOut
	clPool, poolDenoms, err := getRandCLPool(k, sim, ctx)
	if err != nil {
		return nil, err
	}

	// set the swap route to use this pool
	route := []poolmanagertypes.SwapAmountInRoute{{
		PoolId:        clPool.GetId(),
		TokenOutDenom: poolDenoms[1],
	}}

	// find an address that has a balance of the coinIn
	sender, accCoinIn, senderExists := sim.SelAddrWithDenom(ctx, poolDenoms[0])
	if !senderExists {
		return nil, fmt.Errorf("no sender with denom %s exists", poolDenoms[0])
	}

	// select a random amount that is upper-bound by the address's balance of coinIn
	randomCoinSubset := sim.RandSubsetCoins(sdk.NewCoins(sdk.NewCoin(accCoinIn.Denom, accCoinIn.Amount)))

	poolI := clPool.(poolmanagertypes.PoolI)

	// calculate the minimum number of tokens received from input of tokenIn
	tokenOutMin, err := k.CalcOutAmtGivenIn(ctx, poolI, randomCoinSubset[0], poolDenoms[1], clPool.GetSwapFee(ctx))
	if err != nil {
		return nil, err
	}

	return &gammtypes.MsgSwapExactAmountIn{
		Sender:            sender.Address.String(),
		Routes:            route,
		TokenIn:           randomCoinSubset[0],
		TokenOutMinAmount: tokenOutMin.Amount,
	}, nil
}

// RandomSwapExactAmountOut utilizes a random pool and swaps a max amount amount in for an exact amount of the secondary pool token
func RandomCLSwapExactAmountOut(k clkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (*gammtypes.MsgSwapExactAmountOut, error) {
	// get random pool, randomly select one of the pool denoms to be the coinIn, other is coinOut
	clPool, poolDenoms, err := getRandCLPool(k, sim, ctx)
	if err != nil {
		return nil, err
	}

	// set the swap route to use this pool
	route := []poolmanagertypes.SwapAmountOutRoute{{
		PoolId:       clPool.GetId(),
		TokenInDenom: poolDenoms[0],
	}}

	// find an address that has a balance of the coinIn
	senderAcc, accCoin, senderExists := sim.SelAddrWithDenom(ctx, poolDenoms[0])
	if !senderExists {
		return nil, fmt.Errorf("no sender with denom %s exists", poolDenoms[0])
	}

	// set the subset of coins to be upper-bound to the minimum between the address and the pool itself
	randomCoinSubset := sim.RandSubsetCoins(sdk.NewCoins(sdk.NewCoin(accCoin.Denom, accCoin.Amount)))
	// gamm Example: randomCoinInSubset := osmoutils.MinCoins(sdk.NewCoins(coinIn), sdk.NewCoins(accCoin))

	poolI := clPool.(poolmanagertypes.PoolI)

	// utilize CalcOutAmtGivenIn to calculate tokenOut and use tokenOut to calculate tokenInMax
	tokenOut, err := k.CalcOutAmtGivenIn(ctx, poolI, randomCoinSubset[0], poolDenoms[1], clPool.GetSwapFee(ctx))
	if err != nil {
		return nil, err
	}

	// maximum amount we need to complete the swap
	tokenInMax, err := k.CalcInAmtGivenOut(ctx, poolI, tokenOut, poolDenoms[0], clPool.GetSwapFee(ctx))
	if err != nil {
		return nil, err
	}

	// Example:
	// randomCoinSubset: 10osmo
	// CalcOutAmtGivenIn: 1atom  [tokenIn: 10osmo, tokenout:atom]
	// CalcInAmtGivenOut: 10osmo [tokenOut: 1atom, tokenout: osmo]

	return &gammtypes.MsgSwapExactAmountOut{
		Sender:           senderAcc.Address.String(),
		Routes:           route,
		TokenInMaxAmount: tokenInMax.Amount,
		TokenOut:         tokenOut,
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

//RandomTickDivisibility calculates a random number between minTick - maxTick (inclusive) that is divisible by tickSpacing
func RandomTickDivisibility(sim *osmosimtypes.SimCtx, minTick int64, maxTick int64, tickSpacing uint64) (int64, error) {
	rand := sim.GetSeededRand("select random seed")

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
