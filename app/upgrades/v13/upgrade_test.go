package v13_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/osmosis-labs/osmosis/v12/app/apptesting"
	ibc_hooks "github.com/osmosis-labs/osmosis/v12/x/ibc-hooks"
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
}

func (suite *UpgradeTestSuite) SetupTest() {
	suite.Setup()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

const dummyUpgradeHeight = 5

func (suite *UpgradeTestSuite) TestUpgrade() {
	testCases := []struct {
		name         string
		pre_upgrade  func()
		upgrade      func()
		post_upgrade func()
	}{
		{
			"Test that the upgrade succeeds",
			func() {
				acc := suite.App.AccountKeeper.GetAccount(suite.Ctx, ibc_hooks.WasmHookModuleAccountAddr)
				suite.App.AccountKeeper.RemoveAccount(suite.Ctx, acc)
				// Because of SDK version map bug, we can't do the following, and instaed do a massive hack
				// vm := suite.App.UpgradeKeeper.GetModuleVersionMap(suite.Ctx)
				// delete(vm, ibc_hooks.ModuleName)
				// OR
				// vm[ibc_hooks.ModuleName] = 0
				// suite.App.UpgradeKeeper.SetModuleVersionMap(suite.Ctx, vm)
				upgradeStoreKey := suite.App.AppKeepers.GetKey(upgradetypes.StoreKey)
				store := suite.Ctx.KVStore(upgradeStoreKey)
				versionStore := prefix.NewStore(store, []byte{upgradetypes.VersionMapByte})
				versionStore.Delete([]byte(ibc_hooks.ModuleName))

				hasAcc := suite.App.AccountKeeper.HasAccount(suite.Ctx, ibc_hooks.WasmHookModuleAccountAddr)
				suite.Require().False(hasAcc)
			},
			func() {
				suite.Ctx = suite.Ctx.WithBlockHeight(dummyUpgradeHeight - 1)
				plan := upgradetypes.Plan{Name: "v13", Height: dummyUpgradeHeight}
				err := suite.App.UpgradeKeeper.ScheduleUpgrade(suite.Ctx, plan)
				suite.Require().NoError(err)
				plan, exists := suite.App.UpgradeKeeper.GetUpgradePlan(suite.Ctx)
				suite.Require().True(exists)

				suite.Ctx = suite.Ctx.WithBlockHeight(dummyUpgradeHeight)
				suite.Require().NotPanics(func() {
					beginBlockRequest := abci.RequestBeginBlock{}
					suite.App.BeginBlocker(suite.Ctx, beginBlockRequest)
				})
			},
			func() {
				hasAcc := suite.App.AccountKeeper.HasAccount(suite.Ctx, ibc_hooks.WasmHookModuleAccountAddr)
				suite.Require().True(hasAcc)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.pre_upgrade()
			tc.upgrade()
			tc.post_upgrade()
		})
	}
}
