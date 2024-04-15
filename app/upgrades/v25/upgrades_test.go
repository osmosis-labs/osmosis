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

	preMigrationSigningInfo := s.prepareMissedBlocksCounter()

	// Run the upgrade
	dummyUpgrade(s)
	s.Require().NotPanics(func() {
		s.App.BeginBlocker(s.Ctx, abci.RequestBeginBlock{})
	})

	s.executeMissedBlocksCounter(preMigrationSigningInfo)
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

func (s *UpgradeTestSuite) prepareMissedBlocksCounter() slashingtypes.ValidatorSigningInfo {
	cdc := moduletestutil.MakeTestEncodingConfig(slashing.AppModuleBasic{}).Codec
	slashingStoreKey := s.App.AppKeepers.GetKey(slashingtypes.StoreKey)
	store := s.Ctx.KVStore(slashingStoreKey)

	preMigrationSigningInfo := slashingtypes.ValidatorSigningInfo{
		Address:             consAddr.String(),
		StartHeight:         10,
		IndexOffset:         100,
		JailedUntil:         time.Time{},
		Tombstoned:          false,
		MissedBlocksCounter: 1000,
	}

	// store old signing info and bitmap entries
	bz := cdc.MustMarshal(&preMigrationSigningInfo)
	store.Set(v4.ValidatorSigningInfoKey(consAddr), bz)

	return preMigrationSigningInfo
}

func (s *UpgradeTestSuite) executeMissedBlocksCounter(preMigrationSigningInfo slashingtypes.ValidatorSigningInfo) {
	postMigrationSigningInfo, found := s.App.SlashingKeeper.GetValidatorSigningInfo(s.Ctx, consAddr)
	s.Require().True(found)

	// Check that the missed blocks counter was reset to zero
	s.Require().Equal(int64(0), postMigrationSigningInfo.MissedBlocksCounter)

	// Check that all other fields are the same
	s.Require().Equal(preMigrationSigningInfo.Address, postMigrationSigningInfo.Address)
	s.Require().Equal(preMigrationSigningInfo.StartHeight, postMigrationSigningInfo.StartHeight)
	s.Require().Equal(preMigrationSigningInfo.IndexOffset, postMigrationSigningInfo.IndexOffset)
	s.Require().Equal(preMigrationSigningInfo.JailedUntil, postMigrationSigningInfo.JailedUntil)
	s.Require().Equal(preMigrationSigningInfo.Tombstoned, postMigrationSigningInfo.Tombstoned)
}
