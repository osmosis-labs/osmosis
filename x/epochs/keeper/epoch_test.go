package keeper_test

import (
	"time"

	"osmosis.io/x/epochs/types"
)

func (suite *KeeperTestSuite) TestAddEpochInfo() {
	defaultIdentifier := "default_add_epoch_info_id"
	defaultDuration := time.Hour
	startBlockHeight := int64(100)
	startBlockTime := time.Unix(1656907200, 0).UTC()
	tests := map[string]struct {
		addedEpochInfo types.EpochInfo
		expErr         bool
		expEpochInfo   types.EpochInfo
	}{
		"simple_add": {
			addedEpochInfo: types.EpochInfo{
				Identifier:              defaultIdentifier,
				StartTime:               time.Time{},
				Duration:                defaultDuration,
				CurrentEpoch:            0,
				CurrentEpochStartHeight: 0,
				CurrentEpochStartTime:   time.Time{},
				EpochCountingStarted:    false,
			},
			expErr: false,
			expEpochInfo: types.EpochInfo{
				Identifier:              defaultIdentifier,
				StartTime:               startBlockTime,
				Duration:                defaultDuration,
				CurrentEpoch:            0,
				CurrentEpochStartHeight: startBlockHeight,
				CurrentEpochStartTime:   time.Time{},
				EpochCountingStarted:    false,
			},
		},
		"zero_duration": {
			addedEpochInfo: types.EpochInfo{
				Identifier:              defaultIdentifier,
				StartTime:               time.Time{},
				Duration:                time.Duration(0),
				CurrentEpoch:            0,
				CurrentEpochStartHeight: 0,
				CurrentEpochStartTime:   time.Time{},
				EpochCountingStarted:    false,
			},
			expErr: true,
		},
	}
	for name, test := range tests {
		suite.Run(name, func() {
			suite.SetupTest()
			suite.Ctx = suite.Ctx.WithBlockHeight(startBlockHeight).WithBlockTime(startBlockTime)
			err := suite.App.EpochsKeeper.AddEpochInfo(suite.Ctx, test.addedEpochInfo)
			if !test.expErr {
				suite.Require().NoError(err)
				actualEpochInfo := suite.App.EpochsKeeper.GetEpochInfo(suite.Ctx, test.addedEpochInfo.Identifier)
				suite.Require().Equal(test.expEpochInfo, actualEpochInfo)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestDuplicateAddEpochInfo() {
	identifier := "duplicate_add_epoch_info"
	epochInfo := types.NewGenesisEpochInfo(identifier, time.Hour*24*30)
	err := suite.App.EpochsKeeper.AddEpochInfo(suite.Ctx, epochInfo)
	suite.Require().NoError(err)
	err = suite.App.EpochsKeeper.AddEpochInfo(suite.Ctx, epochInfo)
	suite.Require().Error(err)
}

func (suite *KeeperTestSuite) TestEpochLifeCycle() {
	suite.SetupTest()

	epochInfo := types.NewGenesisEpochInfo("monthly", time.Hour*24*30)
	suite.App.EpochsKeeper.AddEpochInfo(suite.Ctx, epochInfo)
	epochInfoSaved := suite.App.EpochsKeeper.GetEpochInfo(suite.Ctx, "monthly")
	// setup expected epoch info
	expectedEpochInfo := epochInfo
	expectedEpochInfo.StartTime = suite.Ctx.BlockTime()
	expectedEpochInfo.CurrentEpochStartHeight = suite.Ctx.BlockHeight()
	suite.Require().Equal(expectedEpochInfo, epochInfoSaved)

	allEpochs := suite.App.EpochsKeeper.AllEpochInfos(suite.Ctx)
	suite.Require().Len(allEpochs, 4)
	suite.Require().Equal(allEpochs[0].Identifier, "day") // alphabetical order
	suite.Require().Equal(allEpochs[1].Identifier, "hour")
	suite.Require().Equal(allEpochs[2].Identifier, "monthly")
	suite.Require().Equal(allEpochs[3].Identifier, "week")
}
