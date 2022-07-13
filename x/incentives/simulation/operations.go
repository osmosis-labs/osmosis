package simulation

import (
	"math/rand"
	"time"

	osmo_simulation "github.com/osmosis-labs/osmosis/v7/simulation/types"

	"github.com/cosmos/cosmos-sdk/baseapp"

	"github.com/osmosis-labs/osmosis/v7/x/incentives/keeper"
	"github.com/osmosis-labs/osmosis/v7/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v7/x/lockup/types"

	"github.com/cosmos/cosmos-sdk/codec"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Simulation operation weights constants.
const (
	DefaultWeightMsgCreateGauge int = 10
	DefaultWeightMsgAddToGauge  int = 10
	OpWeightMsgCreateGauge          = "op_weight_msg_create_gauge"
	OpWeightMsgAddToGauge           = "op_weight_msg_add_to_gauge"
)

// WeightedOperations returns all the operations from the module with their respective weights.
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec, ak stakingTypes.AccountKeeper,
	bk stakingTypes.BankKeeper, ek types.EpochKeeper, k keeper.Keeper,
) simulation.WeightedOperations {
	var (
		weightMsgCreateGauge int
		weightMsgAddToGauge  int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgCreateGauge, &weightMsgCreateGauge, nil,
		func(_ *rand.Rand) {
			weightMsgCreateGauge = simappparams.DefaultWeightMsgCreateValidator
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgAddToGauge, &weightMsgAddToGauge, nil,
		func(_ *rand.Rand) {
			weightMsgAddToGauge = simappparams.DefaultWeightMsgCreateValidator
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgCreateGauge,
			SimulateMsgCreateGauge(ak, bk, ek, k),
		),
		simulation.NewWeightedOperation(
			weightMsgAddToGauge,
			SimulateMsgAddToGauge(ak, bk, k),
		),
	}
}

func genRewardCoins(r *rand.Rand, coins sdk.Coins) (res sdk.Coins) {
	numCoins := 1 + r.Intn(Min(coins.Len(), 1))
	denomIndices := r.Perm(numCoins)
	for i := 0; i < numCoins; i++ {
		denom := coins[denomIndices[i]].Denom
		amt, _ := simtypes.RandPositiveInt(r, coins[i].Amount)
		res = append(res, sdk.Coin{Denom: denom, Amount: amt})
	}

	return
}

func genQueryCondition(r *rand.Rand, blocktime time.Time, coins sdk.Coins, durations []time.Duration) lockuptypes.QueryCondition {
	// TODO: reset to 2 after postlaunch, only allow duration based query type on postlaunch
	// lockQueryType := r.Intn(2)
	lockQueryType := 0
	denom := coins[r.Intn(len(coins))].Denom
	// TODO: for postlaunch, only specific lock durations are allowed
	// durationSecs := r.Intn(1*60*60*24*7) + 1*60*60 // range of 1 week, min 1 hour
	// duration := time.Duration(durationSecs) * time.Second
	durationIndex := r.Intn(len(durations))
	duration := durations[durationIndex]
	timestampSecs := r.Intn(1 * 60 * 60 * 24 * 7) // range of 1 week
	timestamp := blocktime.Add(time.Duration(timestampSecs) * time.Second)

	return lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.LockQueryType(lockQueryType),
		Denom:         denom,
		Duration:      duration,
		Timestamp:     timestamp,
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

// SimulateMsgCreateGauge generates a MsgCreateGauge with random values.
func SimulateMsgCreateGauge(ak stakingTypes.AccountKeeper, bk stakingTypes.BankKeeper, ek types.EpochKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		simCoins := bk.SpendableCoins(ctx, simAccount.Address)
		if simCoins.Len() <= 0 {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgCreateGauge, "Account have no coin"), nil, nil
		}

		isPerpetual := r.Int()%2 == 0
		distributeTo := genQueryCondition(r, ctx.BlockTime(), simCoins, types.DefaultGenesis().LockableDurations)
		rewards := genRewardCoins(r, simCoins)
		startTimeSecs := r.Intn(1 * 60 * 60 * 24 * 7) // range of 1 week
		startTime := ctx.BlockTime().Add(time.Duration(startTimeSecs) * time.Second)
		durationSecs := r.Intn(1*60*60*24*7) + 1*60*60*24 // range of 1 week, min 1 day
		numEpochsPaidOver := uint64(r.Int63n(int64(durationSecs)/(ek.GetEpochInfo(ctx, k.GetParams(ctx).DistrEpochIdentifier).Duration.Milliseconds()/1000))) + 1

		if isPerpetual {
			numEpochsPaidOver = 1
		}

		msg := types.MsgCreateGauge{
			IsPerpetual:       isPerpetual,
			Owner:             simAccount.Address.String(),
			DistributeTo:      distributeTo,
			Coins:             rewards,
			StartTime:         startTime,
			NumEpochsPaidOver: numEpochsPaidOver,
		}

		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		return osmo_simulation.GenAndDeliverTxWithRandFees(
			r, app, txGen, &msg, rewards, ctx, simAccount, ak, bk, types.ModuleName)
	}
}

func SimulateMsgAddToGauge(ak stakingTypes.AccountKeeper, bk stakingTypes.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		simCoins := bk.SpendableCoins(ctx, simAccount.Address)
		if simCoins.Len() <= 0 {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgAddToGauge, "Account have no coin"), nil, nil
		}

		gauge := RandomGauge(ctx, r, k)
		if gauge == nil {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgAddToGauge, "No gauge exists"), nil, nil
		}
		gaugeId := RandomGauge(ctx, r, k).Id

		rewards := genRewardCoins(r, simCoins)

		msg := types.MsgAddToGauge{
			Owner:   simAccount.Address.String(),
			GaugeId: gaugeId,
			Rewards: rewards,
		}

		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		return osmo_simulation.GenAndDeliverTxWithRandFees(
			r, app, txGen, &msg, rewards, ctx, simAccount, ak, bk, types.ModuleName,
		)
	}
}

func RandomGauge(ctx sdk.Context, r *rand.Rand, k keeper.Keeper) *types.Gauge {
	gauges := k.GetGauges(ctx)
	if len(gauges) == 0 {
		return nil
	}
	return &gauges[r.Intn(len(gauges))]
}
