package keeper_test

import (
	"github.com/osmosis-labs/osmosis/v14/x/gamm/types"
)

func (suite *KeeperTestSuite) TestReplaceDistrRecords() {
	tests := []struct {
		name                    string
		testingMigrationRecords []types.GammToConcentratedPoolLink
		isPoolPrepared          bool
		expectErr               bool
	}{
		{
			name: "Non existent pool.",
			testingMigrationRecords: []types.GammToConcentratedPoolLink{{
				GammPoolId: 1,
				ClPoolId:   3,
			}},
			isPoolPrepared: false,
			expectErr:      true,
		},
		{
			name: "Adding two of the same gamm pool id at once should error",
			testingMigrationRecords: []types.GammToConcentratedPoolLink{
				{
					GammPoolId: 1,
					ClPoolId:   3,
				},
				{
					GammPoolId: 1,
					ClPoolId:   3,
				},
			},
			isPoolPrepared: true,
			expectErr:      true,
		},
		{
			name: "Adding unsorted gamm pools should error",
			testingMigrationRecords: []types.GammToConcentratedPoolLink{
				{
					GammPoolId: 2,
					ClPoolId:   4,
				},
				{
					GammPoolId: 1,
					ClPoolId:   3,
				},
			},
			isPoolPrepared: true,
			expectErr:      true,
		},
		{
			name: "Normal case with two records",
			testingMigrationRecords: []types.GammToConcentratedPoolLink{
				{
					GammPoolId: 1,
					ClPoolId:   3,
				},
				{
					GammPoolId: 2,
					ClPoolId:   4,
				},
			},
			isPoolPrepared: true,
			expectErr:      false,
		},
		{
			name: "Try to set one of the GammPoolIds to a cl pool Id",
			testingMigrationRecords: []types.GammToConcentratedPoolLink{
				{
					GammPoolId: 2,
					ClPoolId:   4,
				},
				{
					GammPoolId: 3,
					ClPoolId:   1,
				},
			},
			isPoolPrepared: true,
			expectErr:      true,
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			suite.SetupTest()
			keeper := suite.App.GAMMKeeper

			if test.isPoolPrepared {
				// Our testing environment is as follows:
				// Gamm pool IDs: 1, 2
				// Concentrated pool IDs: 3, 4
				suite.PrepareBalancerPool()
				suite.PrepareBalancerPool()
				suite.PrepareConcentratedPool()
				suite.PrepareConcentratedPool()
			}

			err := keeper.ReplaceMigrationRecords(suite.Ctx, test.testingMigrationRecords...)
			if test.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)

				migrationInfo := keeper.GetMigrationInfo(suite.Ctx)
				suite.Require().Equal(len(test.testingMigrationRecords), len(migrationInfo.GammToConcentratedPoolLinks))
				for i, record := range test.testingMigrationRecords {
					suite.Require().Equal(record.GammPoolId, migrationInfo.GammToConcentratedPoolLinks[i].GammPoolId)
					suite.Require().Equal(record.ClPoolId, migrationInfo.GammToConcentratedPoolLinks[i].ClPoolId)
				}
			}
		})
	}
}

func (suite *KeeperTestSuite) TestUpdateDistrRecords() {
	tests := []struct {
		name                     string
		testingMigrationRecords  []types.GammToConcentratedPoolLink
		expectedResultingRecords []types.GammToConcentratedPoolLink
		isPoolPrepared           bool
		isPrexistingRecordsSet   bool
		expectErr                bool
	}{
		{
			name: "Non existent pool.",
			testingMigrationRecords: []types.GammToConcentratedPoolLink{{
				GammPoolId: 1,
				ClPoolId:   6,
			}},
			isPoolPrepared:         false,
			isPrexistingRecordsSet: false,
			expectErr:              true,
		},
		{
			name: "Adding two of the same gamm pool id at once should error",
			testingMigrationRecords: []types.GammToConcentratedPoolLink{
				{
					GammPoolId: 1,
					ClPoolId:   6,
				},
				{
					GammPoolId: 1,
					ClPoolId:   6,
				},
			},
			isPoolPrepared:         true,
			isPrexistingRecordsSet: true,
			expectErr:              true,
		},
		{
			name: "Adding unsorted gamm pools should error",
			testingMigrationRecords: []types.GammToConcentratedPoolLink{
				{
					GammPoolId: 2,
					ClPoolId:   7,
				},
				{
					GammPoolId: 1,
					ClPoolId:   6,
				},
			},
			isPoolPrepared:         true,
			isPrexistingRecordsSet: true,
			expectErr:              true,
		},
		{
			name: "Normal case with two records",
			testingMigrationRecords: []types.GammToConcentratedPoolLink{
				{
					GammPoolId: 1,
					ClPoolId:   6,
				},
				{
					GammPoolId: 2,
					ClPoolId:   8,
				},
			},
			expectedResultingRecords: []types.GammToConcentratedPoolLink{
				{
					GammPoolId: 1,
					ClPoolId:   6,
				},
				{
					GammPoolId: 2,
					ClPoolId:   8,
				},
				{
					GammPoolId: 3,
					ClPoolId:   7,
				},
			},
			isPoolPrepared:         true,
			isPrexistingRecordsSet: true,
			expectErr:              false,
		},
		{
			name: "Modify existing record, delete existing record, leave a record alone, add new record",
			testingMigrationRecords: []types.GammToConcentratedPoolLink{
				{
					GammPoolId: 1,
					ClPoolId:   6,
				},
				{
					GammPoolId: 2,
					ClPoolId:   0,
				},
				{
					GammPoolId: 4,
					ClPoolId:   8,
				},
			},
			expectedResultingRecords: []types.GammToConcentratedPoolLink{
				{
					GammPoolId: 1,
					ClPoolId:   6,
				},
				{
					GammPoolId: 3,
					ClPoolId:   7,
				},
				{
					GammPoolId: 4,
					ClPoolId:   8,
				},
			},
			isPoolPrepared:         true,
			isPrexistingRecordsSet: true,
			expectErr:              false,
		},
		{
			name: "Try to set one of the GammPoolIds to a cl pool Id",
			testingMigrationRecords: []types.GammToConcentratedPoolLink{
				{
					GammPoolId: 2,
					ClPoolId:   4,
				},
				{
					GammPoolId: 5,
					ClPoolId:   6,
				},
			},
			isPoolPrepared:         true,
			isPrexistingRecordsSet: true,
			expectErr:              true,
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			suite.SetupTest()
			keeper := suite.App.GAMMKeeper

			if test.isPoolPrepared {
				// Our testing environment is as follows:
				// Gamm pool IDs: 1, 2, 3, 4
				// Concentrated pool IDs: 5, 6, 7, 8
				suite.PrepareBalancerPool()
				suite.PrepareBalancerPool()
				suite.PrepareBalancerPool()
				suite.PrepareBalancerPool()
				suite.PrepareConcentratedPool()
				suite.PrepareConcentratedPool()
				suite.PrepareConcentratedPool()
				suite.PrepareConcentratedPool()
			}

			if test.isPrexistingRecordsSet {
				// Set up existing records so we can update them
				existingRecords := []types.GammToConcentratedPoolLink{
					{
						GammPoolId: 1,
						ClPoolId:   5,
					},
					{
						GammPoolId: 2,
						ClPoolId:   6,
					},
					{
						GammPoolId: 3,
						ClPoolId:   7,
					},
				}
				err := keeper.ReplaceMigrationRecords(suite.Ctx, existingRecords...)
				suite.Require().NoError(err)
			}

			err := keeper.UpdateMigrationRecords(suite.Ctx, test.testingMigrationRecords...)
			if test.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)

				migrationInfo := keeper.GetMigrationInfo(suite.Ctx)
				suite.Require().Equal(len(test.expectedResultingRecords), len(migrationInfo.GammToConcentratedPoolLinks))
				for i, record := range test.expectedResultingRecords {
					suite.Require().Equal(record.GammPoolId, migrationInfo.GammToConcentratedPoolLinks[i].GammPoolId)
					suite.Require().Equal(record.ClPoolId, migrationInfo.GammToConcentratedPoolLinks[i].ClPoolId)
				}
			}
		})
	}
}
