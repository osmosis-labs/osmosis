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
	"github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/types"
	cltypes "github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/types"
	gammtypes "github.com/osmosis-labs/osmosis/v14/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v14/x/poolmanager/types"
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

	// generate 3 coins, use 2 to create pool and 1 for fees.
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
	// get random pool
	clPool, _, poolDenoms, err := getRandCLPool(k, sim, ctx)
	if err != nil {
		return nil, err
	}

	// get random user address with the pool denoms
	sender, tokens, senderExists := sim.SelAddrWithDenoms(ctx, poolDenoms)
	if !senderExists {
		return nil, fmt.Errorf("no sender with denoms %s exists", poolDenoms)
	}

	// ensure that we always have 2 tokens
	if len(tokens) < 2 {
		return nil, fmt.Errorf("no pool denoms tokens")
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
		// TODO: Randomzize TokenMinAmount0 and TokenMinAmount1 in next iteration
		TokenMinAmount0: sdk.NewInt(0),
		TokenMinAmount1: sdk.NewInt(0),
	}, nil
}

func RandMsgWithdrawPosition(k clkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (*cltypes.MsgWithdrawPosition, error) {
	// get random pool
	clPool, _, poolDenoms, err := getRandCLPool(k, sim, ctx)
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
		return nil, fmt.Errorf("position doesnot exist")
	}

	if len(positions) == 0 {
		return nil, fmt.Errorf("user doesnot have any position")
	}

	// pick a random position
	randPosition := positions[rand.Intn(len(positions))]

	// Randomize tick values from minTick to maxTick
	minTick, maxTick := clkeeper.GetMinAndMaxTicksFromExponentAtPriceOne(clPool.GetPrecisionFactorAtPriceOne())

	if randPosition.LowerTick < minTick || randPosition.UpperTick > maxTick {
		return nil, fmt.Errorf("invalid ticks")
	}

	withdrawAmount := randPosition.Liquidity.TruncateInt()

	fmt.Println(randPosition)
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
	clPool, _, poolDenoms, err := getRandCLPool(k, sim, ctx)
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
		return nil, fmt.Errorf("position doesnot exist")
	}

	if len(positions) == 0 {
		return nil, fmt.Errorf("user doesnot have any position")
	}

	// pick a random position
	randPosition := positions[rand.Intn(len(positions))]

	// Randomize tick values from minTick to maxTick
	minTick, maxTick := clkeeper.GetMinAndMaxTicksFromExponentAtPriceOne(clPool.GetPrecisionFactorAtPriceOne())

	if randPosition.LowerTick < minTick || randPosition.UpperTick > maxTick {
		return nil, fmt.Errorf("invalid ticks")
	}

	return &cltypes.MsgCollectFees{
		PoolId:    randPosition.PoolId,
		Sender:    sender.Address.String(),
		LowerTick: randPosition.LowerTick,
		UpperTick: randPosition.UpperTick,
	}, nil
}

// RandomSwapExactAmountIn utilizes a random  cl pool and swaps and exact amount in for minimum of the secondary pool token
func RandomSwapExactAmountIn(k clkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (*gammtypes.MsgSwapExactAmountIn, error) {
	// get random pool, randomly select one of the pool denoms to be the coinIn, other is coinOut
	clPool, poolI, poolDenoms, err := getRandCLPool(k, sim, ctx)
	if err != nil {
		return nil, err
	}

	// set the swap route to use this pool
	route := []poolmanagertypes.SwapAmountInRoute{{
		PoolId:        clPool.GetId(),
		TokenOutDenom: poolDenoms[0],
	}}

	// find an address that has a balance of the coinIn
	sender, accCoinIn, senderExists := sim.SelAddrWithDenom(ctx, poolDenoms[1])
	if !senderExists {
		return nil, fmt.Errorf("no sender with denom %s exists", poolDenoms[1])
	}

	// select a random amount that is upper-bound by the address's balance of coinIn
	randomCoinSubset := sim.RandSubsetCoins(sdk.NewCoins(sdk.NewCoin(accCoinIn.Denom, accCoinIn.Amount)))

	// calculate the minimum number of tokens received from input of tokenIn
	tokenOutMin, err := k.CalcOutAmtGivenIn(ctx, poolI, randomCoinSubset[0], poolDenoms[0], clPool.GetSwapFee(ctx))
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

//RandomSwapExactAmountOut utilizes a random cl pool and swaps a max amount amount in for an exact amount of the secondary pool token
func RandomSwapExactAmountOut(k clkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (*gammtypes.MsgSwapExactAmountOut, error) {
	// get random pool, randomly select one of the pool denoms to be the coinIn, other is coinOut
	clPool, poolI, poolDenoms, err := getRandCLPool(k, sim, ctx)
	if err != nil {
		return nil, err
	}

	// set the swap route to use this pool
	route := []poolmanagertypes.SwapAmountOutRoute{{
		PoolId:       clPool.GetId(),
		TokenInDenom: poolDenoms[1],
	}}

	// find an address that has a balance of the coinIn
	senderAcc, accCoinIn, senderExists := sim.SelAddrWithDenom(ctx, poolDenoms[1])
	if !senderExists {
		return nil, fmt.Errorf("no sender with denom %s exists", poolDenoms[1])
	}

	// set the subset of coins to be upper-bound to the minimum between the address and the pool itself
	randomCoinInAmt := sim.RandSubsetCoins(sdk.NewCoins(accCoinIn))

	tokenOut, err := k.CalcOutAmtGivenIn(ctx, poolI, randomCoinInAmt[0], poolDenoms[0], clPool.GetSwapFee(ctx))
	if err != nil {
		return nil, err
	}

	tokenInMax, err := k.CalcInAmtGivenOut(ctx, poolI, tokenOut, poolDenoms[1], clPool.GetSwapFee(ctx))
	if err != nil {
		return nil, err
	}

	return &gammtypes.MsgSwapExactAmountOut{
		Sender:           senderAcc.Address.String(),
		Routes:           route,
		TokenInMaxAmount: tokenInMax.Amount,
		TokenOut:         tokenOut,
	}, nil
}

func createPoolRestriction(k clkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) osmosimtypes.SimAccountConstraint {
	return func(acc legacysimulationtype.Account) bool {
		accCoins := sim.BankKeeper().SpendableCoins(ctx, acc.Address)
		hasTwoCoins := len(accCoins) >= 2
		hasPoolCreationFee := accCoins.AmountOf(PoolCreationFee.Denom).GT(PoolCreationFee.Amount)
		return hasTwoCoins && hasPoolCreationFee
	}
}

// getRandCLPool gets a concnerated liquidity pool with its pool denoms.
func getRandCLPool(k clkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (types.ConcentratedPoolExtension, poolmanagertypes.PoolI, []string, error) {
	// get all pools
	clPools, err := k.GetAllPools(ctx)
	if err != nil {
		return nil, nil, nil, err
	}

	numPools := len(clPools)
	if numPools == 0 {
		return nil, nil, nil, fmt.Errorf("no pools created")
	}

	pool_id := clPools[rand.Intn(numPools)].GetId()

	// check if the pool exists
	poolI, err := k.GetPool(ctx, pool_id)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Pool not found for id %d", pool_id)
	}

	concentratedPool, ok := poolI.(cltypes.ConcentratedPoolExtension)
	if !ok {
		return nil, nil, nil, fmt.Errorf("interface conversion failed")
	}

	poolDenoms := []string{concentratedPool.GetToken0(), concentratedPool.GetToken1()}

	return concentratedPool, poolI, poolDenoms, err
}

// getRandomTickPositions returns random lowerTick and upperTick divisible by tickSpacing value.
func getRandomTickPositions(minTick, maxTick int64, tickSpacing uint64) (int64, int64, error) {
	lower_tick_disivisble := ((maxTick - minTick) / int64(tickSpacing)) + 1 // get random value between minTick, maxTick that is divisible by TickSpacing
	if lower_tick_disivisble < 1 {
		return 0, 0, fmt.Errorf("lower tick divisible by tickspacing not found")
	}
	lowerTick := rand.Int63n(lower_tick_disivisble)*int64(tickSpacing) + minTick //  get random value between minTick, lowerTick that is divisible by TickSpacing

	if lowerTick%int64(tickSpacing) != 0 {
		return 0, 0, fmt.Errorf("lower tick is not divisible by tickspacing %d", tickSpacing)
	}

	upper_tick_disivisble := ((maxTick - lowerTick) / int64(tickSpacing)) + 1
	if upper_tick_disivisble < 1 {
		return 0, 0, fmt.Errorf("upper tick divisible by tickspacing not found")
	}
	upperTick := rand.Int63n(upper_tick_disivisble)*int64(tickSpacing) + lowerTick

	if upperTick%int64(tickSpacing) != 0 {
		return 0, 0, fmt.Errorf("lower tick is not divisible by tickspacing %d", tickSpacing)
	}

	return lowerTick, upperTick, nil
}
