package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	lockupkeeper "github.com/osmosis-labs/osmosis/v7/x/lockup/keeper"
	lockuptypes "github.com/osmosis-labs/osmosis/v7/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
)

// TODO: add test for MsgSuperfluidUndelegate
// TODO: add test for MsgSuperfluidUnbondLock
// TODO: add test for MsgLockAndSuperfluidDelegate

func (suite *KeeperTestSuite) TestMsgSuperfluidDelegate() {
	type param struct {
		coinsToLock         sdk.Coins
		lockOwner           sdk.AccAddress
		duration            time.Duration
		coinsInOwnerAddress sdk.Coins
	}

	tests := []struct {
		name       string
		param      param
		expectPass bool
	}{
		{
			name: "superfluid delegation for not allowed asset",
			param: param{
				coinsToLock:         sdk.Coins{sdk.NewInt64Coin("stake", 10)},       // setup wallet
				lockOwner:           sdk.AccAddress([]byte("addr1---------------")), // setup wallet
				duration:            time.Second,
				coinsInOwnerAddress: sdk.Coins{sdk.NewInt64Coin("stake", 10)},
			},
			expectPass: false,
		},
	}

	for _, test := range tests {
		suite.SetupTest()

		err := simapp.FundAccount(suite.App.BankKeeper, suite.Ctx, test.param.lockOwner, test.param.coinsInOwnerAddress)
		suite.Require().NoError(err)

		lockupMsgServer := lockupkeeper.NewMsgServerImpl(suite.App.LockupKeeper)
		c := sdk.WrapSDKContext(suite.Ctx)
		resp, err := lockupMsgServer.LockTokens(c, lockuptypes.NewMsgLockTokens(test.param.lockOwner, test.param.duration, test.param.coinsToLock))

		valAddrs := suite.SetupValidators([]stakingtypes.BondStatus{stakingtypes.Bonded})

		msgServer := keeper.NewMsgServerImpl(suite.App.SuperfluidKeeper)
		_, err = msgServer.SuperfluidDelegate(c, types.NewMsgSuperfluidDelegate(test.param.lockOwner, resp.ID, valAddrs[0]))

		if test.expectPass {
			suite.Require().NoError(err)
		} else {
			suite.Require().Error(err)
		}
	}
}
