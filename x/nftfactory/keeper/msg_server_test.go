package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	nftfactory "github.com/osmosis-labs/osmosis/v12/x/nftfactory/keeper"
	"github.com/osmosis-labs/osmosis/v12/x/nftfactory/types"
)

func (suite *KeeperTestSuite) TestCreateDenom() {
	suite.SetupTest()

	type param struct {
		id        string
		sender    sdk.AccAddress
		denomName string
		data      string
	}

	tests := []struct {
		name       string
		param      param
		expectPass bool
	}{
		{
			name: "valid create Denom!",
			param: param{
				id:        "1",
				sender:    sdk.AccAddress([]byte("addr1---------------")),
				denomName: "bitcoin",
				data:      "",
			},
			expectPass: true,
		},
		{
			name: "create a denom thats already been created",
			param: param{
				id:        "1",
				sender:    sdk.AccAddress([]byte("addr1---------------")),
				denomName: "bitcoin",
				data:      "",
			},
			expectPass: false,
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {

			// setup message server
			msgServer := nftfactory.NewMsgServerImpl(suite.App.NftFactoryKeeper)
			c := sdk.WrapSDKContext(suite.Ctx)

			// call the create validator set preference
			_, err := msgServer.CreateDenom(c, types.NewMsgCreateDenom(test.param.sender, test.param.id, test.param.denomName, test.param.data))
			if test.expectPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
