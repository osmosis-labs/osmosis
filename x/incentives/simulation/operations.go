package simulation

import (
	"math/rand"
	"time"

	"github.com/osmosis-labs/osmosis/osmomath"
	osmosimtypes "github.com/osmosis-labs/osmosis/v27/simulation/simtypes"

	"github.com/cosmos/cosmos-sdk/baseapp"

	"github.com/osmosis-labs/osmosis/v27/x/incentives/keeper"
	"github.com/osmosis-labs/osmosis/v27/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	stakingsim "github.com/cosmos/cosmos-sdk/x/staking/simulation"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Simulation operation weights constants.
const (
	DefaultWeightMsgCreateGauge int = 10
	DefaultWeightMsgAddToGauge  int = 10
	OpWeightMsgCreateGauge          = "op_weight_msg_create_gauge"
	OpWeightMsgAddToGauge           = "op_weight_msg_add_to_gauge"
)

var (
	typeMsgCreateGauge = sdk.MsgTypeURL(&types.MsgCreateGauge{})
	typeMsgAddToGauge  = sdk.MsgTypeURL(&types.MsgAddToGauge{})
)

// WeightedOperations returns all the operations from the module with their respective weights.
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec, ak stakingTypes.AccountKeeper,
	bk osmosimtypes.BankKeeper, ek types.EpochKeeper, k keeper.Keeper,
) simulation.WeightedOperations {
	var (
		weightMsgCreateGauge int
		weightMsgAddToGauge  int
	)

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	protoCdc := codec.NewProtoCodec(interfaceRegistry)

	appParams.GetOrGenerate(OpWeightMsgCreateGauge, &weightMsgCreateGauge, nil,
		func(_ *rand.Rand) {
			weightMsgCreateGauge = stakingsim.DefaultWeightMsgCreateValidator
		},
	)

	appParams.GetOrGenerate(OpWeightMsgAddToGauge, &weightMsgAddToGauge, nil,
		func(_ *rand.Rand) {
			weightMsgAddToGauge = stakingsim.DefaultWeightMsgCreateValidator
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgCreateGauge,
			SimulateMsgCreateGauge(protoCdc, ak, bk, ek, k),
		),
		simulation.NewWeightedOperation(
			weightMsgAddToGauge,
			SimulateMsgAddToGauge(protoCdc, ak, bk, k),
		),
	}
}

// genRewardCoins generates a random number of coin denoms with a respective random value for each coin.
func genRewardCoins(r *rand.Rand, coins sdk.Coins, fee osmomath.Int) (res sdk.Coins) {
	numCoins := 1 + r.Intn(Min(coins.Len(), 1))
	denomIndices := r.Perm(numCoins)
	for i := 0; i < numCoins; i++ {
		var (
			amt osmomath.Int
			err error
		)
		denom := coins[denomIndices[i]].Denom
		if denom == sdk.DefaultBondDenom {
			amt, err = simtypes.RandPositiveInt(r, coins[i].Amount.Sub(fee))
			if err != nil {
				panic(err)
			}
		} else {
			amt, err = simtypes.RandPositiveInt(r, coins[i].Amount)
			if err != nil {
				panic(err)
			}
		}
		res = append(res, sdk.Coin{Denom: denom, Amount: amt})
	}
	return
}

// genQueryCondition returns a single lockup QueryCondition, which is generated from a single coin randomly selected from the provided coin array
func genQueryCondition(r *rand.Rand, blocktime time.Time, coins sdk.Coins, durations []time.Duration) lockuptypes.QueryCondition {
	lockQueryType := 0
	denom := coins[r.Intn(len(coins))].Denom
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

// SimulateMsgCreateGauge generates and executes a MsgCreateGauge with random parameters
func SimulateMsgCreateGauge(cdc *codec.ProtoCodec, ak stakingTypes.AccountKeeper, bk osmosimtypes.BankKeeper, ek types.EpochKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		simCoins := bk.SpendableCoins(ctx, simAccount.Address)
		if simCoins.AmountOf(sdk.DefaultBondDenom).LT(types.CreateGaugeFee) {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgCreateGauge, "Account have no coin"), nil, nil
		}

		isPerpetual := r.Int()%2 == 0
		distributeTo := genQueryCondition(r, ctx.BlockTime(), simCoins, types.DefaultGenesis().LockableDurations)
		rewards := genRewardCoins(r, simCoins, types.CreateGaugeFee)
		startTimeSecs := r.Intn(1 * 60 * 60 * 24 * 7) // range of 1 week
		startTime := ctx.BlockTime().Add(time.Duration(startTimeSecs) * time.Second)
		durationSecs := r.Intn(1*60*60*24*7) + 1*60*60*24 // range of 1 week, min 1 day
		numEpochsPaidOver := uint64(r.Int63n(int64(durationSecs)/(ek.GetEpochInfo(ctx, k.GetParams(ctx).DistrEpochIdentifier).Duration.Milliseconds()/1000))) + 1

		if isPerpetual {
			numEpochsPaidOver = 1
		}

		msg := types.MsgCreateGauge{
			Owner:             simAccount.Address.String(),
			IsPerpetual:       isPerpetual,
			DistributeTo:      distributeTo,
			Coins:             rewards,
			StartTime:         startTime,
			NumEpochsPaidOver: numEpochsPaidOver,
		}

		opMsg, err := osmosimtypes.GenerateAndDeliverTx(r, app, ctx, chainID, cdc, ak, bk, simAccount, types.ModuleName, &msg, typeMsgCreateGauge, false)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreateGauge, "unable to generate and deliver tx"), nil, err
		}

		return opMsg, nil, nil
	}
}

// SimulateMsgAddToGauge generates and executes a MsgAddToGauge with random parameters
func SimulateMsgAddToGauge(cdc *codec.ProtoCodec, ak stakingTypes.AccountKeeper, bk osmosimtypes.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		simCoins := bk.SpendableCoins(ctx, simAccount.Address)
		if simCoins.AmountOf(sdk.DefaultBondDenom).LT(types.AddToGaugeFee) {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgAddToGauge, "Account have no coin"), nil, nil
		}

		gauge := RandomGauge(ctx, r, k)
		if gauge == nil {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgAddToGauge, "No gauge exists"), nil, nil
		} else if gauge.IsFinishedGauge(ctx.BlockTime()) {
			// TODO: Ideally we'd still run this but expect failure.
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgAddToGauge, "Selected a gauge that is finished"), nil, nil
		}

		rewards := genRewardCoins(r, simCoins, types.AddToGaugeFee)
		msg := types.MsgAddToGauge{
			Owner:   simAccount.Address.String(),
			GaugeId: gauge.Id,
			Rewards: rewards,
		}

		opMsg, err := osmosimtypes.GenerateAndDeliverTx(r, app, ctx, chainID, cdc, ak, bk, simAccount, types.ModuleName, &msg, typeMsgAddToGauge, false)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgAddToGauge, "unable to generate and deliver tx"), nil, err
		}

		return opMsg, nil, nil
	}
}

// RandomGauge takes a context, then returns a random existing gauge.
func RandomGauge(ctx sdk.Context, r *rand.Rand, k keeper.Keeper) *types.Gauge {
	gauges := k.GetGauges(ctx)
	if len(gauges) == 0 {
		return nil
	}
	return &gauges[r.Intn(len(gauges))]
}
