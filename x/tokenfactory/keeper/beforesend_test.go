package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type SendMsgTestCase struct {
	desc       string
	msg        func(denom string) *banktypes.MsgSend
	expectPass bool
}

func (suite *KeeperTestSuite) TestBeforeSendListener() {
	for _, tc := range []struct {
		desc     string
		wasmFile string
		sendMsgs []SendMsgTestCase
	}{
		{
			desc: "should not allow sending 100 amount of *any* denom",
			sendMsgs: []SendMsgTestCase{
				{
					desc: "sending 1 of factorydenom should not error",
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
					desc: "sending 100 of non-factorydenom should not error",
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
					desc: "sending 100 of factorydenom should work",
					msg: func(factorydenom string) *banktypes.MsgSend {
						return banktypes.NewMsgSend(
							suite.TestAccs[0],
							suite.TestAccs[1],
							sdk.NewCoins(sdk.NewInt64Coin("foo", 100)),
						)
					},
					expectPass: true,
				},
				{
					desc: "having 100 coin within coins should not work",
					msg: func(factorydenom string) *banktypes.MsgSend {
						return banktypes.NewMsgSend(
							suite.TestAccs[0],
							suite.TestAccs[1],
							sdk.NewCoins(sdk.NewInt64Coin(factorydenom, 100), sdk.NewInt64Coin("foo", 1)),
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

			suite.FundAcc(suite.TestAccs[0], sdk.Coins{sdk.NewInt64Coin("foo", 100000000000)})

			// set basic token factory listener
			tokenFactoryDenom := suite.SetBasicTokenFactoryDenom()
			suite.SetBasicTokenFacotryListener(tokenFactoryDenom)

			for _, sendTc := range tc.sendMsgs {
				_, err := suite.bankMsgServer.Send(sdk.WrapSDKContext(suite.Ctx), sendTc.msg(tokenFactoryDenom))
				if sendTc.expectPass {
					suite.Require().NoError(err, "test: %v", sendTc.desc)
				} else {
					suite.Require().Error(err, "test: %v", sendTc.desc)
				}
			}
		})
	}
}
