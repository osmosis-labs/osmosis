package keeper_test

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v16/app/apptesting"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/types"
	gammmigration "github.com/osmosis-labs/osmosis/v16/x/gamm/types/migration"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v16/x/pool-incentives/types"
)

func (s *KeeperTestSuite) TestMigrate() {
	defaultAccount := apptesting.CreateRandomAccounts(1)[0]
	defaultGammShares := sdk.NewCoin("gamm/pool/1", sdk.MustNewDecFromStr("100000000000000000000").RoundInt())
	invalidGammShares := sdk.NewCoin("gamm/pool/1", sdk.MustNewDecFromStr("190000000000000000001").RoundInt())
	defaultAccountFunds := sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(200000000000)), sdk.NewCoin("usdc", sdk.NewInt(200000000000)))

	// Explanation of additive tolerance of 100000:
	//
	// The balance in the CL pool should be equal to the portion of the user's previous GAMM balances that could be
	// joined into a full range CL position. These are not exactly equivalent because GAMM pools covers prices (0, inf)
	// while CL pools cover prices (minSpotPrice, maxSpotPrice), where minSpotPrice and maxSpotPrice are close to the GAMM
	// boundaries but not exactly on them.
	//
	// # Base equations for full range asset amounts:
	// Expected amount of asset 0: (liquidity * (maxSqrtPrice - curSqrtPrice)) / (maxSqrtPrice * curSqrtPrice)
	// Expected amount of asset 1: liquidity * (curSqrtPrice - minSqrtPrice)
	//
	// # Using scripts in x/concentrated-liquidity/python/swap_test.py, we compute the following:
	// expectedAsset0 = floor((liquidity * (maxSqrtPrice - curSqrtPrice)) / (maxSqrtPrice * curSqrtPrice)) = 99999999999.000000000000000000
	// expectedAsset1 = floor(liquidity * (curSqrtPrice - minSqrtPrice)) = 99999900000.000000000000000000
	//
	// We add 1 to account for ExitPool rounding exit amount up. This is not an issue since the balance is deducted from the user regardless.
	// These leaves us with full transfer of asset 0 and a (correct) transfer of asset 1 amounting to full GAMM balance minus 100000.
	// We expect this tolerance to be sufficient as long as our test cases are on the same order of magnitude.
	defaultErrorTolerance := osmomath.ErrTolerance{
		AdditiveTolerance: sdk.NewDec(100000),
		RoundingDir:       osmomath.RoundDown,
	}
	defaultJoinTime := s.Ctx.BlockTime()

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
		tokenOutMins           sdk.Coins
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
			expectedErr:            errorsmod.Wrap(sdkerrors.ErrInsufficientFunds, fmt.Sprintf("%s is smaller than %s", defaultGammShares, invalidGammShares)),
		},
		// test token out mins
		{
			name: "token out mins does not exceed actual token out",
			param: param{
				sender:                defaultAccount,
				sharesToMigrateDenom:  defaultGammShares.Denom,
				sharesToMigrateAmount: defaultGammShares.Amount,
			},
			sharesToCreate:         defaultGammShares.Amount,
			tokenOutMins:           sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(90000000000)), sdk.NewCoin(USDC, sdk.NewInt(90000000000))),
			expectedLiquidity:      sdk.MustNewDecFromStr("100000000000.000000010000000000"),
			setupPoolMigrationLink: true,
			errTolerance:           defaultErrorTolerance,
		},
		{
			name: "token out mins exceed actual token out",
			param: param{
				sender:                defaultAccount,
				sharesToMigrateDenom:  defaultGammShares.Denom,
				sharesToMigrateAmount: defaultGammShares.Amount,
			},
			sharesToCreate:         defaultGammShares.Amount,
			tokenOutMins:           sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(110000000000)), sdk.NewCoin(USDC, sdk.NewInt(110000000000))),
			expectedLiquidity:      sdk.MustNewDecFromStr("100000000000.000000010000000000"),
			setupPoolMigrationLink: true,
			expectedErr: errorsmod.Wrapf(types.ErrLimitMinAmount,
				"Exit pool returned %s , minimum tokens out specified as %s",
				sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(100000000000)), sdk.NewCoin(USDC, sdk.NewInt(100000000000))), sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(110000000000)), sdk.NewCoin(USDC, sdk.NewInt(110000000000)))),
			errTolerance: defaultErrorTolerance,
		},
		{
			name: "one of the token out mins exceed tokens out",
			param: param{
				sender:                defaultAccount,
				sharesToMigrateDenom:  defaultGammShares.Denom,
				sharesToMigrateAmount: defaultGammShares.Amount,
			},
			sharesToCreate:         defaultGammShares.Amount,
			tokenOutMins:           sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(110000000000)), sdk.NewCoin(USDC, sdk.NewInt(100000000000))),
			expectedLiquidity:      sdk.MustNewDecFromStr("100000000000.000000010000000000"),
			setupPoolMigrationLink: true,
			expectedErr: errorsmod.Wrapf(types.ErrLimitMinAmount,
				"Exit pool returned %s , minimum tokens out specified as %s",
				sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(100000000000)), sdk.NewCoin(USDC, sdk.NewInt(100000000000))), sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(110000000000)), sdk.NewCoin(USDC, sdk.NewInt(100000000000)))),
			errTolerance: defaultErrorTolerance,
		},
	}

	for _, test := range tests {
		s.SetupTest()
		s.Ctx = s.Ctx.WithBlockTime(defaultJoinTime)
		keeper := s.App.GAMMKeeper

		// Prepare both balancer and concentrated pools
		s.FundAcc(test.param.sender, defaultAccountFunds)
		balancerPoolId := s.PrepareBalancerPoolWithCoins(sdk.NewCoin("eth", sdk.NewInt(100000000000)), sdk.NewCoin("usdc", sdk.NewInt(100000000000)))
		balancerPool, err := s.App.GAMMKeeper.GetPoolAndPoke(s.Ctx, balancerPoolId)
		s.Require().NoError(err)
		clPool := s.PrepareConcentratedPool()

		// Set up canonical link between balancer and cl pool
		if test.setupPoolMigrationLink {
			record := gammmigration.BalancerToConcentratedPoolLink{BalancerPoolId: balancerPoolId, ClPoolId: clPool.GetId()}
			err = keeper.ReplaceMigrationRecords(s.Ctx, []gammmigration.BalancerToConcentratedPoolLink{record})
			s.Require().NoError(err)
		}

		// Note gamm and cl pool addresses
		balancerPoolAddress := balancerPool.GetAddress()
		clPoolAddress := clPool.GetAddress()

		// Join balancer pool to create gamm shares directed in the test case
		_, _, err = s.App.GAMMKeeper.JoinPoolNoSwap(s.Ctx, test.param.sender, balancerPoolId, test.sharesToCreate, sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(999999999999999)), sdk.NewCoin("usdc", sdk.NewInt(999999999999999))))
		s.Require().NoError(err)

		// Note balancer pool balance after joining balancer pool
		gammPoolEthBalancePostJoin := s.App.BankKeeper.GetBalance(s.Ctx, balancerPoolAddress, ETH)
		gammPoolUsdcBalancePostJoin := s.App.BankKeeper.GetBalance(s.Ctx, balancerPoolAddress, USDC)

		// Note users gamm share balance after joining balancer pool
		userGammBalancePostJoin := s.App.BankKeeper.GetBalance(s.Ctx, test.param.sender, "gamm/pool/1")

		// Create migrate message
		balancerPool, err = s.App.GAMMKeeper.GetPoolAndPoke(s.Ctx, balancerPoolId)
		s.Require().NoError(err)
		sharesToMigrate := sdk.NewCoin(test.param.sharesToMigrateDenom, test.param.sharesToMigrateAmount)
		expectedCoinsOut, err := balancerPool.CalcExitPoolCoinsFromShares(s.Ctx, sharesToMigrate.Amount, sdk.ZeroDec())
		s.Require().NoError(err)

		// Migrate the user's gamm shares to a full range concentrated liquidity position
		userBalancesBeforeMigration := s.App.BankKeeper.GetAllBalances(s.Ctx, test.param.sender)
		positionId, amount0, amount1, _, poolIdLeaving, poolIdEntering, err := keeper.MigrateUnlockedPositionFromBalancerToConcentrated(s.Ctx, test.param.sender, sharesToMigrate, test.tokenOutMins)
		userBalancesAfterMigration := s.App.BankKeeper.GetAllBalances(s.Ctx, test.param.sender)
		if test.expectedErr != nil {
			s.Require().Error(err)
			s.Require().ErrorContains(err, test.expectedErr.Error())

			// Expect zero values for both pool ids
			s.Require().Zero(poolIdLeaving)
			s.Require().Zero(poolIdEntering)

			// Assure the user's gamm shares still exist
			userGammBalanceAfterFailedMigration := s.App.BankKeeper.GetBalance(s.Ctx, test.param.sender, "gamm/pool/1")
			s.Require().Equal(userGammBalancePostJoin.String(), userGammBalanceAfterFailedMigration.String())

			// Assure cl pool has no balance after a failed migration.
			clPoolEthBalanceAfterFailedMigration := s.App.BankKeeper.GetBalance(s.Ctx, clPoolAddress, ETH)
			clPoolUsdcBalanceAfterFailedMigration := s.App.BankKeeper.GetBalance(s.Ctx, clPoolAddress, USDC)
			s.Require().Equal(sdk.NewInt(0), clPoolEthBalanceAfterFailedMigration.Amount)
			s.Require().Equal(sdk.NewInt(0), clPoolUsdcBalanceAfterFailedMigration.Amount)

			// Assure the position was not created.
			// TODO: When we implement lock breaking, we need to change time.Time{} to the lock's end time.
			_, err := s.App.ConcentratedLiquidityKeeper.GetPositionLiquidity(s.Ctx, positionId)
			s.Require().Error(err)
			continue
		}
		s.Require().NoError(err)

		// Expect the poolIdLeaving to be the balancer pool id
		// Expect the poolIdEntering to be the concentrated liquidity pool id
		s.Require().Equal(balancerPoolId, poolIdLeaving)
		s.Require().Equal(clPool.GetId(), poolIdEntering)

		// Determine how much of the user's balance was not used in the migration
		// This amount should be returned to the user.
		expectedUserFinalEthBalanceDiff := expectedCoinsOut.AmountOf(ETH).Sub(amount0)
		expectedUserFinalUsdcBalanceDiff := expectedCoinsOut.AmountOf(USDC).Sub(amount1)
		s.Require().Equal(userBalancesBeforeMigration.AmountOf(ETH).Add(expectedUserFinalEthBalanceDiff).String(), userBalancesAfterMigration.AmountOf(ETH).String())
		s.Require().Equal(userBalancesBeforeMigration.AmountOf(USDC).Add(expectedUserFinalUsdcBalanceDiff).String(), userBalancesAfterMigration.AmountOf(USDC).String())

		// Assure the expected position was created.
		// TODO: When we implement lock breaking, we need to change time.Time{} to the lock's end time.
		position, err := s.App.ConcentratedLiquidityKeeper.GetPositionLiquidity(s.Ctx, positionId)
		s.Require().NoError(err)
		s.Require().Equal(test.expectedLiquidity, position)

		// Note gamm pool balance after migration
		gammPoolEthBalancePostMigrate := s.App.BankKeeper.GetBalance(s.Ctx, balancerPoolAddress, ETH)
		gammPoolUsdcBalancePostMigrate := s.App.BankKeeper.GetBalance(s.Ctx, balancerPoolAddress, USDC)

		// Note user amount transferred to cl pool from balancer pool
		userEthBalanceTransferredToClPool := gammPoolEthBalancePostJoin.Sub(gammPoolEthBalancePostMigrate)
		userUsdcBalanceTransferredToClPool := gammPoolUsdcBalancePostJoin.Sub(gammPoolUsdcBalancePostMigrate)

		// Note cl pool balance after migration
		clPoolEthBalanceAfterMigration := s.App.BankKeeper.GetBalance(s.Ctx, clPoolAddress, ETH)
		clPoolUsdcBalanceAfterMigration := s.App.BankKeeper.GetBalance(s.Ctx, clPoolAddress, USDC)

		// The balance in the cl pool should be equal to what the user previously had in the gamm pool.
		// This test is within 100 shares due to rounding that occurs from utilizing .000000000000000001 instead of 0.
		s.Require().Equal(0, test.errTolerance.Compare(userEthBalanceTransferredToClPool.Amount, clPoolEthBalanceAfterMigration.Amount))
		s.Require().Equal(0, test.errTolerance.Compare(userUsdcBalanceTransferredToClPool.Amount, clPoolUsdcBalanceAfterMigration.Amount))

		// Assert user amount transferred to cl pool from gamm pool should be equal to the amount we migrated from the migrate message.
		// This test is within 100 shares due to rounding that occurs from utilizing .000000000000000001 instead of 0.
		s.Require().Equal(0, test.errTolerance.Compare(userEthBalanceTransferredToClPool.Amount, amount0))
		s.Require().Equal(0, test.errTolerance.Compare(userUsdcBalanceTransferredToClPool.Amount, amount1))
	}
}

func (s *KeeperTestSuite) TestReplaceMigrationRecords() {
	tests := []struct {
		name                        string
		testingMigrationRecords     []gammmigration.BalancerToConcentratedPoolLink
		overwriteBalancerDenom0     string
		overwriteBalancerDenom1     string
		createFourAssetBalancerPool bool
		expectErr                   bool
	}{
		{
			name: "Non existent balancer pool",
			testingMigrationRecords: []gammmigration.BalancerToConcentratedPoolLink{{
				BalancerPoolId: 5,
				ClPoolId:       3,
			}},
			expectErr: true,
		},
		{
			name: "Non existent concentrated pool",
			testingMigrationRecords: []gammmigration.BalancerToConcentratedPoolLink{{
				BalancerPoolId: 1,
				ClPoolId:       5,
			}},
			expectErr: true,
		},
		{
			name: "Adding two of the same balancer pool id at once should error",
			testingMigrationRecords: []gammmigration.BalancerToConcentratedPoolLink{
				{
					BalancerPoolId: 1,
					ClPoolId:       3,
				},
				{
					BalancerPoolId: 1,
					ClPoolId:       4,
				},
			},
			expectErr: true,
		},
		{
			name: "Adding two of the same cl pool id at once should error",
			testingMigrationRecords: []gammmigration.BalancerToConcentratedPoolLink{
				{
					BalancerPoolId: 1,
					ClPoolId:       3,
				},
				{
					BalancerPoolId: 2,
					ClPoolId:       3,
				},
			},
			expectErr: true,
		},
		{
			name: "Normal case with two records",
			testingMigrationRecords: []gammmigration.BalancerToConcentratedPoolLink{
				{
					BalancerPoolId: 1,
					ClPoolId:       3,
				},
				{
					BalancerPoolId: 2,
					ClPoolId:       4,
				},
			},
			expectErr: false,
		},
		{
			name: "Try to set one of the BalancerPoolIds to a cl pool Id",
			testingMigrationRecords: []gammmigration.BalancerToConcentratedPoolLink{
				{
					BalancerPoolId: 2,
					ClPoolId:       4,
				},
				{
					BalancerPoolId: 3,
					ClPoolId:       1,
				},
			},
			expectErr: true,
		},
		{
			name: "Try to set one of the ClPoolIds to a balancer pool Id",
			testingMigrationRecords: []gammmigration.BalancerToConcentratedPoolLink{
				{
					BalancerPoolId: 2,
					ClPoolId:       1,
				},
			},
			expectErr: true,
		},
		{
			name: "Mismatch denom0 between the two pools",
			testingMigrationRecords: []gammmigration.BalancerToConcentratedPoolLink{
				{
					BalancerPoolId: 1,
					ClPoolId:       3,
				},
			},
			overwriteBalancerDenom0: "uosmo",
			expectErr:               true,
		},
		{
			name: "Mismatch denom1 between the two pools",
			testingMigrationRecords: []gammmigration.BalancerToConcentratedPoolLink{
				{
					BalancerPoolId: 1,
					ClPoolId:       3,
				},
			},
			overwriteBalancerDenom1: "uosmo",
			expectErr:               true,
		},
		{
			name: "Balancer pool has more than two tokens",
			testingMigrationRecords: []gammmigration.BalancerToConcentratedPoolLink{
				{
					BalancerPoolId: 5,
					ClPoolId:       3,
				},
			},
			createFourAssetBalancerPool: true,
			expectErr:                   true,
		},
	}

	for _, test := range tests {
		test := test
		s.Run(test.name, func() {
			s.SetupTest()
			keeper := s.App.GAMMKeeper

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
				s.PrepareBalancerPoolWithCoins(poolCoins...)
			}
			for i := 0; i < 2; i++ {
				s.PrepareCustomConcentratedPool(s.TestAccs[0], ETH, USDC, defaultTickSpacing, sdk.ZeroDec())
			}
			// Four asset balancer pool ID if created: 5
			if test.createFourAssetBalancerPool {
				s.PrepareBalancerPool()
			}

			err := keeper.ReplaceMigrationRecords(s.Ctx, test.testingMigrationRecords)
			if test.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				migrationInfo, err := keeper.GetAllMigrationInfo(s.Ctx)
				s.Require().NoError(err)
				s.Require().Equal(len(test.testingMigrationRecords), len(migrationInfo.BalancerToConcentratedPoolLinks))
				for i, record := range test.testingMigrationRecords {
					s.Require().Equal(record.BalancerPoolId, migrationInfo.BalancerToConcentratedPoolLinks[i].BalancerPoolId)
					s.Require().Equal(record.ClPoolId, migrationInfo.BalancerToConcentratedPoolLinks[i].ClPoolId)
				}
			}
		})
	}
}

func (s *KeeperTestSuite) TestUpdateMigrationRecords() {
	tests := []struct {
		name                        string
		testingMigrationRecords     []gammmigration.BalancerToConcentratedPoolLink
		expectedResultingRecords    []gammmigration.BalancerToConcentratedPoolLink
		isPoolPrepared              bool
		isPreexistingRecordsSet     bool
		overwriteBalancerDenom0     string
		overwriteBalancerDenom1     string
		createFourAssetBalancerPool bool
		expectErr                   bool
	}{
		{
			name: "Non existent balancer pool.",
			testingMigrationRecords: []gammmigration.BalancerToConcentratedPoolLink{{
				BalancerPoolId: 9,
				ClPoolId:       6,
			}},
			isPreexistingRecordsSet: false,
			expectErr:               true,
		},
		{
			name: "Non existent concentrated pool.",
			testingMigrationRecords: []gammmigration.BalancerToConcentratedPoolLink{{
				BalancerPoolId: 1,
				ClPoolId:       9,
			}},
			isPreexistingRecordsSet: false,
			expectErr:               true,
		},
		{
			name: "Adding two of the same balancer pool ids at once should error",
			testingMigrationRecords: []gammmigration.BalancerToConcentratedPoolLink{
				{
					BalancerPoolId: 1,
					ClPoolId:       6,
				},
				{
					BalancerPoolId: 1,
					ClPoolId:       7,
				},
			},
			isPreexistingRecordsSet: true,
			expectErr:               true,
		},
		{
			name: "Adding two of the same cl pool ids at once should error",
			testingMigrationRecords: []gammmigration.BalancerToConcentratedPoolLink{
				{
					BalancerPoolId: 1,
					ClPoolId:       6,
				},
				{
					BalancerPoolId: 2,
					ClPoolId:       6,
				},
			},
			isPreexistingRecordsSet: true,
			expectErr:               true,
		},
		{
			name: "Normal case with two records",
			testingMigrationRecords: []gammmigration.BalancerToConcentratedPoolLink{
				{
					BalancerPoolId: 1,
					ClPoolId:       6,
				},
				{
					BalancerPoolId: 2,
					ClPoolId:       8,
				},
			},
			expectedResultingRecords: []gammmigration.BalancerToConcentratedPoolLink{
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
			isPreexistingRecordsSet: true,
			expectErr:               false,
		},
		{
			name: "Normal case with two records no preexisting records",
			testingMigrationRecords: []gammmigration.BalancerToConcentratedPoolLink{
				{
					BalancerPoolId: 1,
					ClPoolId:       6,
				},
				{
					BalancerPoolId: 2,
					ClPoolId:       8,
				},
			},
			expectedResultingRecords: []gammmigration.BalancerToConcentratedPoolLink{
				{
					BalancerPoolId: 1,
					ClPoolId:       6,
				},
				{
					BalancerPoolId: 2,
					ClPoolId:       8,
				},
			},
			isPreexistingRecordsSet: false,
			expectErr:               false,
		},
		{
			name: "Modify existing record, delete existing record, leave a record alone, add new record",
			testingMigrationRecords: []gammmigration.BalancerToConcentratedPoolLink{
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
			expectedResultingRecords: []gammmigration.BalancerToConcentratedPoolLink{
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
			isPreexistingRecordsSet: true,
			expectErr:               false,
		},
		{
			name: "Try to set one of the BalancerPoolIds to a cl pool Id",
			testingMigrationRecords: []gammmigration.BalancerToConcentratedPoolLink{
				{
					BalancerPoolId: 2,
					ClPoolId:       4,
				},
				{
					BalancerPoolId: 5,
					ClPoolId:       6,
				},
			},
			isPreexistingRecordsSet: true,
			expectErr:               true,
		},
		{
			name: "Try to set one of the ClPoolIds to a balancer pool Id",
			testingMigrationRecords: []gammmigration.BalancerToConcentratedPoolLink{
				{
					BalancerPoolId: 2,
					ClPoolId:       1,
				},
			},
			isPreexistingRecordsSet: true,
			expectErr:               true,
		},
		{
			name: "Mismatch denom0 between the two pools",
			testingMigrationRecords: []gammmigration.BalancerToConcentratedPoolLink{
				{
					BalancerPoolId: 1,
					ClPoolId:       6,
				},
			},
			overwriteBalancerDenom0: "osmo",
			isPreexistingRecordsSet: false,
			expectErr:               true,
		},
		{
			name: "Mismatch denom1 between the two pools",
			testingMigrationRecords: []gammmigration.BalancerToConcentratedPoolLink{
				{
					BalancerPoolId: 1,
					ClPoolId:       6,
				},
			},
			overwriteBalancerDenom1: "osmo",
			isPreexistingRecordsSet: false,
			expectErr:               true,
		},
		{
			name: "Balancer pool has more than two tokens",
			testingMigrationRecords: []gammmigration.BalancerToConcentratedPoolLink{
				{
					BalancerPoolId: 9,
					ClPoolId:       6,
				},
			},
			isPreexistingRecordsSet:     false,
			createFourAssetBalancerPool: true,
			expectErr:                   true,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()
			keeper := s.App.GAMMKeeper

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
				s.PrepareBalancerPoolWithCoins(poolCoins...)
			}
			for i := 0; i < 4; i++ {
				s.PrepareCustomConcentratedPool(s.TestAccs[0], ETH, USDC, defaultTickSpacing, sdk.ZeroDec())
			}
			// Four asset balancer pool ID if created: 9
			if test.createFourAssetBalancerPool {
				s.PrepareBalancerPool()
			}

			if test.isPreexistingRecordsSet {
				// Set up existing records so we can update them
				existingRecords := []gammmigration.BalancerToConcentratedPoolLink{
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
				err := keeper.ReplaceMigrationRecords(s.Ctx, existingRecords)
				s.Require().NoError(err)
			}

			err := keeper.UpdateMigrationRecords(s.Ctx, test.testingMigrationRecords)
			if test.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				migrationInfo, err := keeper.GetAllMigrationInfo(s.Ctx)
				s.Require().NoError(err)
				s.Require().Equal(len(test.expectedResultingRecords), len(migrationInfo.BalancerToConcentratedPoolLinks))
				for i, record := range test.expectedResultingRecords {
					s.Require().Equal(record.BalancerPoolId, migrationInfo.BalancerToConcentratedPoolLinks[i].BalancerPoolId)
					s.Require().Equal(record.ClPoolId, migrationInfo.BalancerToConcentratedPoolLinks[i].ClPoolId)
				}
			}
		})
	}
}

func (s *KeeperTestSuite) TestGetLinkedConcentratedPoolID() {
	tests := []struct {
		name                   string
		poolIdLeaving          []uint64
		expectedPoolIdEntering []uint64
		expectedErr            error
	}{
		{
			name:                   "Happy path",
			poolIdLeaving:          []uint64{1, 2, 3},
			expectedPoolIdEntering: []uint64{4, 5, 6},
		},
		{
			name:          "error: set poolIdLeaving to a concentrated pool ID",
			poolIdLeaving: []uint64{4},
			expectedErr:   types.ConcentratedPoolMigrationLinkNotFoundError{PoolIdLeaving: 4},
		},
		{
			name:          "error: set poolIdLeaving to a non existent pool ID",
			poolIdLeaving: []uint64{7},
			expectedErr:   types.ConcentratedPoolMigrationLinkNotFoundError{PoolIdLeaving: 7},
		},
	}

	for _, test := range tests {
		test := test
		s.Run(test.name, func() {
			s.SetupTest()
			keeper := s.App.GAMMKeeper

			// Our testing environment is as follows:
			// Balancer pool IDs: 1, 2, 3
			// Concentrated pool IDs: 3, 4, 5
			s.PrepareMultipleBalancerPools(3)
			s.PrepareMultipleConcentratedPools(3)

			keeper.OverwriteMigrationRecordsAndRedirectDistrRecords(s.Ctx, DefaultMigrationRecords)

			for i, poolIdLeaving := range test.poolIdLeaving {
				poolIdEntering, err := keeper.GetLinkedConcentratedPoolID(s.Ctx, poolIdLeaving)
				if test.expectedErr != nil {
					s.Require().Error(err)
					s.Require().ErrorIs(err, test.expectedErr)
					s.Require().Zero(poolIdEntering)
				} else {
					s.Require().NoError(err)
					s.Require().Equal(test.expectedPoolIdEntering[i], poolIdEntering)
				}
			}
		})
	}
}

func (s *KeeperTestSuite) TestGetLinkedBalancerPoolID() {
	tests := []struct {
		name                  string
		poolIdEntering        []uint64
		expectedPoolIdLeaving []uint64

		skipLinking bool
		expectedErr []error
	}{
		{
			name:                  "Happy path",
			poolIdEntering:        []uint64{4, 5, 6},
			expectedPoolIdLeaving: []uint64{1, 2, 3},
		},
		{
			name:           "error: set poolIdEntering to a balancer pool ID",
			poolIdEntering: []uint64{3},
			expectedErr:    []error{types.BalancerPoolMigrationLinkNotFoundError{PoolIdEntering: 3}},
		},
		{
			name:           "error: set poolIdEntering to a non existent pool ID",
			poolIdEntering: []uint64{7},
			expectedErr:    []error{types.BalancerPoolMigrationLinkNotFoundError{PoolIdEntering: 7}},
		},
		{
			name:                  "error: pools exist but link does not",
			poolIdEntering:        []uint64{4, 5, 6},
			expectedPoolIdLeaving: []uint64{1, 2, 3},
			skipLinking:           true,
			expectedErr: []error{
				types.BalancerPoolMigrationLinkNotFoundError{PoolIdEntering: 4},
				types.BalancerPoolMigrationLinkNotFoundError{PoolIdEntering: 5},
				types.BalancerPoolMigrationLinkNotFoundError{PoolIdEntering: 6},
			},
		},
	}

	for _, test := range tests {
		test := test
		s.Run(test.name, func() {
			s.SetupTest()
			keeper := s.App.GAMMKeeper

			// Our testing environment is as follows:
			// Balancer pool IDs: 1, 2, 3
			// Concentrated pool IDs: 3, 4, 5
			s.PrepareMultipleBalancerPools(3)
			s.PrepareMultipleConcentratedPools(3)

			if !test.skipLinking {
				keeper.OverwriteMigrationRecordsAndRedirectDistrRecords(s.Ctx, DefaultMigrationRecords)
			}

			s.Require().True(len(test.poolIdEntering) > 0)
			for i, poolIdEntering := range test.poolIdEntering {
				poolIdLeaving, err := keeper.GetLinkedBalancerPoolID(s.Ctx, poolIdEntering)
				if test.expectedErr != nil {
					s.Require().Error(err)
					s.Require().ErrorIs(err, test.expectedErr[i])
					s.Require().Zero(poolIdLeaving)
				} else {
					s.Require().NoError(err)
					s.Require().Equal(test.expectedPoolIdLeaving[i], poolIdLeaving)
				}
			}
		})
	}
}

func (s *KeeperTestSuite) TestGetAllMigrationInfo() {
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
		s.Run(test.name, func() {
			s.SetupTest()
			keeper := s.App.GAMMKeeper

			// Our testing environment is as follows:
			// Balancer pool IDs: 1, 2, 3
			// Concentrated pool IDs: 3, 4, 5
			s.PrepareMultipleBalancerPools(3)
			s.PrepareMultipleConcentratedPools(3)

			if !test.skipLinking {
				keeper.OverwriteMigrationRecordsAndRedirectDistrRecords(s.Ctx, DefaultMigrationRecords)
			}

			migrationRecords, err := s.App.GAMMKeeper.GetAllMigrationInfo(s.Ctx)
			s.Require().NoError(err)
			if !test.skipLinking {
				s.Require().Equal(migrationRecords, DefaultMigrationRecords)
			} else {
				s.Require().Equal(len(migrationRecords.BalancerToConcentratedPoolLinks), 0)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestRedirectDistributionRecord() {
	suite.Setup()

	var (
		defaultUsdcAmount = sdk.NewInt(7300000000)
		defaultOsmoAmount = sdk.NewInt(10000000000)
		usdcCoin          = sdk.NewCoin("uusdc", defaultUsdcAmount)
		osmoCoin          = sdk.NewCoin("uosmo", defaultOsmoAmount)
	)

	longestLockableDuration, err := suite.App.PoolIncentivesKeeper.GetLongestLockableDuration(suite.Ctx)
	suite.Require().NoError(err)

	tests := map[string]struct {
		poolLiquidity sdk.Coins
		cfmmPoolId    uint64
		clPoolId      uint64
		expectError   error
	}{
		"happy path": {
			poolLiquidity: sdk.NewCoins(usdcCoin, osmoCoin),
			cfmmPoolId:    uint64(1),
			clPoolId:      uint64(3),
		},
		"error: cfmm pool ID doesn't exist": {
			poolLiquidity: sdk.NewCoins(usdcCoin, osmoCoin),
			cfmmPoolId:    uint64(4),
			clPoolId:      uint64(3),
			expectError:   poolincentivestypes.NoGaugeAssociatedWithPoolError{PoolId: 4, Duration: longestLockableDuration},
		},
		"error: cl pool ID doesn't exist": {
			poolLiquidity: sdk.NewCoins(usdcCoin, osmoCoin),
			cfmmPoolId:    uint64(1),
			clPoolId:      uint64(4),
			expectError:   poolincentivestypes.NoGaugeAssociatedWithPoolError{PoolId: 4, Duration: longestLockableDuration},
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			suite.SetupTest()

			// Create primary balancer pool.
			balancerId := suite.PrepareBalancerPoolWithCoins(tc.poolLiquidity...)
			balancerPool, err := suite.App.PoolManagerKeeper.GetPool(suite.Ctx, balancerId)
			suite.Require().NoError(err)

			// Create another balancer pool to test that its gauge links are unchanged
			balancerId2 := suite.PrepareBalancerPoolWithCoins(tc.poolLiquidity...)

			// Get gauges for both balancer pools.
			gaugeToRedirect, err := suite.App.PoolIncentivesKeeper.GetPoolGaugeId(suite.Ctx, balancerPool.GetId(), longestLockableDuration)
			suite.Require().NoError(err)
			gaugeToNotRedirect, err := suite.App.PoolIncentivesKeeper.GetPoolGaugeId(suite.Ctx, balancerId2, longestLockableDuration)
			suite.Require().NoError(err)

			// Distribution info prior to redirecting.
			originalDistrInfo := poolincentivestypes.DistrInfo{
				TotalWeight: sdk.NewInt(100),
				Records: []poolincentivestypes.DistrRecord{
					{
						GaugeId: gaugeToRedirect,
						Weight:  sdk.NewInt(50),
					},
					{
						GaugeId: gaugeToNotRedirect,
						Weight:  sdk.NewInt(50),
					},
				},
			}
			suite.App.PoolIncentivesKeeper.SetDistrInfo(suite.Ctx, originalDistrInfo)

			// Create concentrated pool.
			clPool := suite.PrepareCustomConcentratedPool(suite.TestAccs[0], tc.poolLiquidity[1].Denom, tc.poolLiquidity[0].Denom, 100, sdk.MustNewDecFromStr("0.001"))

			// Redirect distribution record from the primary balancer pool to the concentrated pool.
			err = suite.App.GAMMKeeper.RedirectDistributionRecord(suite.Ctx, tc.cfmmPoolId, tc.clPoolId)
			if tc.expectError != nil {
				suite.Require().Error(err)
				return
			}
			suite.Require().NoError(err)

			// Validate that the balancer gauge is now linked to the new concentrated pool.
			concentratedPoolGaugeId, err := suite.App.PoolIncentivesKeeper.GetPoolGaugeId(suite.Ctx, clPool.GetId(), suite.App.IncentivesKeeper.GetEpochInfo(suite.Ctx).Duration)
			suite.Require().NoError(err)
			distrInfo := suite.App.PoolIncentivesKeeper.GetDistrInfo(suite.Ctx)
			suite.Require().Equal(distrInfo.Records[0].GaugeId, concentratedPoolGaugeId)

			// Validate that distribution record from another pool is not redirected.
			suite.Require().Equal(distrInfo.Records[1].GaugeId, gaugeToNotRedirect)

			// Validate that old gauge still exist
			_, err = suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeToRedirect)
			suite.Require().NoError(err)
		})
	}
}
