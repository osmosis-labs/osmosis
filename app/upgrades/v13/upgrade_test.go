package v13_test

import (
	"fmt"
	"testing"

	ibchookstypes "github.com/osmosis-labs/osmosis/x/ibc-hooks/types"

	ibcratelimittypes "github.com/osmosis-labs/osmosis/v15/x/ibc-rate-limit/types"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
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

func dummyUpgrade(suite *UpgradeTestSuite) {
	suite.Ctx = suite.Ctx.WithBlockHeight(dummyUpgradeHeight - 1)
	plan := upgradetypes.Plan{Name: "v13", Height: dummyUpgradeHeight}
	err := suite.App.UpgradeKeeper.ScheduleUpgrade(suite.Ctx, plan)
	suite.Require().NoError(err)
	_, exists := suite.App.UpgradeKeeper.GetUpgradePlan(suite.Ctx)
	suite.Require().True(exists)

	suite.Ctx = suite.Ctx.WithBlockHeight(dummyUpgradeHeight)
	suite.Require().NotPanics(func() {
		beginBlockRequest := abci.RequestBeginBlock{}
		suite.App.BeginBlocker(suite.Ctx, beginBlockRequest)
	})
}

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
				// The module doesn't need an account anymore, but when the upgrade happened we did:
				// acc := suite.App.AccountKeeper.GetAccount(suite.Ctx, ibc_hooks.WasmHookModuleAccountAddr)
				// suite.App.AccountKeeper.RemoveAccount(suite.Ctx, acc)

				// Because of SDK version map bug, we can't do the following, and instaed do a massive hack
				// vm := suite.App.UpgradeKeeper.GetModuleVersionMap(suite.Ctx)
				// delete(vm, ibchookstypes.ModuleName)
				// OR
				// vm[ibchookstypes.ModuleName] = 0
				// suite.App.UpgradeKeeper.SetModuleVersionMap(suite.Ctx, vm)
				upgradeStoreKey := suite.App.AppKeepers.GetKey(upgradetypes.StoreKey)
				store := suite.Ctx.KVStore(upgradeStoreKey)
				versionStore := prefix.NewStore(store, []byte{upgradetypes.VersionMapByte})
				versionStore.Delete([]byte(ibchookstypes.ModuleName))

				// Same comment as above: this was the case when the upgrade happened, but we don't have accounts anymore
				// hasAcc := suite.App.AccountKeeper.HasAccount(suite.Ctx, ibc_hooks.WasmHookModuleAccountAddr)
				// suite.Require().False(hasAcc)
			},
			func() { dummyUpgrade(suite) },
			func() {
				// Same comment as pre-upgrade. We had an account, but now we don't anymore
				// hasAcc := suite.App.AccountKeeper.HasAccount(suite.Ctx, ibc_hooks.WasmHookModuleAccountAddr)
				// suite.Require().True(hasAcc)
			},
		},
		{
			"Test that the contract address is set in the params",
			func() {},
			func() { dummyUpgrade(suite) },
			func() {
				// The contract has been uploaded and the param is set
				paramSpace, ok := suite.App.ParamsKeeper.GetSubspace(ibcratelimittypes.ModuleName)
				suite.Require().True(ok)
				var contract string
				paramSpace.GetIfExists(suite.Ctx, ibcratelimittypes.KeyContractAddress, &contract)
				suite.Require().NotEmpty(contract)
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
