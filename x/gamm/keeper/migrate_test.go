package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/types"
)

func (suite *KeeperTestSuite) TestMigrate() {
	defaultAccount := suite.TestAccs[0]
	defaultGammShares := sdk.NewCoin("gamm/pool/1", sdk.MustNewDecFromStr("100000000000000000000").RoundInt())
	invalidGammShares := sdk.NewCoin("gamm/pool/1", sdk.MustNewDecFromStr("190000000000000000001").RoundInt())
	defaultAccountFunds := sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(200000000000)), sdk.NewCoin("usdc", sdk.NewInt(200000000000)))
	defaultErrorTolerance := osmomath.ErrTolerance{
		AdditiveTolerance: sdk.NewDec(100),
		RoundingDir:       osmomath.RoundDown,
	}
	defaultJoinTime := suite.Ctx.BlockTime()

	type param struct {
		sender                sdk.AccAddress
		sharesToMigrateDenom  string
		sharesToMigrateAmount sdk.Int
	}

	tests := []struct {
		name                   string
		param                  param
		expectedErr            error
		sharesToCreate         sdk.Int
		expectedLiquidity      sdk.Dec
		setupPoolMigrationLink bool
		errTolerance           osmomath.ErrTolerance
	}{
		{
			name: "migrate all of the shares (with pool migration link)",
			param: param{
				sender:                defaultAccount,
				sharesToMigrateDenom:  defaultGammShares.Denom,
				sharesToMigrateAmount: defaultGammShares.Amount,
			},
			sharesToCreate:         defaultGammShares.Amount,
			expectedLiquidity:      sdk.MustNewDecFromStr("100000000000.000000010000000000"),
			setupPoolMigrationLink: true,
			errTolerance:           defaultErrorTolerance,
		},
		{
			name: "migrate all of the shares (no pool migration link)",
			param: param{
				sender:                defaultAccount,
				sharesToMigrateDenom:  defaultGammShares.Denom,
				sharesToMigrateAmount: defaultGammShares.Amount,
			},
			sharesToCreate:         defaultGammShares.Amount,
			expectedLiquidity:      sdk.MustNewDecFromStr("100000000000.000000010000000000"),
			setupPoolMigrationLink: false,
			expectedErr:            types.ConcentratedPoolMigrationLinkNotFoundError{PoolIdLeaving: 1},
			errTolerance:           defaultErrorTolerance,
		},
		{
			name: "migrate half of the shares",
			param: param{
				sender:                defaultAccount,
				sharesToMigrateDenom:  defaultGammShares.Denom,
				sharesToMigrateAmount: defaultGammShares.Amount.Quo(sdk.NewInt(2)),
			},
			sharesToCreate:         defaultGammShares.Amount,
			expectedLiquidity:      sdk.MustNewDecFromStr("50000000000.000000005000000000"),
			setupPoolMigrationLink: true,
			errTolerance:           defaultErrorTolerance,
		},
		{
			name: "double the created shares, migrate 1/4 of the shares",
			param: param{
				sender:                defaultAccount,
				sharesToMigrateDenom:  defaultGammShares.Denom,
				sharesToMigrateAmount: defaultGammShares.Amount.Quo(sdk.NewInt(2)),
			},
			sharesToCreate:         defaultGammShares.Amount.Mul(sdk.NewInt(2)),
			expectedLiquidity:      sdk.MustNewDecFromStr("49999999999.000000004999999999"),
			setupPoolMigrationLink: true,
			errTolerance:           defaultErrorTolerance,
		},
		{
			name: "error: attempt to migrate more shares than the user has",
			param: param{
				sender:                defaultAccount,
				sharesToMigrateDenom:  defaultGammShares.Denom,
				sharesToMigrateAmount: invalidGammShares.Amount,
			},
			sharesToCreate:         defaultGammShares.Amount,
			expectedLiquidity:      sdk.MustNewDecFromStr("100000000000.000000010000000000"),
			setupPoolMigrationLink: true,
			expectedErr:            sdkerrors.Wrap(sdkerrors.ErrInsufficientFunds, fmt.Sprintf("%s is smaller than %s", defaultGammShares, invalidGammShares)),
		},
	}

	for _, test := range tests {
		suite.SetupTest()
		suite.Ctx = suite.Ctx.WithBlockTime(defaultJoinTime)
		keeper := suite.App.GAMMKeeper

		// Prepare both balancer and concentrated pools
		suite.FundAcc(test.param.sender, defaultAccountFunds)
		balancerPoolId := suite.PrepareBalancerPoolWithCoins(sdk.NewCoin("eth", sdk.NewInt(100000000000)), sdk.NewCoin("usdc", sdk.NewInt(100000000000)))
		balancerPool, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, balancerPoolId)
		suite.Require().NoError(err)
		clPool := suite.PrepareConcentratedPool()

		// Set up canonical link between balancer and cl pool
		if test.setupPoolMigrationLink {
			record := types.BalancerToConcentratedPoolLink{BalancerPool: types.PoolID{PoolId: balancerPoolId}, ClPool: types.PoolID{PoolId: clPool.GetId()}}
			err = keeper.ReplaceMigrationRecords(suite.Ctx, []types.BalancerToConcentratedPoolLink{record})
			suite.Require().NoError(err)
		}

		// Note gamm and cl pool addresses
		balancerPoolAddress := balancerPool.GetAddress()
		clPoolAddress := clPool.GetAddress()

		// Join balancer pool to create gamm shares directed in the test case
		_, _, err = suite.App.GAMMKeeper.JoinPoolNoSwap(suite.Ctx, test.param.sender, balancerPoolId, test.sharesToCreate, sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(999999999999999)), sdk.NewCoin("usdc", sdk.NewInt(999999999999999))))
		suite.Require().NoError(err)

		// Note balancer pool balance after joining balancer pool
		gammPoolEthBalancePostJoin := suite.App.BankKeeper.GetBalance(suite.Ctx, balancerPoolAddress, ETH)
		gammPoolUsdcBalancePostJoin := suite.App.BankKeeper.GetBalance(suite.Ctx, balancerPoolAddress, USDC)

		// Note users gamm share balance after joining balancer pool
		userGammBalancePostJoin := suite.App.BankKeeper.GetBalance(suite.Ctx, test.param.sender, "gamm/pool/1")

		// Create migrate message
		balancerPool, err = suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, balancerPoolId)
		suite.Require().NoError(err)
		sharesToMigrate := sdk.NewCoin(test.param.sharesToMigrateDenom, test.param.sharesToMigrateAmount)
		expectedCoinsOut, err := balancerPool.CalcExitPoolCoinsFromShares(suite.Ctx, sharesToMigrate.Amount, sdk.ZeroDec())
		suite.Require().NoError(err)

		// Migrate the user's gamm shares to a full range concentrated liquidity position
		userBalancesBeforeMigration := suite.App.BankKeeper.GetAllBalances(suite.Ctx, test.param.sender)
		positionId, amount0, amount1, _, _, poolIdLeaving, poolIdEntering, err := keeper.MigrateFromBalancerToConcentrated(suite.Ctx, test.param.sender, sharesToMigrate)
		userBalancesAfterMigration := suite.App.BankKeeper.GetAllBalances(suite.Ctx, test.param.sender)
		if test.expectedErr != nil {
			suite.Require().Error(err)
			suite.Require().ErrorContains(err, test.expectedErr.Error())

			// Expect zero values for both pool ids
			suite.Require().Zero(poolIdLeaving)
			suite.Require().Zero(poolIdEntering)

			// Assure the user's gamm shares still exist
			userGammBalanceAfterFailedMigration := suite.App.BankKeeper.GetBalance(suite.Ctx, test.param.sender, "gamm/pool/1")
			suite.Require().Equal(userGammBalancePostJoin.String(), userGammBalanceAfterFailedMigration.String())

			// Assure cl pool has no balance after a failed migration.
			clPoolEthBalanceAfterFailedMigration := suite.App.BankKeeper.GetBalance(suite.Ctx, clPoolAddress, ETH)
			clPoolUsdcBalanceAfterFailedMigration := suite.App.BankKeeper.GetBalance(suite.Ctx, clPoolAddress, USDC)
			suite.Require().Equal(sdk.NewInt(0), clPoolEthBalanceAfterFailedMigration.Amount)
			suite.Require().Equal(sdk.NewInt(0), clPoolUsdcBalanceAfterFailedMigration.Amount)

			// Assure the position was not created.
			// TODO: When we implement lock breaking, we need to change time.Time{} to the lock's end time.
			_, err := suite.App.ConcentratedLiquidityKeeper.GetPositionLiquidity(suite.Ctx, positionId)
			suite.Require().Error(err)
			continue
		}
		suite.Require().NoError(err)

		// Expect the poolIdLeaving to be the balancer pool id
		// Expect the poolIdEntering to be the concentrated liquidity pool id
		suite.Require().Equal(balancerPoolId, poolIdLeaving)
		suite.Require().Equal(clPool.GetId(), poolIdEntering)

		// Determine how much of the user's balance was not used in the migration
		// This amount should be returned to the user.
		expectedUserFinalEthBalanceDiff := expectedCoinsOut.AmountOf(ETH).Sub(amount0)
		expectedUserFinalUsdcBalanceDiff := expectedCoinsOut.AmountOf(USDC).Sub(amount1)
		suite.Require().Equal(userBalancesBeforeMigration.AmountOf(ETH).Add(expectedUserFinalEthBalanceDiff).String(), userBalancesAfterMigration.AmountOf(ETH).String())
		suite.Require().Equal(userBalancesBeforeMigration.AmountOf(USDC).Add(expectedUserFinalUsdcBalanceDiff).String(), userBalancesAfterMigration.AmountOf(USDC).String())

		// Assure the expected position was created.
		// TODO: When we implement lock breaking, we need to change time.Time{} to the lock's end time.
		position, err := suite.App.ConcentratedLiquidityKeeper.GetPositionLiquidity(suite.Ctx, positionId)
		suite.Require().NoError(err)
		suite.Require().Equal(test.expectedLiquidity, position)

		// Note gamm pool balance after migration
		gammPoolEthBalancePostMigrate := suite.App.BankKeeper.GetBalance(suite.Ctx, balancerPoolAddress, ETH)
		gammPoolUsdcBalancePostMigrate := suite.App.BankKeeper.GetBalance(suite.Ctx, balancerPoolAddress, USDC)

		// Note user amount transferred to cl pool from balancer pool
		userEthBalanceTransferredToClPool := gammPoolEthBalancePostJoin.Sub(gammPoolEthBalancePostMigrate)
		userUsdcBalanceTransferredToClPool := gammPoolUsdcBalancePostJoin.Sub(gammPoolUsdcBalancePostMigrate)

		// Note cl pool balance after migration
		clPoolEthBalanceAfterMigration := suite.App.BankKeeper.GetBalance(suite.Ctx, clPoolAddress, ETH)
		clPoolUsdcBalanceAfterMigration := suite.App.BankKeeper.GetBalance(suite.Ctx, clPoolAddress, USDC)

		// The balance in the cl pool should be equal to what the user previously had in the gamm pool.
		// This test is within 100 shares due to rounding that occurs from utilizing .000000000000000001 instead of 0.
		suite.Require().Equal(0, test.errTolerance.Compare(userEthBalanceTransferredToClPool.Amount, clPoolEthBalanceAfterMigration.Amount))
		suite.Require().Equal(0, test.errTolerance.Compare(userUsdcBalanceTransferredToClPool.Amount, clPoolUsdcBalanceAfterMigration.Amount))

		// Assert user amount transferred to cl pool from gamm pool should be equal to the amount we migrated from the migrate message.
		// This test is within 100 shares due to rounding that occurs from utilizing .000000000000000001 instead of 0.
		suite.Require().Equal(0, test.errTolerance.Compare(userEthBalanceTransferredToClPool.Amount, amount0))
		suite.Require().Equal(0, test.errTolerance.Compare(userUsdcBalanceTransferredToClPool.Amount, amount1))
	}
}

func (suite *KeeperTestSuite) TestReplaceMigrationRecords() {
	tests := []struct {
		name                        string
		testingMigrationRecords     []types.BalancerToConcentratedPoolLink
		overwriteBalancerDenom0     string
		overwriteBalancerDenom1     string
		createFourAssetBalancerPool bool
		expectErr                   bool
	}{
		{
			name: "Non existent balancer pool",
			testingMigrationRecords: []types.BalancerToConcentratedPoolLink{{
				BalancerPool: types.PoolID{PoolId: 5},
				ClPool:       types.PoolID{PoolId: 3},
			}},
			expectErr: true,
		},
		{
			name: "Non existent concentrated pool",
			testingMigrationRecords: []types.BalancerToConcentratedPoolLink{{
				BalancerPool: types.PoolID{PoolId: 1},
				ClPool:       types.PoolID{PoolId: 5},
			}},
			expectErr: true,
		},
		{
			name: "Adding two of the same balancer pool id at once should error",
			testingMigrationRecords: []types.BalancerToConcentratedPoolLink{

				{
					BalancerPool: types.PoolID{PoolId: 1},
					ClPool:       types.PoolID{PoolId: 3},
				},
				{
					BalancerPool: types.PoolID{PoolId: 1},
					ClPool:       types.PoolID{PoolId: 4},
				},
			},
			expectErr: true,
		},
		{
			name: "Adding two of the same cl pool id at once should error",
			testingMigrationRecords: []types.BalancerToConcentratedPoolLink{
				{
					BalancerPool: types.PoolID{PoolId: 1},
					ClPool:       types.PoolID{PoolId: 3},
				},
				{
					BalancerPool: types.PoolID{PoolId: 2},
					ClPool:       types.PoolID{PoolId: 3},
				},
			},
			expectErr: true,
		},
		{
			name: "Normal case with two records",
			testingMigrationRecords: []types.BalancerToConcentratedPoolLink{
				{
					BalancerPool: types.PoolID{PoolId: 1},
					ClPool:       types.PoolID{PoolId: 3},
				},
				{
					BalancerPool: types.PoolID{PoolId: 2},
					ClPool:       types.PoolID{PoolId: 4},
				},
			},
			expectErr: false,
		},
		{
			name: "Try to set one of the BalancerPoolIds to a cl pool Id",
			testingMigrationRecords: []types.BalancerToConcentratedPoolLink{
				{
					BalancerPool: types.PoolID{PoolId: 2},
					ClPool:       types.PoolID{PoolId: 4},
				},
				{
					BalancerPool: types.PoolID{PoolId: 3},
					ClPool:       types.PoolID{PoolId: 1},
				},
			},
			expectErr: true,
		},
		{
			name: "Try to set one of the ClPoolIds to a balancer pool Id",
			testingMigrationRecords: []types.BalancerToConcentratedPoolLink{
				{
					BalancerPool: types.PoolID{PoolId: 2},
					ClPool:       types.PoolID{PoolId: 1},
				},
			},
			expectErr: true,
		},
		{
			name: "Mismatch denom0 between the two pools",
			testingMigrationRecords: []types.BalancerToConcentratedPoolLink{
				{
					BalancerPool: types.PoolID{PoolId: 1},
					ClPool:       types.PoolID{PoolId: 3},
				},
			},
			overwriteBalancerDenom0: "uosmo",
			expectErr:               true,
		},
		{
			name: "Mismatch denom1 between the two pools",
			testingMigrationRecords: []types.BalancerToConcentratedPoolLink{
				{
					BalancerPool: types.PoolID{PoolId: 1},
					ClPool:       types.PoolID{PoolId: 3},
				},
			},
			overwriteBalancerDenom1: "uosmo",
			expectErr:               true,
		},
		{
			name: "Balancer pool has more than two tokens",
			testingMigrationRecords: []types.BalancerToConcentratedPoolLink{
				{
					BalancerPool: types.PoolID{PoolId: 5},
					ClPool:       types.PoolID{PoolId: 3},
				},
			},
			createFourAssetBalancerPool: true,
			expectErr:                   true,
		},
	}

	for _, test := range tests {
		test := test
		suite.Run(test.name, func() {
			suite.SetupTest()
			keeper := suite.App.GAMMKeeper

			defaultBalancerCoin0 := sdk.NewCoin(ETH, sdk.NewInt(1000000000))
			defaultBalancerCoin1 := sdk.NewCoin(USDC, sdk.NewInt(1000000000))

			if test.overwriteBalancerDenom0 != "" {
				defaultBalancerCoin0.Denom = test.overwriteBalancerDenom0
			}
			if test.overwriteBalancerDenom1 != "" {
				defaultBalancerCoin1.Denom = test.overwriteBalancerDenom1
			}

			// Our testing environment is as follows:
			// Balancer pool IDs: 1, 2
			// Concentrated pool IDs: 3, 4
			for i := 0; i < 2; i++ {
				poolCoins := sdk.NewCoins(defaultBalancerCoin0, defaultBalancerCoin1)
				suite.PrepareBalancerPoolWithCoins(poolCoins...)
			}
			for i := 0; i < 2; i++ {
				suite.PrepareCustomConcentratedPool(suite.TestAccs[0], ETH, USDC, defaultTickSpacing, DefaultExponentAtPriceOne, sdk.ZeroDec())
			}
			// Four asset balancer pool ID if created: 5
			if test.createFourAssetBalancerPool {
				suite.PrepareBalancerPool()
			}

			err := keeper.ReplaceMigrationRecords(suite.Ctx, test.testingMigrationRecords)
			if test.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)

				migrationInfo, err := keeper.GetAllMigrationInfo(suite.Ctx)
				suite.Require().NoError(err)
				suite.Require().Equal(len(test.testingMigrationRecords), len(migrationInfo.BalancerToConcentratedPoolLinks))
				for i, record := range test.testingMigrationRecords {
					suite.Require().Equal(record.BalancerPool.PoolId, migrationInfo.BalancerToConcentratedPoolLinks[i].BalancerPool.PoolId)
					suite.Require().Equal(record.ClPool.PoolId, migrationInfo.BalancerToConcentratedPoolLinks[i].ClPool.PoolId)
				}
			}
		})
	}
}

func (suite *KeeperTestSuite) TestUpdateMigrationRecords() {
	tests := []struct {
		name                        string
		testingMigrationRecords     []types.BalancerToConcentratedPoolLink
		expectedResultingRecords    []types.BalancerToConcentratedPoolLink
		isPoolPrepared              bool
		isPreexistingRecordsSet     bool
		overwriteBalancerDenom0     string
		overwriteBalancerDenom1     string
		createFourAssetBalancerPool bool
		expectErr                   bool
	}{
		{
			name: "Non existent balancer pool.",
			testingMigrationRecords: []types.BalancerToConcentratedPoolLink{{
				BalancerPool: types.PoolID{PoolId: 9},
				ClPool:       types.PoolID{PoolId: 6},
			}},
			isPreexistingRecordsSet: false,
			expectErr:               true,
		},
		{
			name: "Non existent concentrated pool.",
			testingMigrationRecords: []types.BalancerToConcentratedPoolLink{{
				BalancerPool: types.PoolID{PoolId: 1},
				ClPool:       types.PoolID{PoolId: 9},
			}},
			isPreexistingRecordsSet: false,
			expectErr:               true,
		},
		{
			name: "Adding two of the same balancer pool ids at once should error",
			testingMigrationRecords: []types.BalancerToConcentratedPoolLink{
				{
					BalancerPool: types.PoolID{PoolId: 1},
					ClPool:       types.PoolID{PoolId: 6},
				},
				{
					BalancerPool: types.PoolID{PoolId: 1},
					ClPool:       types.PoolID{PoolId: 7},
				},
			},
			isPreexistingRecordsSet: true,
			expectErr:               true,
		},
		{
			name: "Adding two of the same cl pool ids at once should error",
			testingMigrationRecords: []types.BalancerToConcentratedPoolLink{
				{
					BalancerPool: types.PoolID{PoolId: 1},
					ClPool:       types.PoolID{PoolId: 6},
				},
				{
					BalancerPool: types.PoolID{PoolId: 2},
					ClPool:       types.PoolID{PoolId: 6},
				},
			},
			isPreexistingRecordsSet: true,
			expectErr:               true,
		},
		{
			name: "Normal case with two records",
			testingMigrationRecords: []types.BalancerToConcentratedPoolLink{
				{
					BalancerPool: types.PoolID{PoolId: 1},
					ClPool:       types.PoolID{PoolId: 6},
				},
				{
					BalancerPool: types.PoolID{PoolId: 2},
					ClPool:       types.PoolID{PoolId: 8},
				},
			},
			expectedResultingRecords: []types.BalancerToConcentratedPoolLink{
				{
					BalancerPool: types.PoolID{PoolId: 1},
					ClPool:       types.PoolID{PoolId: 6},
				},
				{
					BalancerPool: types.PoolID{PoolId: 2},
					ClPool:       types.PoolID{PoolId: 8},
				},
				{
					BalancerPool: types.PoolID{PoolId: 3},
					ClPool:       types.PoolID{PoolId: 7},
				},
			},
			isPreexistingRecordsSet: true,
			expectErr:               false,
		},
		{
			name: "Normal case with two records no preexisting records",
			testingMigrationRecords: []types.BalancerToConcentratedPoolLink{
				{
					BalancerPool: types.PoolID{PoolId: 1},
					ClPool:       types.PoolID{PoolId: 6},
				},
				{
					BalancerPool: types.PoolID{PoolId: 2},
					ClPool:       types.PoolID{PoolId: 8},
				},
			},
			expectedResultingRecords: []types.BalancerToConcentratedPoolLink{
				{
					BalancerPool: types.PoolID{PoolId: 1},
					ClPool:       types.PoolID{PoolId: 6},
				},
				{
					BalancerPool: types.PoolID{PoolId: 2},
					ClPool:       types.PoolID{PoolId: 8},
				},
			},
			isPreexistingRecordsSet: false,
			expectErr:               false,
		},
		{
			name: "Modify existing record, delete existing record, leave a record alone, add new record",
			testingMigrationRecords: []types.BalancerToConcentratedPoolLink{
				{
					BalancerPool: types.PoolID{PoolId: 1},
					ClPool:       types.PoolID{PoolId: 6},
				},
				{
					BalancerPool: types.PoolID{PoolId: 2},
					ClPool:       types.PoolID{PoolId: 0},
				},
				{
					BalancerPool: types.PoolID{PoolId: 4},
					ClPool:       types.PoolID{PoolId: 8},
				},
			},
			expectedResultingRecords: []types.BalancerToConcentratedPoolLink{
				{
					BalancerPool: types.PoolID{PoolId: 1},
					ClPool:       types.PoolID{PoolId: 6},
				},
				{
					BalancerPool: types.PoolID{PoolId: 3},
					ClPool:       types.PoolID{PoolId: 7},
				},
				{
					BalancerPool: types.PoolID{PoolId: 4},
					ClPool:       types.PoolID{PoolId: 8},
				},
			},
			isPreexistingRecordsSet: true,
			expectErr:               false,
		},
		{
			name: "Try to set one of the BalancerPoolIds to a cl pool Id",
			testingMigrationRecords: []types.BalancerToConcentratedPoolLink{
				{
					BalancerPool: types.PoolID{PoolId: 2},
					ClPool:       types.PoolID{PoolId: 4},
				},
				{
					BalancerPool: types.PoolID{PoolId: 5},
					ClPool:       types.PoolID{PoolId: 6},
				},
			},
			isPreexistingRecordsSet: true,
			expectErr:               true,
		},
		{
			name: "Try to set one of the ClPoolIds to a balancer pool Id",
			testingMigrationRecords: []types.BalancerToConcentratedPoolLink{
				{
					BalancerPool: types.PoolID{PoolId: 2},
					ClPool:       types.PoolID{PoolId: 1},
				},
			},
			isPreexistingRecordsSet: true,
			expectErr:               true,
		},
		{
			name: "Mismatch denom0 between the two pools",
			testingMigrationRecords: []types.BalancerToConcentratedPoolLink{
				{
					BalancerPool: types.PoolID{PoolId: 1},
					ClPool:       types.PoolID{PoolId: 6},
				},
			},
			overwriteBalancerDenom0: "osmo",
			isPreexistingRecordsSet: false,
			expectErr:               true,
		},
		{
			name: "Mismatch denom1 between the two pools",
			testingMigrationRecords: []types.BalancerToConcentratedPoolLink{
				{
					BalancerPool: types.PoolID{PoolId: 1},
					ClPool:       types.PoolID{PoolId: 6},
				},
			},
			overwriteBalancerDenom1: "osmo",
			isPreexistingRecordsSet: false,
			expectErr:               true,
		},
		{
			name: "Balancer pool has more than two tokens",
			testingMigrationRecords: []types.BalancerToConcentratedPoolLink{
				{
					BalancerPool: types.PoolID{PoolId: 9},
					ClPool:       types.PoolID{PoolId: 6},
				},
			},
			isPreexistingRecordsSet:     false,
			createFourAssetBalancerPool: true,
			expectErr:                   true,
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			suite.SetupTest()
			keeper := suite.App.GAMMKeeper

			defaultBalancerCoin0 := sdk.NewCoin(ETH, sdk.NewInt(1000000000))
			defaultBalancerCoin1 := sdk.NewCoin(USDC, sdk.NewInt(1000000000))

			if test.overwriteBalancerDenom0 != "" {
				defaultBalancerCoin0.Denom = test.overwriteBalancerDenom0
			}
			if test.overwriteBalancerDenom1 != "" {
				defaultBalancerCoin1.Denom = test.overwriteBalancerDenom1
			}

			// Our testing environment is as follows:
			// Balancer pool IDs: 1, 2, 3, 4
			// Concentrated pool IDs: 5, 6, 7, 8
			for i := 0; i < 4; i++ {
				poolCoins := sdk.NewCoins(defaultBalancerCoin0, defaultBalancerCoin1)
				suite.PrepareBalancerPoolWithCoins(poolCoins...)
			}
			for i := 0; i < 4; i++ {
				suite.PrepareCustomConcentratedPool(suite.TestAccs[0], ETH, USDC, defaultTickSpacing, DefaultExponentAtPriceOne, sdk.ZeroDec())
			}
			// Four asset balancer pool ID if created: 9
			if test.createFourAssetBalancerPool {
				suite.PrepareBalancerPool()
			}

			if test.isPreexistingRecordsSet {
				// Set up existing records so we can update them
				existingRecords := []types.BalancerToConcentratedPoolLink{
					{
						BalancerPool: types.PoolID{PoolId: 1},
						ClPool:       types.PoolID{PoolId: 5},
					},
					{
						BalancerPool: types.PoolID{PoolId: 2},
						ClPool:       types.PoolID{PoolId: 6},
					},
					{
						BalancerPool: types.PoolID{PoolId: 3},
						ClPool:       types.PoolID{PoolId: 7},
					},
				}
				err := keeper.ReplaceMigrationRecords(suite.Ctx, existingRecords)
				suite.Require().NoError(err)
			}

			err := keeper.UpdateMigrationRecords(suite.Ctx, test.testingMigrationRecords)
			if test.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)

				migrationInfo, err := keeper.GetAllMigrationInfo(suite.Ctx)
				suite.Require().NoError(err)
				fmt.Println(test.expectedResultingRecords)
				fmt.Println(migrationInfo.BalancerToConcentratedPoolLinks)
				suite.Require().Equal(len(test.expectedResultingRecords), len(migrationInfo.BalancerToConcentratedPoolLinks))
				for i, record := range test.expectedResultingRecords {
					suite.Require().Equal(record.BalancerPool.PoolId, migrationInfo.BalancerToConcentratedPoolLinks[i].BalancerPool.PoolId)
					suite.Require().Equal(record.ClPool.PoolId, migrationInfo.BalancerToConcentratedPoolLinks[i].ClPool.PoolId)
				}
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGetLinkedConcentratedPoolID() {
	tests := []struct {
		name                   string
		poolIdLeaving          []uint64
		expectedPoolIdEntering []uint64
		expectErr              bool
	}{
		{
			name:                   "Happy path",
			poolIdLeaving:          []uint64{1, 2, 3},
			expectedPoolIdEntering: []uint64{4, 5, 6},
			expectErr:              false,
		},
		{
			name:          "error: set poolIdLeaving to a concentrated pool ID",
			poolIdLeaving: []uint64{4},
			expectErr:     true,
		},
		{
			name:          "error: set poolIdLeaving to a non existent pool ID",
			poolIdLeaving: []uint64{7},
			expectErr:     true,
		},
	}

	for _, test := range tests {
		test := test
		suite.Run(test.name, func() {
			suite.SetupTest()
			keeper := suite.App.GAMMKeeper

			// Our testing environment is as follows:
			// Balancer pool IDs: 1, 2, 3
			// Concentrated pool IDs: 3, 4, 5
			suite.PrepareMultipleBalancerPools(3)
			suite.PrepareMultipleConcentratedPools(3)

			keeper.SetMigrationInfo(suite.Ctx, DefaultMigrationRecords)

			for i, poolIdLeaving := range test.poolIdLeaving {
				poolIdEntering, err := keeper.GetLinkedConcentratedPoolID(suite.Ctx, poolIdLeaving)
				if test.expectErr {
					suite.Require().Error(err)
					suite.Require().Zero(poolIdEntering)
				} else {
					suite.Require().NoError(err)
					suite.Require().Equal(test.expectedPoolIdEntering[i], poolIdEntering)
				}
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGetLinkedBalancerPoolID() {
	tests := []struct {
		name                  string
		poolIdEntering        []uint64
		expectedPoolIdLeaving []uint64

		skipLinking bool
		expectErr   bool
	}{
		{
			name:                  "Happy path",
			poolIdEntering:        []uint64{4, 5, 6},
			expectedPoolIdLeaving: []uint64{1, 2, 3},
			expectErr:             false,
		},
		{
			name:           "error: set poolIdEntering to a balancer pool ID",
			poolIdEntering: []uint64{3},
			expectErr:      true,
		},
		{
			name:           "error: set poolIdEntering to a non existent pool ID",
			poolIdEntering: []uint64{7},
			expectErr:      true,
		},
		{
			name:                  "error: pools exist but link does not",
			poolIdEntering:        []uint64{4, 5, 6},
			expectedPoolIdLeaving: []uint64{1, 2, 3},
			skipLinking:           true,
			expectErr:             true,
		},
	}

	for _, test := range tests {
		test := test
		suite.Run(test.name, func() {
			suite.SetupTest()
			keeper := suite.App.GAMMKeeper

			// Our testing environment is as follows:
			// Balancer pool IDs: 1, 2, 3
			// Concentrated pool IDs: 3, 4, 5
			suite.PrepareMultipleBalancerPools(3)
			suite.PrepareMultipleConcentratedPools(3)

			if !test.skipLinking {
				keeper.SetMigrationInfo(suite.Ctx, DefaultMigrationRecords)
			}

			suite.Require().True(len(test.poolIdEntering) > 0)
			for i, poolIdEntering := range test.poolIdEntering {
				poolIdLeaving, err := keeper.GetLinkedBalancerPoolID(suite.Ctx, poolIdEntering)
				if test.expectErr {
					suite.Require().Error(err)
					suite.Require().Zero(poolIdLeaving)
				} else {
					suite.Require().NoError(err)
					suite.Require().Equal(test.expectedPoolIdLeaving[i], poolIdLeaving)
				}
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGetAllMigrationInfo() {
	tests := []struct {
		name        string
		skipLinking bool
	}{
		{
			name: "Happy path",
		},
		{
			name:        "No record to get",
			skipLinking: true,
		},
	}

	for _, test := range tests {
		test := test
		suite.Run(test.name, func() {
			suite.SetupTest()
			keeper := suite.App.GAMMKeeper

			// Our testing environment is as follows:
			// Balancer pool IDs: 1, 2, 3
			// Concentrated pool IDs: 3, 4, 5
			suite.PrepareMultipleBalancerPools(3)
			suite.PrepareMultipleConcentratedPools(3)

			if !test.skipLinking {
				keeper.SetMigrationInfo(suite.Ctx, DefaultMigrationRecords)
			}

			migrationRecords, err := suite.App.GAMMKeeper.GetAllMigrationInfo(suite.Ctx)
			suite.Require().NoError(err)
			if !test.skipLinking {
				suite.Require().Equal(migrationRecords, DefaultMigrationRecords)
			} else {
				suite.Require().Equal(len(migrationRecords.BalancerToConcentratedPoolLinks), 0)
			}

		})
	}
}
