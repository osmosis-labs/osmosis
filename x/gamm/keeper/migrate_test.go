package keeper_test

import (
	"github.com/osmosis-labs/osmosis/v14/x/gamm/types"
)

func (suite *KeeperTestSuite) TestReplaceMigrationRecords() {
	tests := []struct {
		name                    string
		testingMigrationRecords []types.BalancerToConcentratedPoolLink
		isPoolPrepared          bool
		expectErr               bool
	}{
		{
			name: "Non existent pool.",
			testingMigrationRecords: []types.BalancerToConcentratedPoolLink{{
				BalancerPoolId: 1,
				ClPoolId:       3,
			}},
			isPoolPrepared: false,
			expectErr:      true,
		},
		{
			name: "Adding two of the same balancer pool id at once should error",
			testingMigrationRecords: []types.BalancerToConcentratedPoolLink{
				{
					BalancerPoolId: 1,
					ClPoolId:       3,
				},
				{
					BalancerPoolId: 1,
					ClPoolId:       4,
				},
			},
			isPoolPrepared: true,
			expectErr:      true,
		},
		{
			name: "Adding two of the same cl pool id at once should error",
			testingMigrationRecords: []types.BalancerToConcentratedPoolLink{
				{
					BalancerPoolId: 1,
					ClPoolId:       3,
				},
				{
					BalancerPoolId: 2,
					ClPoolId:       3,
				},
			},
			isPoolPrepared: true,
			expectErr:      true,
		},
		{
			name: "Adding unsorted balancer pools should error",
			testingMigrationRecords: []types.BalancerToConcentratedPoolLink{
				{
					BalancerPoolId: 2,
					ClPoolId:       4,
				},
				{
					BalancerPoolId: 1,
					ClPoolId:       3,
				},
			},
			isPoolPrepared: true,
			expectErr:      true,
		},
		{
			name: "Normal case with two records",
			testingMigrationRecords: []types.BalancerToConcentratedPoolLink{
				{
					BalancerPoolId: 1,
					ClPoolId:       3,
				},
				{
					BalancerPoolId: 2,
					ClPoolId:       4,
				},
			},
			isPoolPrepared: true,
			expectErr:      false,
		},
		{
			name: "Try to set one of the BalancerPoolIds to a cl pool Id",
			testingMigrationRecords: []types.BalancerToConcentratedPoolLink{
				{
					BalancerPoolId: 2,
					ClPoolId:       4,
				},
				{
					BalancerPoolId: 3,
					ClPoolId:       1,
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
				// Balancer pool IDs: 1, 2
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
				suite.Require().Equal(len(test.testingMigrationRecords), len(migrationInfo.BalancerToConcentratedPoolLinks))
				for i, record := range test.testingMigrationRecords {
					suite.Require().Equal(record.BalancerPoolId, migrationInfo.BalancerToConcentratedPoolLinks[i].BalancerPoolId)
					suite.Require().Equal(record.ClPoolId, migrationInfo.BalancerToConcentratedPoolLinks[i].ClPoolId)
				}
			}
		})
	}
}

func (suite *KeeperTestSuite) TestUpdateMigrationRecords() {
	tests := []struct {
		name                     string
		testingMigrationRecords  []types.BalancerToConcentratedPoolLink
		expectedResultingRecords []types.BalancerToConcentratedPoolLink
		isPoolPrepared           bool
		isPreexistingRecordsSet  bool
		expectErr                bool
	}{
		{
			name: "Non existent pool.",
			testingMigrationRecords: []types.BalancerToConcentratedPoolLink{{
				BalancerPoolId: 1,
				ClPoolId:       6,
			}},
			isPoolPrepared:          false,
			isPreexistingRecordsSet: false,
			expectErr:               true,
		},
		{
			name: "Adding two of the same balancer pool ids at once should error",
			testingMigrationRecords: []types.BalancerToConcentratedPoolLink{
				{
					BalancerPoolId: 1,
					ClPoolId:       6,
				},
				{
					BalancerPoolId: 1,
					ClPoolId:       7,
				},
			},
			isPoolPrepared:          true,
			isPreexistingRecordsSet: true,
			expectErr:               true,
		},
		{
			name: "Adding two of the same cl pool ids at once should error",
			testingMigrationRecords: []types.BalancerToConcentratedPoolLink{
				{
					BalancerPoolId: 1,
					ClPoolId:       6,
				},
				{
					BalancerPoolId: 2,
					ClPoolId:       6,
				},
			},
			isPoolPrepared:          true,
			isPreexistingRecordsSet: true,
			expectErr:               true,
		},
		{
			name: "Adding unsorted balancer pools should error",
			testingMigrationRecords: []types.BalancerToConcentratedPoolLink{
				{
					BalancerPoolId: 2,
					ClPoolId:       7,
				},
				{
					BalancerPoolId: 1,
					ClPoolId:       6,
				},
			},
			isPoolPrepared:          true,
			isPreexistingRecordsSet: true,
			expectErr:               true,
		},
		{
			name: "Normal case with two records",
			testingMigrationRecords: []types.BalancerToConcentratedPoolLink{
				{
					BalancerPoolId: 1,
					ClPoolId:       6,
				},
				{
					BalancerPoolId: 2,
					ClPoolId:       8,
				},
			},
			expectedResultingRecords: []types.BalancerToConcentratedPoolLink{
				{
					BalancerPoolId: 1,
					ClPoolId:       6,
				},
				{
					BalancerPoolId: 2,
					ClPoolId:       8,
				},
				{
					BalancerPoolId: 3,
					ClPoolId:       7,
				},
			},
			isPoolPrepared:          true,
			isPreexistingRecordsSet: true,
			expectErr:               false,
		},
		{
			name: "Modify existing record, delete existing record, leave a record alone, add new record",
			testingMigrationRecords: []types.BalancerToConcentratedPoolLink{
				{
					BalancerPoolId: 1,
					ClPoolId:       6,
				},
				{
					BalancerPoolId: 2,
					ClPoolId:       0,
				},
				{
					BalancerPoolId: 4,
					ClPoolId:       8,
				},
			},
			expectedResultingRecords: []types.BalancerToConcentratedPoolLink{
				{
					BalancerPoolId: 1,
					ClPoolId:       6,
				},
				{
					BalancerPoolId: 3,
					ClPoolId:       7,
				},
				{
					BalancerPoolId: 4,
					ClPoolId:       8,
				},
			},
			isPoolPrepared:          true,
			isPreexistingRecordsSet: true,
			expectErr:               false,
		},
		{
			name: "Try to set one of the BalancerPoolIds to a cl pool Id",
			testingMigrationRecords: []types.BalancerToConcentratedPoolLink{
				{
					BalancerPoolId: 2,
					ClPoolId:       4,
				},
				{
					BalancerPoolId: 5,
					ClPoolId:       6,
				},
			},
			isPoolPrepared:          true,
			isPreexistingRecordsSet: true,
			expectErr:               true,
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			suite.SetupTest()
			keeper := suite.App.GAMMKeeper

			if test.isPoolPrepared {
				// Our testing environment is as follows:
				// Balancer pool IDs: 1, 2, 3, 4
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

			if test.isPreexistingRecordsSet {
				// Set up existing records so we can update them
				existingRecords := []types.BalancerToConcentratedPoolLink{
					{
						BalancerPoolId: 1,
						ClPoolId:       5,
					},
					{
						BalancerPoolId: 2,
						ClPoolId:       6,
					},
					{
						BalancerPoolId: 3,
						ClPoolId:       7,
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
				suite.Require().Equal(len(test.expectedResultingRecords), len(migrationInfo.BalancerToConcentratedPoolLinks))
				for i, record := range test.expectedResultingRecords {
					suite.Require().Equal(record.BalancerPoolId, migrationInfo.BalancerToConcentratedPoolLinks[i].BalancerPoolId)
					suite.Require().Equal(record.ClPoolId, migrationInfo.BalancerToConcentratedPoolLinks[i].ClPoolId)
				}
			}
		})
	}
}
