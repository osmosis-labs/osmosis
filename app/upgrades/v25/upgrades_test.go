package v25_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	v4 "github.com/cosmos/cosmos-sdk/x/slashing/migrations/v4"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	abci "github.com/cometbft/cometbft/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v24/app/apptesting"
	v25 "github.com/osmosis-labs/osmosis/v24/app/upgrades/v25"
	"github.com/osmosis-labs/osmosis/v24/x/cosmwasmpool/model"
	cwpooltypes "github.com/osmosis-labs/osmosis/v24/x/cosmwasmpool/types"
)

const (
	v25UpgradeHeight = int64(10)
)

var (
	consAddr = sdk.ConsAddress(sdk.AccAddress([]byte("addr1_______________")))
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestUpgrade() {
	s.Setup()

	// Create Astroport pools pre-upgrade
	for _, poolId := range v25.AstroportPoolIds {
		s.App.CosmwasmPoolKeeper.SetPool(s.Ctx, &model.CosmWasmPool{
			ContractAddress: "foo",
			PoolId:          poolId,
			CodeId:          580,
			InstantiateMsg:  []byte("bar"),
		})
	}
	s.requirePoolsHaveCodeId(v25.AstroportPoolIds[:], 580)

	preMigrationSigningInfo := s.prepareMissedBlocksCounterTest()

	// Run the upgrade
	dummyUpgrade(s)
	s.Require().NotPanics(func() {
		s.App.BeginBlocker(s.Ctx, abci.RequestBeginBlock{})
	})

	s.executeMissedBlocksCounterTest(preMigrationSigningInfo)

	// Pool Migration Tests
	//

	// Test that the Astroport pools have been updated
	s.requirePoolsHaveCodeId(v25.AstroportPoolIds[:], 666)

}

func dummyUpgrade(s *UpgradeTestSuite) {
	s.Ctx = s.Ctx.WithBlockHeight(v25UpgradeHeight - 1)
	plan := upgradetypes.Plan{Name: "v25", Height: v25UpgradeHeight}
	err := s.App.UpgradeKeeper.ScheduleUpgrade(s.Ctx, plan)
	s.Require().NoError(err)
	_, exists := s.App.UpgradeKeeper.GetUpgradePlan(s.Ctx)
	s.Require().True(exists)

	s.Ctx = s.Ctx.WithBlockHeight(v25UpgradeHeight)
}

func (s *UpgradeTestSuite) prepareMissedBlocksCounterTest() slashingtypes.ValidatorSigningInfo {
	cdc := moduletestutil.MakeTestEncodingConfig(slashing.AppModuleBasic{}).Codec
	slashingStoreKey := s.App.AppKeepers.GetKey(slashingtypes.StoreKey)
	store := s.Ctx.KVStore(slashingStoreKey)

	// Replicate current mainnet state where, someones missed block counter is greater than their actual missed blocks
	preMigrationSigningInfo := slashingtypes.ValidatorSigningInfo{
		Address:             consAddr.String(),
		StartHeight:         10,
		IndexOffset:         100,
		JailedUntil:         time.Time{},
		Tombstoned:          false,
		MissedBlocksCounter: 1000,
	}

	// Set the missed blocks for the validator
	for i := 0; i < 10; i++ {
		err := s.App.SlashingKeeper.SetMissedBlockBitmapValue(s.Ctx, consAddr, int64(i), true)
		s.Require().NoError(err)
	}

	// Validate that the missed block bitmap value is of length 10 (the real missed blocks value), differing from the missed blocks counter of 1000 (the incorrect value)
	missedBlocks, err := s.App.SlashingKeeper.GetValidatorMissedBlocks(s.Ctx, consAddr)
	s.Require().NoError(err)
	s.Require().Len(missedBlocks, 10)

	// store old signing info and bitmap entries
	bz := cdc.MustMarshal(&preMigrationSigningInfo)
	store.Set(v4.ValidatorSigningInfoKey(consAddr), bz)

	return preMigrationSigningInfo
}

func (s *UpgradeTestSuite) executeMissedBlocksCounterTest(preMigrationSigningInfo slashingtypes.ValidatorSigningInfo) {
	postMigrationSigningInfo, found := s.App.SlashingKeeper.GetValidatorSigningInfo(s.Ctx, consAddr)
	s.Require().True(found)

	// Check that the missed blocks counter was set to the correct value
	s.Require().Equal(int64(10), postMigrationSigningInfo.MissedBlocksCounter)

	// Check that all other fields are the same
	s.Require().Equal(preMigrationSigningInfo.Address, postMigrationSigningInfo.Address)
	s.Require().Equal(preMigrationSigningInfo.StartHeight, postMigrationSigningInfo.StartHeight)
	s.Require().Equal(preMigrationSigningInfo.IndexOffset, postMigrationSigningInfo.IndexOffset)
	s.Require().Equal(preMigrationSigningInfo.JailedUntil, postMigrationSigningInfo.JailedUntil)
	s.Require().Equal(preMigrationSigningInfo.Tombstoned, postMigrationSigningInfo.Tombstoned)
}

func (s *UpgradeTestSuite) requirePoolsHaveCodeId(pools []uint64, codeId uint64) {
	for _, poolId := range pools {
		pool, err := s.App.CosmwasmPoolKeeper.GetPool(s.Ctx, poolId)
		s.Require().NoError(err)
		cwPool, ok := pool.(cwpooltypes.CosmWasmExtension)
		s.Require().True(ok)
		s.Require().EqualValues(codeId, cwPool.GetCodeId())
	}
}
