package keeper_test

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"

	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	"github.com/osmosis-labs/osmosis/osmomath"
	appparams "github.com/osmosis-labs/osmosis/v25/app/params"
	"github.com/osmosis-labs/osmosis/v25/x/txfees/keeper"
	"github.com/osmosis-labs/osmosis/v25/x/txfees/types"
)

func (s *KeeperTestSuite) TestFeeDecorator() {
	s.SetupTest(false)

	mempoolFeeOpts := types.NewDefaultMempoolFeeOptions()
	mempoolFeeOpts.MinGasPriceForHighGasTx = osmomath.MustNewDecFromStr("0.0025")
	baseDenom, _ := s.App.TxFeesKeeper.GetBaseDenom(s.Ctx)
	consensusMinFeeAmt := int64(25)
	point1BaseDenomMinGasPrices := sdk.NewDecCoins(sdk.NewDecCoinFromDec(baseDenom,
		osmomath.MustNewDecFromStr("0.1")))

	// uion is setup with a relative price of 1:1
	uion := "uion"

	type testcase struct {
		name         string
		txFee        sdk.Coins
		minGasPrices sdk.DecCoins // if blank, set to 0
		gasRequested uint64       // if blank, set to base gas
		isCheckTx    bool
		isSimulate   bool // if blank, is false
		expectPass   bool
	}

	tests := []testcase{}
	txType := []string{"delivertx", "checktx"}
	succesType := []string{"does", "doesn't"}
	for isCheckTx := 0; isCheckTx <= 1; isCheckTx++ {
		tests = append(tests, []testcase{
			{
				name:       fmt.Sprintf("no min gas price - %s. Fails w/ consensus minimum", txType[isCheckTx]),
				txFee:      sdk.NewCoins(),
				isCheckTx:  isCheckTx == 1,
				expectPass: false,
			},
			{
				name:       fmt.Sprintf("LT Consensus min gas price - %s", txType[isCheckTx]),
				txFee:      sdk.NewCoins(sdk.NewInt64Coin(baseDenom, consensusMinFeeAmt-1)),
				isCheckTx:  isCheckTx == 1,
				expectPass: false,
			},
			{
				name:       fmt.Sprintf("Consensus min gas price - %s", txType[isCheckTx]),
				txFee:      sdk.NewCoins(sdk.NewInt64Coin(baseDenom, consensusMinFeeAmt)),
				isCheckTx:  isCheckTx == 1,
				expectPass: true,
			},
			{
				name:       fmt.Sprintf("multiple fee coins - %s", txType[isCheckTx]),
				txFee:      sdk.NewCoins(sdk.NewInt64Coin(baseDenom, 1), sdk.NewInt64Coin(uion, 1)),
				isCheckTx:  isCheckTx == 1,
				expectPass: false,
			},
			{
				name:         fmt.Sprintf("works with valid basedenom fee - %s", txType[isCheckTx]),
				txFee:        sdk.NewCoins(sdk.NewInt64Coin(baseDenom, 1000)),
				minGasPrices: point1BaseDenomMinGasPrices,
				isCheckTx:    isCheckTx == 1,
				expectPass:   true,
			},
			{
				name:         fmt.Sprintf("works with valid converted fee - %s", txType[isCheckTx]),
				txFee:        sdk.NewCoins(sdk.NewInt64Coin(uion, 1000)),
				minGasPrices: point1BaseDenomMinGasPrices,
				isCheckTx:    isCheckTx == 1,
				expectPass:   true,
			},
			{
				name:         fmt.Sprintf("%s work with insufficient mempool fee in %s", succesType[isCheckTx], txType[isCheckTx]),
				txFee:        sdk.NewCoins(sdk.NewInt64Coin(baseDenom, consensusMinFeeAmt)), // consensus minimum
				minGasPrices: point1BaseDenomMinGasPrices,
				isCheckTx:    isCheckTx == 1,
				expectPass:   isCheckTx != 1,
			},
			{
				name:         fmt.Sprintf("%s work with insufficient converted mempool fee in %s", succesType[isCheckTx], txType[isCheckTx]),
				txFee:        sdk.NewCoins(sdk.NewInt64Coin(uion, 25)), // consensus minimum
				minGasPrices: point1BaseDenomMinGasPrices,
				isCheckTx:    isCheckTx == 1,
				expectPass:   isCheckTx != 1,
			},
			{
				name:       "invalid fee denom",
				txFee:      sdk.NewCoins(sdk.NewInt64Coin("moooooo", 1000)),
				isCheckTx:  isCheckTx == 1,
				expectPass: false,
			},
		}...)
	}

	custTests := []testcase{
		{
			name:         "min gas price not containing basedenom gets treated as min gas price 0",
			txFee:        sdk.NewCoins(sdk.NewInt64Coin(uion, 1000)),
			minGasPrices: sdk.NewDecCoins(sdk.NewInt64DecCoin(uion, 1000000)),
			isCheckTx:    true,
			expectPass:   true,
		},
		{
			name:         "tx with gas wanted more than allowed should not pass",
			txFee:        sdk.NewCoins(sdk.NewInt64Coin(uion, 100000000)),
			minGasPrices: sdk.NewDecCoins(sdk.NewInt64DecCoin(uion, 1)),
			gasRequested: mempoolFeeOpts.MaxGasWantedPerTx + 1,
			isCheckTx:    true,
			expectPass:   false,
		},
		{
			name:         "tx with high gas and not enough fee should no pass",
			txFee:        sdk.NewCoins(sdk.NewInt64Coin(uion, 1)),
			minGasPrices: sdk.NewDecCoins(sdk.NewInt64DecCoin(uion, 1)),
			gasRequested: mempoolFeeOpts.HighGasTxThreshold,
			isCheckTx:    true,
			expectPass:   false,
		},
		{
			name:         "tx with high gas and enough fee should pass",
			txFee:        sdk.NewCoins(sdk.NewInt64Coin(uion, 10*1000)),
			minGasPrices: sdk.NewDecCoins(sdk.NewInt64DecCoin(uion, 1)),
			gasRequested: mempoolFeeOpts.HighGasTxThreshold,
			isCheckTx:    true,
			expectPass:   true,
		},
		{
			name:         "simulate 0 fee passes",
			txFee:        sdk.Coins{},
			minGasPrices: sdk.NewDecCoins(sdk.NewInt64DecCoin(uion, 1)),
			gasRequested: mempoolFeeOpts.HighGasTxThreshold,
			isCheckTx:    true,
			isSimulate:   true,
			expectPass:   true,
		},
	}
	tests = append(tests, custTests...)

	for _, tc := range tests {
		// reset pool and accounts for each test
		s.SetupTest(false)
		s.Run(tc.name, func() {
			// See DeductFeeDecorator AnteHandler for how this is used
			s.FundAcc(sdk.MustAccAddressFromBech32("osmo1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqmcn030"), sdk.NewCoins(sdk.NewInt64Coin(appparams.BaseCoinUnit, 1)))

			err := s.SetupTxFeeAnteHandlerAndChargeFee(s.clientCtx, tc.minGasPrices, tc.gasRequested, tc.isCheckTx, tc.isSimulate, tc.txFee)
			if tc.expectPass {
				// ensure fee was collected
				if !tc.txFee.IsZero() {
					moduleName := authtypes.FeeCollectorName
					if tc.txFee[0].Denom != baseDenom {
						moduleName = types.NonNativeTxFeeCollectorName
					}
					moduleAddr := s.App.AccountKeeper.GetModuleAddress(moduleName)
					s.Require().Equal(tc.txFee[0], s.App.BankKeeper.GetBalance(s.Ctx, moduleAddr, tc.txFee[0].Denom), tc.name)
				}
				s.Require().NoError(err, "test: %s", tc.name)
			} else {
				s.Require().Error(err, "test: %s", tc.name)
			}
		})
	}
}

func (s *KeeperTestSuite) TestMempoolFeeDecorator_AnteHandle_MsgTransfer() {
	s.SetupTest(false)
	mfd := keeper.NewMempoolFeeDecorator(*s.App.TxFeesKeeper, types.NewDefaultMempoolFeeOptions())

	// Test cases
	testCases := []struct {
		name        string
		msg         sdk.Msg
		expectedErr error
	}{
		{
			name: "MsgTransfer with valid size",
			msg: &transfertypes.MsgTransfer{
				SourcePort:       "transfer",
				SourceChannel:    "channel-0",
				Token:            sdk.NewCoin("uosmo", sdk.NewInt(1000)),
				Sender:           "osmo1sender",
				Receiver:         "osmo1receiver",
				TimeoutHeight:    clienttypes.Height{},
				TimeoutTimestamp: 0,
				Memo:             "valid memo",
			},
		},
		{
			name: "MsgTransfer in total too large",
			msg: &transfertypes.MsgTransfer{
				SourcePort:       "transfer",
				SourceChannel:    "channel-0",
				Token:            sdk.NewCoin("uosmo", sdk.NewInt(1000)),
				Sender:           string(make([]byte, 35001)),
				Receiver:         string(make([]byte, 65000)),
				TimeoutHeight:    clienttypes.Height{},
				TimeoutTimestamp: 0,
				Memo:             string(make([]byte, 400000)),
			},
			expectedErr: errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "msg size is too large"),
		},
		{
			name: "MsgTransfer with memo too large",
			msg: &transfertypes.MsgTransfer{
				SourcePort:       "transfer",
				SourceChannel:    "channel-0",
				Token:            sdk.NewCoin("uosmo", sdk.NewInt(1000)),
				Sender:           "osmo1sender",
				Receiver:         "osmo1receiver",
				TimeoutHeight:    clienttypes.Height{},
				TimeoutTimestamp: 0,
				Memo:             string(make([]byte, 400001)), // 400KB + 1
			},
			expectedErr: errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "memo is too large"),
		},
		{
			name: "MsgTransfer with receiver too large",
			msg: &transfertypes.MsgTransfer{
				SourcePort:       "transfer",
				SourceChannel:    "channel-0",
				Token:            sdk.NewCoin("uosmo", sdk.NewInt(1000)),
				Sender:           "osmo1sender",
				Receiver:         string(make([]byte, 65001)), // 65KB + 1
				TimeoutHeight:    clienttypes.Height{},
				TimeoutTimestamp: 0,
				Memo:             "valid memo",
			},
			expectedErr: errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "receiver address is too large"),
		},
		{
			name: "MsgSendTx with valid packet data size",
			msg: &icacontrollertypes.MsgSendTx{
				Owner:        "osmo1owner",
				ConnectionId: "connection-0",
				PacketData: icatypes.InterchainAccountPacketData{
					Type: icatypes.EXECUTE_TX,
					Data: make([]byte, 400000),
					Memo: "valid memo",
				},
			},
		},
		{
			name: "MsgSendTx with packet data size too large",
			msg: &icacontrollertypes.MsgSendTx{
				Owner:        "osmo1owner",
				ConnectionId: "connection-0",
				PacketData: icatypes.InterchainAccountPacketData{
					Type: icatypes.EXECUTE_TX,
					Data: make([]byte, 400000),         // 400KB
					Memo: string(make([]byte, 100000)), // 100KB
				},
			},
			expectedErr: errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "packet data is too large"),
		},
		{
			name: "MsgSendTx with owner address too large",
			msg: &icacontrollertypes.MsgSendTx{
				Owner:        string(make([]byte, 65001)), // 65KB + 1,
				ConnectionId: "connection-0",
				PacketData: icatypes.InterchainAccountPacketData{
					Type: icatypes.EXECUTE_TX,
					Data: make([]byte, 400000),
					Memo: "valid memo",
				},
			},
			expectedErr: errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "owner address is too large"),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			baseDenom, _ := s.App.TxFeesKeeper.GetBaseDenom(s.Ctx)
			txFee := sdk.NewCoins(sdk.NewCoin(baseDenom, sdk.NewInt(250000)))
			tx, err := s.prepareTx(tc.msg, txFee)
			s.Require().NoError(err)

			_, err = mfd.AnteHandle(s.Ctx, tx, false, nextAnteHandler)

			if tc.expectedErr != nil {
				s.Require().Error(err)
				s.Require().Equal(tc.expectedErr.Error(), err.Error())
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) prepareTx(msg sdk.Msg, txFee sdk.Coins) (sdk.Tx, error) {
	txBuilder := s.clientCtx.TxConfig.NewTxBuilder()
	priv0, _, addr0 := testdata.KeyTestPubAddr()
	acc1 := s.App.AccountKeeper.NewAccountWithAddress(s.Ctx, addr0)
	s.App.AccountKeeper.SetAccount(s.Ctx, acc1)
	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv0}, []uint64{0}, []uint64{0}
	signerData := authsigning.SignerData{
		ChainID:       s.Ctx.ChainID(),
		AccountNumber: accNums[0],
		Sequence:      accSeqs[0],
	}

	sigV2, err := clienttx.SignWithPrivKey(
		1,
		signerData,
		txBuilder,
		privs[0],
		s.clientCtx.TxConfig,
		accSeqs[0],
	)
	if err != nil {
		return nil, err
	}

	err = testutil.FundAccount(s.App.BankKeeper, s.Ctx, addr0, txFee)
	if err != nil {
		return nil, err
	}

	tx := s.BuildTx(txBuilder, []sdk.Msg{msg}, sigV2, "", txFee, 100000000)
	return tx, nil
}

func nextAnteHandler(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
	return ctx, nil
}
