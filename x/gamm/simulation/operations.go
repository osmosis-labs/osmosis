package simulation

import (
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"

	"github.com/cosmos/cosmos-sdk/codec"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/osmosis-labs/osmosis/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/x/gamm/types"
)

// Simulation operation weights constants
const (
	DefaultWeightMsgCreatePool int = 10
	OpWeightMsgCreatePool          = "op_weight_msg_create_pool"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONMarshaler, ak stakingTypes.AccountKeeper,
	bk stakingTypes.BankKeeper, k keeper.Keeper,
) simulation.WeightedOperations {
	var (
		weightMsgCreatePool int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgCreatePool, &weightMsgCreatePool, nil,
		func(_ *rand.Rand) {
			weightMsgCreatePool = simappparams.DefaultWeightMsgCreateValidator
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgCreatePool,
			SimulateMsgCreatePool(ak, bk, k),
		),
	}
}

func genFuturePoolGovernor(r *rand.Rand, addr sdk.Address, tokenList []string) string {
	choice := r.Int31n(4)
	if choice == 0 { // No governor
		return ""
	} else if choice == 1 { // Single address governor
		return addr.String()
	} else if choice == 2 { // LP token governor
		return "1d"
	} else { // Other token governor
		token := tokenList[r.Intn(len(tokenList))]
		return token + ",1d"
	}
}

func genPoolAssets(r *rand.Rand, acct simtypes.Account, coins sdk.Coins) []types.PoolAsset {
	numCoins := 2 + r.Intn(Min(coins.Len(), 6))
	denomIndices := r.Perm(numCoins)
	denoms := make([]string, numCoins)
	for i := 0; i < numCoins; i++ {
		denoms[i] = coins[denomIndices[i]].Denom
	}

	return []types.PoolAsset{}
}

func genPoolParams(r *rand.Rand, blockTime time.Time, assets []types.PoolAsset) types.PoolParams {
	swapFeeInt := int64(r.Intn(1e5))
	swapFee := sdk.NewDecWithPrec(swapFeeInt, 5)

	exitFeeInt := int64(r.Intn(1e5))
	exitFee := sdk.NewDecWithPrec(exitFeeInt, 5)

	// TODO: Randomly generate LBP params
	return types.PoolParams{
		SwapFee: swapFee,
		ExitFee: exitFee,
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

// SimulateMsgCreatePool generates a MsgCreatePool with random values
func SimulateMsgCreatePool(ak stakingTypes.AccountKeeper, bk stakingTypes.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		simCoins := bk.GetAllBalances(ctx, simAccount.Address)
		if simCoins.Len() <= 1 {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgCreatePool, "Account doesn't have 2 different coin types"), nil, nil
		}

		poolAssets := genPoolAssets(r, simAccount, simCoins)
		poolParams := genPoolParams(r, ctx.BlockTime(), poolAssets)

		// TODO: Replace []string{} with all token types on chain.
		futurePoolGovernor := genFuturePoolGovernor(r, simAccount.Address, []string{})
		msg := types.MsgCreatePool{
			Sender:             simAccount.Address.String(),
			FuturePoolGovernor: futurePoolGovernor,
			PoolAssets:         poolAssets,
			PoolParams:         poolParams,
		}

		account := ak.GetAccount(ctx, simAccount.Address)
		// spendable := bk.SpendableCoins(ctx, account.GetAddress())

		var fees sdk.Coins

		// coins, hasNeg := spendable.SafeSub(sdk.Coins{selfDelegation})
		// if !hasNeg {
		// 	fees, err = simtypes.RandomFees(r, ctx, coins)
		// 	if err != nil {
		// 		return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreateValidator, "unable to generate fees"), nil, err
		// 	}
		// }

		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		tx, err := helpers.GenTx(
			txGen,
			[]sdk.Msg{&msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			simAccount.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to generate mock tx"), nil, err
		}

		_, _, err = app.Deliver(txGen.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to deliver tx"), nil, err
		}

		return simtypes.NewOperationMsg(&msg, true, ""), nil, nil
	}
}
