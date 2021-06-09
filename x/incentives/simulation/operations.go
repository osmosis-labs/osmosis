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
	DefaultWeightMsgCreateGauge int = 10
	OpWeightMsgCreateGauge          = "op_weight_msg_create_pool"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONMarshaler, ak stakingTypes.AccountKeeper,
	bk stakingTypes.BankKeeper, ek types.EpochKeeper, k keeper.Keeper,
) simulation.WeightedOperations {
	var (
		weightMsgCreateGauge int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgCreateGauge, &weightMsgCreateGauge, nil,
		func(_ *rand.Rand) {
			weightMsgCreateGauge = simappparams.DefaultWeightMsgCreateValidator
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgCreateGauge,
			SimulateMsgCreateGauge(ak, bk, ek, k),
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
	durationSecs := r.Intn(1*60*60)
	duration := time.Duration(durationSecs) * time.Second
	timestamp := time.Date(0, 0, 0, 0, 0, durationSecs, 0, time.UTC)

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

// SimulateMsgCreateGauge generates a MsgCreateGauge with random values
func SimulateMsgCreateGauge(ak stakingTypes.AccountKeeper, bk stakingTypes.BankKeeper, ek types.EpochKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		simCoins := bk.SpendableCoins(ctx, simAccount.Address)
		if simCoins.Len() <= 1 {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgCreateGauge, "Account doesn't have 2 different coin types"), nil, nil
		}

		isPerpetual := r.Int()%2 == 0
		distributeTo := genQueryCondition(r, simCoins)
		rewards := genRewardCoins(r, simCoins)
		yearSecs := r.Intn(1*60*60*24*365)
		startTime := time.Date(0, 0, 0, 0, 0, 0, yearSecs, time.UTC)
		numEpochsPaidOver := uint64(r.Int63n(int64(yearSecs)/(ek.GetEpochInfo(ctx, k.GetParams(ctx).DistrEpochIdentifier).Duration.Milliseconds()/1000)))

		msg := types.MsgCreateGauge{
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
