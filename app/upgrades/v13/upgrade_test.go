package v13_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcratelimittypes "github.com/osmosis-labs/osmosis/v12/x/ibc-rate-limit/types"

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
			"Test that rate limits are setup",
			func() {},
			func() { dummyUpgrade(suite) },
			func() {
				// The contract has been uploaded and the param is set
				paramSpace, ok := suite.App.ParamsKeeper.GetSubspace(ibcratelimittypes.ModuleName)
				suite.Require().True(ok)
				var contract string
				paramSpace.GetIfExists(suite.Ctx, ibcratelimittypes.KeyContractAddress, &contract)
				suite.Require().NotEmpty(contract)

				// The quotas are configured?
				contractAddr, err := sdk.AccAddressFromBech32(contract)
				suite.Require().NoError(err)
				denoms := []string{
					"uosmo",
					"ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2", // atom
					"ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858", // usdc
					"ibc/EA1D43981D5C9A1C4AAEA9C23BB1D4FA126BA9BC7020A25E0AE4AA841EA25DC5", // weth
					"ibc/D1542AA8762DB13087D8364F3EA6509FD6F009A34F00426AF9E4F9FA85CBBF1F", // wbtc
				}
				for _, denom := range denoms {
					bytes, err := suite.App.WasmKeeper.QuerySmart(suite.Ctx, contractAddr, []byte(fmt.Sprintf(`{"get_quotas": {"channel_id": "any", "denom": "%s"}}`, denom)))
					suite.Require().NoError(err)
					suite.Require().Contains(string(bytes), "weekly")
				}
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
