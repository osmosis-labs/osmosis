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
	"github.com/osmosis-labs/osmosis/x/lockup/keeper"
	"github.com/osmosis-labs/osmosis/x/lockup/types"
)

// Simulation operation weights constants
const (
	DefaultWeightMsgLockTokens int = 10
	OpWeightMsgLockTokens          = "op_weight_msg_create_lockup"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONMarshaler, ak stakingTypes.AccountKeeper,
	bk stakingTypes.BankKeeper, k keeper.Keeper,
) simulation.WeightedOperations {
	var (
		weightMsgLockTokens int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgLockTokens, &weightMsgLockTokens, nil,
		func(_ *rand.Rand) {
			weightMsgLockTokens = DefaultWeightMsgLockTokens
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgLockTokens,
			SimulateMsgLockTokens(ak, bk, k),
		),
	}
}

func genLockTokens(r *rand.Rand, acct simtypes.Account, coins sdk.Coins) (res sdk.Coins) {
	numCoins := r.Intn(Min(coins.Len(), 6)-1)+1
	denomIndices := r.Perm(numCoins)
	for i := 0; i < numCoins; i++ {
		denom := coins[denomIndices[i]].Denom
		amt, _ := simtypes.RandPositiveInt(r, coins[i].Amount)
		res = append(res, sdk.Coin{Denom: denom, Amount: amt})
	}

	res.Sort()
	return
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

// SimulateMsgLockTokens generates a MsgLockTokens with random values
func SimulateMsgLockTokens(ak stakingTypes.AccountKeeper, bk stakingTypes.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		simCoins := bk.SpendableCoins(ctx, simAccount.Address)
		lockTokens := genLockTokens(r, simAccount, simCoins)

		// random duration within 1 hour
		durationSecs := r.Intn(1*60*60)
		duration := time.Duration(durationSecs) * time.Second

		msg := types.MsgLockTokens{
			Owner: simAccount.Address.String(),
			Duration: duration,
			Coins: lockTokens,
		}

		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		return osmo_simulation.GenAndDeliverTxWithRandFees(
			r, app, txGen, &msg, lockTokens, ctx, simAccount, ak, bk, types.ModuleName)
	}
}
