package simulation

//// DONTCOVER
//
//import (
//	"math/rand"
//	"strings"
//
//	"github.com/cosmos/cosmos-sdk/baseapp"
//	"github.com/cosmos/cosmos-sdk/codec"
//	"github.com/cosmos/cosmos-sdk/testutil/sims"
//	sdk "github.com/cosmos/cosmos-sdk/types"
//	"github.com/cosmos/cosmos-sdk/types/module/testutil"
//	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
//
//	"github.com/cosmos/cosmos-sdk/x/simulation"
//	appParams "github.com/osmosis-labs/osmosis/v23/app/params"
//	"github.com/osmosis-labs/osmosis/v23/x/market/types"
//)
//
//// Simulation operation weights constants
//const (
//	OpWeightMsgSwap = "op_weight_msg_swap" //#nosec
//)
//
//// WeightedOperations returns all the operations from the module with their respective weights
//func WeightedOperations(
//	appParams simtypes.AppParams,
//	cdc codec.JSONCodec,
//	ak types.AccountKeeper,
//	bk types.BankKeeper,
//) simulation.WeightedOperations {
//	var weightMsgSwap int
//	appParams.GetOrGenerate(cdc, OpWeightMsgSwap, &weightMsgSwap, nil,
//		func(*rand.Rand) {
//			weightMsgSwap = 100
//		},
//	)
//
//	return simulation.WeightedOperations{
//		simulation.NewWeightedOperation(
//			weightMsgSwap,
//			SimulateMsgSwap(ak, bk),
//		),
//	}
//}
//
//// SimulateMsgSwap generates a MsgSwap with random values.
//// nolint: funlen
//func SimulateMsgSwap(
//	ak types.AccountKeeper,
//	bk types.BankKeeper,
//) simtypes.Operation {
//	return func(
//		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
//	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
//		simAccount, _ := simtypes.RandomAcc(r, accs)
//		account := ak.GetAccount(ctx, simAccount.Address)
//
//		spendable := bk.SpendableCoins(ctx, simAccount.Address)
//		fees, err := simtypes.RandomFees(r, ctx, spendable)
//		if err != nil {
//			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgSwap, "unable to generate fees"), nil, err
//		}
//
//		var whitelist []string
//		ok.IterateLunaExchangeRates(ctx, func(denom string, ex sdk.Dec) bool {
//			whitelist = append(whitelist, denom)
//			return false
//		})
//
//		var offerDenom string
//		var askDenom string
//		whitelistLen := len(whitelist)
//		if whitelistLen == 0 {
//			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgSwap, "no available exchange rates"), nil, nil
//		}
//
//		if randVal := simtypes.RandIntBetween(r, 0, whitelistLen*2); randVal < whitelistLen {
//			offerDenom = appParams.BaseCoinUnit
//			askDenom = whitelist[randVal]
//		} else {
//			offerDenom = whitelist[randVal-whitelistLen]
//			askDenom = appParams.BaseCoinUnit
//		}
//
//		amount := simtypes.RandomAmount(r, spendable.AmountOf(offerDenom).Sub(fees.AmountOf(offerDenom)))
//		if amount.Equal(sdk.ZeroInt()) {
//			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgSwap, "not enough offer denom amount"), nil, nil
//		}
//
//		msg := types.NewMsgSwap(simAccount.Address, sdk.NewCoin(offerDenom, amount), askDenom)
//
//		txGen := testutil.MakeTestEncodingConfig().TxConfig
//		tx, err := sims.GenSignedMockTx(
//			r,
//			txGen,
//			[]sdk.Msg{msg},
//			fees,
//			sims.DefaultGenTxGas,
//			chainID,
//			[]uint64{account.GetAccountNumber()},
//			[]uint64{account.GetSequence()},
//			simAccount.PrivKey,
//		)
//		if err != nil {
//			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to generate mock tx"), nil, err
//		}
//
//		_, _, err = app.SimDeliver(txGen.TxEncoder(), tx)
//		if err != nil {
//			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to deliver tx"), nil, err
//		}
//
//		return simtypes.NewOperationMsg(msg, true, "", nil), nil, nil
//	}
//}
//
//// SimulateMsgSwapSend generates a MsgSwapSend with random values.
//// nolint: funlen
//func SimulateMsgSwapSend(
//	ak types.AccountKeeper,
//	bk types.BankKeeper,
//	ok types.OracleKeeper,
//) simtypes.Operation {
//	return func(
//		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
//	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
//		simAccount, _ := simtypes.RandomAcc(r, accs)
//		receiverAccount, _ := simtypes.RandomAcc(r, accs)
//		account := ak.GetAccount(ctx, simAccount.Address)
//
//		spendable := bk.SpendableCoins(ctx, simAccount.Address)
//		fees, err := simtypes.RandomFees(r, ctx, spendable)
//		if err != nil {
//			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgSwapSend, "unable to generate fees"), nil, err
//		}
//
//		var whitelist []string
//		ok.IterateLunaExchangeRates(ctx, func(denom string, ex sdk.Dec) bool {
//			whitelist = append(whitelist, denom)
//			return false
//		})
//
//		var offerDenom string
//		var askDenom string
//		whitelistLen := len(whitelist)
//		if whitelistLen == 0 {
//			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgSwapSend, "no available exchange rates"), nil, nil
//		}
//
//		if randVal := simtypes.RandIntBetween(r, 0, whitelistLen*2); randVal < whitelistLen {
//			offerDenom = appParams.BaseCoinUnit
//			askDenom = whitelist[randVal]
//		} else {
//			offerDenom = whitelist[randVal-whitelistLen]
//			askDenom = appParams.BaseCoinUnit
//		}
//
//		// Check send_enabled status of offer denom
//		if !bk.IsSendEnabledCoin(ctx, sdk.Coin{Denom: offerDenom, Amount: sdk.NewInt(1)}) {
//			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgSwapSend, err.Error()), nil, nil
//		}
//
//		amount := simtypes.RandomAmount(r, spendable.AmountOf(offerDenom).Sub(fees.AmountOf(offerDenom)))
//		if amount.Equal(sdk.ZeroInt()) {
//			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgSwapSend, "not enough offer denom amount"), nil, nil
//		}
//
//		msg := types.NewMsgSwapSend(simAccount.Address, receiverAccount.Address, sdk.NewCoin(offerDenom, amount), askDenom)
//
//		txGen := testutil.MakeTestEncodingConfig().TxConfig
//		tx, err := sims.GenSignedMockTx(
//			r,
//			txGen,
//			[]sdk.Msg{msg},
//			fees,
//			sims.DefaultGenTxGas,
//			chainID,
//			[]uint64{account.GetAccountNumber()},
//			[]uint64{account.GetSequence()},
//			simAccount.PrivKey,
//		)
//		if err != nil {
//			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to generate mock tx"), nil, err
//		}
//
//		_, _, err = app.SimDeliver(txGen.TxEncoder(), tx)
//		if err != nil {
//			if strings.Contains(err.Error(), "insufficient fee") {
//				return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "ignore tax error"), nil, nil
//			}
//
//			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to deliver tx"), nil, err
//		}
//
//		return simtypes.NewOperationMsg(msg, true, "", nil), nil, nil
//	}
//}
