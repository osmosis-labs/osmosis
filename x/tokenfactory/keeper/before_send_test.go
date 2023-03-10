package keeper_test

import (
	"fmt"
	"io/ioutil"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/x/tokenfactory/types"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type SendMsgTestCase struct {
	desc       string
	msg        func(denom string) *banktypes.MsgSend
	expectPass bool
}

func (suite *KeeperTestSuite) TestBeforeSendHook() {
	for _, tc := range []struct {
		desc     string
		wasmFile string
		sendMsgs []SendMsgTestCase
	}{
		{
			desc:     "should not allow sending 100 amount of *any* denom",
			wasmFile: "./testdata/no100.wasm",
			sendMsgs: []SendMsgTestCase{
				{
					desc: "sending 100 of factorydenom should not work",
					msg: func(factorydenom string) *banktypes.MsgSend {
						return banktypes.NewMsgSend(
							suite.TestAccs[0],
							suite.TestAccs[1],
							sdk.NewCoins(sdk.NewInt64Coin(factorydenom, 1)),
						)
					},
					expectPass: true,
				},
				{
					desc: "sending 100 of factorydenom should not work",
					msg: func(factorydenom string) *banktypes.MsgSend {
						return banktypes.NewMsgSend(
							suite.TestAccs[0],
							suite.TestAccs[1],
							sdk.NewCoins(sdk.NewInt64Coin(factorydenom, 100)),
						)
					},
					expectPass: false,
				},
				{
					desc: "sending 100 of a non-factorydenom should not work",
					msg: func(factorydenom string) *banktypes.MsgSend {
						return banktypes.NewMsgSend(
							suite.TestAccs[0],
							suite.TestAccs[1],
							sdk.NewCoins(sdk.NewInt64Coin(factorydenom, 1), sdk.NewInt64Coin("uosmo", 100)),
						)
					},
					expectPass: false,
				},
			},
		},
	} {
		suite.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			// setup test
			suite.SetupTest()

			// upload and instantiate wasm code
			wasmCode, err := ioutil.ReadFile(tc.wasmFile)
			suite.Require().NoError(err, "test: %v", tc.desc)
			codeID, _, err := suite.contractKeeper.Create(suite.Ctx, suite.TestAccs[0], wasmCode, nil)
			suite.Require().NoError(err, "test: %v", tc.desc)
			cosmwasmAddress, _, err := suite.contractKeeper.Instantiate(suite.Ctx, codeID, suite.TestAccs[0], suite.TestAccs[0], []byte("{}"), "", sdk.NewCoins())
			suite.Require().NoError(err, "test: %v", tc.desc)

			// create new denom
			res, err := suite.msgServer.CreateDenom(sdk.WrapSDKContext(suite.Ctx), types.NewMsgCreateDenom(suite.TestAccs[0].String(), "bitcoin"))
			suite.Require().NoError(err, "test: %v", tc.desc)
			denom := res.GetNewTokenDenom()

			// mint enough coins to the creator
			suite.msgServer.Mint(sdk.WrapSDKContext(suite.Ctx), types.NewMsgMint(suite.TestAccs[0].String(), sdk.NewInt64Coin(denom, 1000000000)))

			// set beforesend hook to the new denom
			_, err = suite.msgServer.SetBeforeSendHook(sdk.WrapSDKContext(suite.Ctx), types.NewMsgSetBeforeSendHook(suite.TestAccs[0].String(), denom, cosmwasmAddress.String()))
			suite.Require().NoError(err, "test: %v", tc.desc)

			for _, sendTc := range tc.sendMsgs {
				_, err := suite.bankMsgServer.Send(sdk.WrapSDKContext(suite.Ctx), sendTc.msg(denom))
				if sendTc.expectPass {
					suite.Require().NoError(err, "test: %v", sendTc.desc)
				} else {
					suite.Require().Error(err, "test: %v", sendTc.desc)
				}
			}
		})
	}
}
