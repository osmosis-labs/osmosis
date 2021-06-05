package simulation

import (
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	osmo_simulation "github.com/osmosis-labs/osmosis/x/simulation"

	"github.com/cosmos/cosmos-sdk/codec"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/osmosis-labs/osmosis/x/incentives/keeper"
	"github.com/osmosis-labs/osmosis/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
)

// Simulation operation weights constants
const (
	DefaultWeightMsgCreatePot int = 10
	OpWeightMsgCreatePot          = "op_weight_msg_create_pool"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONMarshaler, ak stakingTypes.AccountKeeper,
	bk stakingTypes.BankKeeper, k keeper.Keeper,
) simulation.WeightedOperations {
	var (
		weightMsgCreatePot int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgCreatePot, &weightMsgCreatePot, nil,
		func(_ *rand.Rand) {
			weightMsgCreatePot = simappparams.DefaultWeightMsgCreateValidator
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgCreatePot,
			SimulateMsgCreatePot(ak, bk, k),
		),
	}
}

func genRewardCoins(r *rand.Rand, coins sdk.Coins) (res sdk.Coins) {
	numCoins := r.Intn(Min(coins.Len()-1, 2))+1
	denomIndices := r.Perm(numCoins)
	for i := 0; i < numCoins; i++ {
		denom := coins[denomIndices[i]].Denom
		amt, _ := simtypes.RandPositiveInt(r, coins[i].Amount)
		res = append(res, sdk.Coin{Denom: denom, Amount: amt})
	}

	return
}

func genQueryCondition(r *rand.Rand, coins sdk.Coins) lockuptypes.QueryCondition {
	lockQueryType := r.Intn(2)
	denom := coins[r.Intn(len(coins))].Denom
	yearSecs := r.Intn(1*60*60*24*365)
	duration := time.Date(0, 0, 0, 0, 0, yearSecs, 0, time.UTC).Sub(time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC))
	timestamp := time.Date(0, 0, 0, 0, 0, yearSecs, 0, time.UTC)

	return lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.LockQueryType(lockQueryType),
		Denom: denom,
		Duration: duration,
		Timestamp: timestamp,
	}
}

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// SimulateMsgCreatePot generates a MsgCreatePot with random values
func SimulateMsgCreatePot(ak stakingTypes.AccountKeeper, bk stakingTypes.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		simCoins := bk.SpendableCoins(ctx, simAccount.Address)
		if simCoins.Len() <= 1 {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgCreatePot, "Account doesn't have 2 different coin types"), nil, nil
		}

		isPerpetual := r.Int()%2 == 0
		distributeTo := genQueryCondition(r, simCoins)
		rewards := genRewardCoins(r, simCoins)
		yearSecs := r.Intn(1*60*60*24*365)
		startTime := time.Date(0, 0, 0, 0, 0, 0, yearSecs, time.UTC)
		numEpochsPaidOver := r.Int63n(int64(yearSecs/types.DefaultParams().DistrEpochIdentifier.Duration.Seconds()))

		msg := types.MsgCreatePot{
			IsPerpetual: isPerpetual,
			Owner: simAccount.Address.String(),
			DistributeTo: distributeTo,
			Coins: rewards,
			StartTime: startTime,
			NumEpochsPaidOver: numEpochsPaidOver,
		}

		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		return osmo_simulation.GenAndDeliverTxWithRandFees(
			r, app, txGen, &msg, rewards, ctx, simAccount, ak, bk, types.ModuleName)
	}
}
