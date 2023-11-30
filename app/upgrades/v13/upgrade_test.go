package v13_test

import (
	"fmt"
	"testing"

	ibchookstypes "github.com/osmosis-labs/osmosis/x/ibc-hooks/types"

	ibcratelimittypes "github.com/osmosis-labs/osmosis/v21/x/ibc-rate-limit/types"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v21/app/apptesting"
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
}

func (s *UpgradeTestSuite) SetupTest() {
	s.Setup()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

const dummyUpgradeHeight = 5

func dummyUpgrade(s *UpgradeTestSuite) {
	s.Ctx = s.Ctx.WithBlockHeight(dummyUpgradeHeight - 1)
	plan := upgradetypes.Plan{Name: "v13", Height: dummyUpgradeHeight}
	err := s.App.UpgradeKeeper.ScheduleUpgrade(s.Ctx, plan)
	s.Require().NoError(err)
	_, exists := s.App.UpgradeKeeper.GetUpgradePlan(s.Ctx)
	s.Require().True(exists)

	s.Ctx = s.Ctx.WithBlockHeight(dummyUpgradeHeight)
	s.Require().NotPanics(func() {
		beginBlockRequest := abci.RequestBeginBlock{}
		s.App.BeginBlocker(s.Ctx, beginBlockRequest)
	})
}

func (s *UpgradeTestSuite) TestUpgrade() {
	s.SkipIfWSL()
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
				// acc := s.App.AccountKeeper.GetAccount(s.Ctx, ibc_hooks.WasmHookModuleAccountAddr)
				// s.App.AccountKeeper.RemoveAccount(s.Ctx, acc)

				// Because of SDK version map bug, we can't do the following, and instaed do a massive hack
				// vm := s.App.UpgradeKeeper.GetModuleVersionMap(s.Ctx)
				// delete(vm, ibchookstypes.ModuleName)
				// OR
				// vm[ibchookstypes.ModuleName] = 0
				// s.App.UpgradeKeeper.SetModuleVersionMap(s.Ctx, vm)
				upgradeStoreKey := s.App.AppKeepers.GetKey(upgradetypes.StoreKey)
				store := s.Ctx.KVStore(upgradeStoreKey)
				versionStore := prefix.NewStore(store, []byte{upgradetypes.VersionMapByte})
				versionStore.Delete([]byte(ibchookstypes.ModuleName))

				// Same comment as above: this was the case when the upgrade happened, but we don't have accounts anymore
				// hasAcc := s.App.AccountKeeper.HasAccount(s.Ctx, ibc_hooks.WasmHookModuleAccountAddr)
				// s.Require().False(hasAcc)
			},
			func() { dummyUpgrade(s) },
			func() {
				// Same comment as pre-upgrade. We had an account, but now we don't anymore
				// hasAcc := s.App.AccountKeeper.HasAccount(s.Ctx, ibc_hooks.WasmHookModuleAccountAddr)
				// s.Require().True(hasAcc)
			},
		},
		{
			"Test that the contract address is set in the params",
			func() {},
			func() { dummyUpgrade(s) },
			func() {
				// The contract has been uploaded and the param is set
				paramSpace, ok := s.App.ParamsKeeper.GetSubspace(ibcratelimittypes.ModuleName)
				s.Require().True(ok)
				var contract string
				paramSpace.GetIfExists(s.Ctx, ibcratelimittypes.KeyContractAddress, &contract)
				s.Require().NotEmpty(contract)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.name), func() {
			s.SetupTest() // reset

			tc.pre_upgrade()
			tc.upgrade()
			tc.post_upgrade()
		})
	}
}
