package lockuptesting

import (
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v7/app/apptesting"

	lockupkeeper "github.com/osmosis-labs/osmosis/v7/x/lockup/keeper"
	lockuptypes "github.com/osmosis-labs/osmosis/v7/x/lockup/types"
)

type LockupTestHelper struct {
	*apptesting.KeeperTestHelper
}

func (lockupTestHelper *LockupTestHelper) LockTokens(addr sdk.AccAddress, coins sdk.Coins, duration time.Duration) (lockID uint64) {
	msgServer := lockupkeeper.NewMsgServerImpl(lockupTestHelper.App.LockupKeeper)
	err := simapp.FundAccount(lockupTestHelper.App.BankKeeper, lockupTestHelper.Ctx, addr, coins)
	lockupTestHelper.Require().NoError(err)
	msgResponse, err := msgServer.LockTokens(sdk.WrapSDKContext(lockupTestHelper.Ctx), lockuptypes.NewMsgLockTokens(addr, duration, coins))
	lockupTestHelper.Require().NoError(err)
	return msgResponse.ID
}
