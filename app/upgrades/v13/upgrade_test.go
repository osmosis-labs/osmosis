package v13_test

import (
	"fmt"
	"reflect"
	"testing"

	gamm "github.com/osmosis-labs/osmosis/v13/x/gamm/keeper"

	ibcratelimittypes "github.com/osmosis-labs/osmosis/v13/x/ibc-rate-limit/types"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/osmosis-labs/osmosis/v13/app/apptesting"
	v13 "github.com/osmosis-labs/osmosis/v13/app/upgrades/v13"
	ibc_hooks "github.com/osmosis-labs/osmosis/v13/x/ibc-hooks"
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
	plan, exists := suite.App.UpgradeKeeper.GetUpgradePlan(suite.Ctx)
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
			func() { dummyUpgrade(suite) },
			func() {
				hasAcc := suite.App.AccountKeeper.HasAccount(suite.Ctx, ibc_hooks.WasmHookModuleAccountAddr)
				suite.Require().True(hasAcc)
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

func (suite *UpgradeTestSuite) TestMigrateNextPoolIdAndCreatePool() {
	suite.SetupTest() // reset

	const (
		expectedNextPoolId uint64 = 3
	)

	var (
		gammKeeperType = reflect.TypeOf(&gamm.Keeper{})
	)

	ctx := suite.Ctx
	gammKeeper := suite.App.GAMMKeeper
	swaprouterKeeper := suite.App.SwapRouterKeeper

	// Set next pool id to given constant, because creating pools doesn't
	// increment id on current version
	gammKeeper.SetNextPoolId(ctx, expectedNextPoolId)
	nextPoolId := gammKeeper.GetNextPoolId(ctx)
	suite.Require().Equal(expectedNextPoolId, nextPoolId)

	// system under test.
	v13.MigrateNextPoolId(ctx, gammKeeper, swaprouterKeeper)

	// validate swaprouter's next pool id.
	actualNextPoolId := swaprouterKeeper.GetNextPoolId(ctx)
	suite.Require().Equal(expectedNextPoolId, actualNextPoolId)

	// validate gamm pool count.
	actualGammPoolCount := gammKeeper.GetPoolCount(ctx)
	suite.Require().Equal(expectedNextPoolId-1, actualGammPoolCount)

	// create a pool after migration.
	actualCreatedPoolId := suite.PrepareBalancerPool()
	suite.Require().Equal(expectedNextPoolId, actualCreatedPoolId)

	// validate that module route mapping has been created for each pool id.
	for poolId := uint64(1); poolId < expectedNextPoolId; poolId++ {
		swapModule, err := swaprouterKeeper.GetSwapModule(ctx, poolId)
		suite.Require().NoError(err)

		suite.Require().Equal(gammKeeperType, reflect.TypeOf(swapModule))
	}
}
