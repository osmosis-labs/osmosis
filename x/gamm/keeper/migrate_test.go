package keeper_test

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	"github.com/osmosis-labs/osmosis/v27/x/gamm/types"
	gammmigration "github.com/osmosis-labs/osmosis/v27/x/gamm/types/migration"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v27/x/pool-incentives/types"
)

const (
	validPoolId = uint64(1)
)

var (
	DAIIBCDenom  = "ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7"
	USDCIBCDenom = "ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858"

	defaultDaiAmount, _ = osmomath.NewIntFromString("73000000000000000000000")
	defaultDenom0mount  = osmomath.NewInt(10000000000)
	desiredDenom0       = appparams.BaseCoinUnit
	desiredDenom0Coin   = sdk.NewCoin(desiredDenom0, defaultDenom0mount)
	daiCoin             = sdk.NewCoin(DAIIBCDenom, defaultDaiAmount)
	usdcCoin            = sdk.NewCoin(USDCIBCDenom, defaultDaiAmount)
)

func (s *KeeperTestSuite) TestMigrate() {
	defaultAccount := apptesting.CreateRandomAccounts(1)[0]
	defaultGammShares := sdk.NewCoin("gamm/pool/1", osmomath.MustNewDecFromStr("100000000000000000000").RoundInt())
	invalidGammShares := sdk.NewCoin("gamm/pool/1", osmomath.MustNewDecFromStr("190000000000000000001").RoundInt())
	defaultAccountFunds := sdk.NewCoins(sdk.NewCoin("eth", osmomath.NewInt(200000000000)), sdk.NewCoin("usdc", osmomath.NewInt(200000000000)))

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
		AdditiveTolerance: osmomath.NewDec(100000),
		RoundingDir:       osmomath.RoundDown,
	}
	defaultJoinTime := s.Ctx.BlockTime()

	type param struct {
		sender                sdk.AccAddress
		sharesToMigrateDenom  string
		sharesToMigrateAmount osmomath.Int
	}

	tests := []struct {
		name                   string
		param                  param
		expectedErr            error
		sharesToCreate         osmomath.Int
		tokenOutMins           sdk.Coins
		expectedLiquidity      osmomath.Dec
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
			expectedLiquidity:      osmomath.MustNewDecFromStr("100000000000.000000010000000000"),
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
			expectedLiquidity:      osmomath.MustNewDecFromStr("100000000000.000000010000000000"),
			setupPoolMigrationLink: false,
			expectedErr:            types.ConcentratedPoolMigrationLinkNotFoundError{PoolIdLeaving: 1},
			errTolerance:           defaultErrorTolerance,
		},
		{
			name: "migrate half of the shares",
			param: param{
				sender:                defaultAccount,
				sharesToMigrateDenom:  defaultGammShares.Denom,
				sharesToMigrateAmount: defaultGammShares.Amount.Quo(osmomath.NewInt(2)),
			},
			sharesToCreate:         defaultGammShares.Amount,
			expectedLiquidity:      osmomath.MustNewDecFromStr("50000000000.000000005000000000"),
			setupPoolMigrationLink: true,
			errTolerance:           defaultErrorTolerance,
		},
		{
			name: "double the created shares, migrate 1/4 of the shares",
			param: param{
				sender:                defaultAccount,
				sharesToMigrateDenom:  defaultGammShares.Denom,
				sharesToMigrateAmount: defaultGammShares.Amount.Quo(osmomath.NewInt(2)),
			},
			sharesToCreate:         defaultGammShares.Amount.Mul(osmomath.NewInt(2)),
			expectedLiquidity:      osmomath.MustNewDecFromStr("49999999999.000000004999999999"),
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
			expectedLiquidity:      osmomath.MustNewDecFromStr("100000000000.000000010000000000"),
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
			tokenOutMins:           sdk.NewCoins(sdk.NewCoin(ETH, osmomath.NewInt(90000000000)), sdk.NewCoin(USDC, osmomath.NewInt(90000000000))),
			expectedLiquidity:      osmomath.MustNewDecFromStr("100000000000.000000010000000000"),
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
			tokenOutMins:           sdk.NewCoins(sdk.NewCoin(ETH, osmomath.NewInt(110000000000)), sdk.NewCoin(USDC, osmomath.NewInt(110000000000))),
			expectedLiquidity:      osmomath.MustNewDecFromStr("100000000000.000000010000000000"),
			setupPoolMigrationLink: true,
			expectedErr: errorsmod.Wrapf(types.ErrLimitMinAmount,
				"Exit pool returned %s , minimum tokens out specified as %s",
				sdk.NewCoins(sdk.NewCoin(ETH, osmomath.NewInt(100000000000)), sdk.NewCoin(USDC, osmomath.NewInt(100000000000))), sdk.NewCoins(sdk.NewCoin(ETH, osmomath.NewInt(110000000000)), sdk.NewCoin(USDC, osmomath.NewInt(110000000000)))),
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
			tokenOutMins:           sdk.NewCoins(sdk.NewCoin(ETH, osmomath.NewInt(110000000000)), sdk.NewCoin(USDC, osmomath.NewInt(100000000000))),
			expectedLiquidity:      osmomath.MustNewDecFromStr("100000000000.000000010000000000"),
			setupPoolMigrationLink: true,
			expectedErr: errorsmod.Wrapf(types.ErrLimitMinAmount,
				"Exit pool returned %s , minimum tokens out specified as %s",
				sdk.NewCoins(sdk.NewCoin(ETH, osmomath.NewInt(100000000000)), sdk.NewCoin(USDC, osmomath.NewInt(100000000000))), sdk.NewCoins(sdk.NewCoin(ETH, osmomath.NewInt(110000000000)), sdk.NewCoin(USDC, osmomath.NewInt(100000000000)))),
			errTolerance: defaultErrorTolerance,
		},
	}

	for _, test := range tests {
		s.SetupTest()
		s.Ctx = s.Ctx.WithBlockTime(defaultJoinTime)
		keeper := s.App.GAMMKeeper

		// Prepare both balancer and concentrated pools
		s.FundAcc(test.param.sender, defaultAccountFunds)
		balancerPoolId := s.PrepareBalancerPoolWithCoins(sdk.NewCoin("eth", osmomath.NewInt(100000000000)), sdk.NewCoin("usdc", osmomath.NewInt(100000000000)))
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
		_, _, err = s.App.GAMMKeeper.JoinPoolNoSwap(s.Ctx, test.param.sender, balancerPoolId, test.sharesToCreate, sdk.NewCoins(sdk.NewCoin("eth", osmomath.NewInt(999999999999999)), sdk.NewCoin("usdc", osmomath.NewInt(999999999999999))))
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
		expectedCoinsOut, err := balancerPool.CalcExitPoolCoinsFromShares(s.Ctx, sharesToMigrate.Amount, osmomath.ZeroDec())
		s.Require().NoError(err)

		// Migrate the user's gamm shares to a full range concentrated liquidity position
		userBalancesBeforeMigration := s.App.BankKeeper.GetAllBalances(s.Ctx, test.param.sender)
		positionData, migratedPools, err := keeper.MigrateUnlockedPositionFromBalancerToConcentrated(s.Ctx, test.param.sender, sharesToMigrate, test.tokenOutMins)
		userBalancesAfterMigration := s.App.BankKeeper.GetAllBalances(s.Ctx, test.param.sender)
		if test.expectedErr != nil {
			s.Require().Error(err)
			s.Require().ErrorContains(err, test.expectedErr.Error())

			// Expect zero values for both pool ids
			s.Require().Zero(migratedPools.LeavingID)
			s.Require().Zero(migratedPools.EnteringID)

			// Assure the user's gamm shares still exist
			userGammBalanceAfterFailedMigration := s.App.BankKeeper.GetBalance(s.Ctx, test.param.sender, "gamm/pool/1")
			s.Require().Equal(userGammBalancePostJoin.String(), userGammBalanceAfterFailedMigration.String())

			// Assure cl pool has no balance after a failed migration.
			clPoolEthBalanceAfterFailedMigration := s.App.BankKeeper.GetBalance(s.Ctx, clPoolAddress, ETH)
			clPoolUsdcBalanceAfterFailedMigration := s.App.BankKeeper.GetBalance(s.Ctx, clPoolAddress, USDC)
			s.Require().Equal(osmomath.NewInt(0), clPoolEthBalanceAfterFailedMigration.Amount)
			s.Require().Equal(osmomath.NewInt(0), clPoolUsdcBalanceAfterFailedMigration.Amount)

			// Assure the position was not created.
			// TODO: When we implement lock breaking, we need to change time.Time{} to the lock's end time.
			_, err := s.App.ConcentratedLiquidityKeeper.GetPositionLiquidity(s.Ctx, positionData.ID)
			s.Require().Error(err)
			continue
		}
		s.Require().NoError(err)

		// Expect the poolIdLeaving to be the balancer pool id
		// Expect the poolIdEntering to be the concentrated liquidity pool id
		s.Require().Equal(balancerPoolId, migratedPools.LeavingID)
		s.Require().Equal(clPool.GetId(), migratedPools.EnteringID)

		// Determine how much of the user's balance was not used in the migration
		// This amount should be returned to the user.
		expectedUserFinalEthBalanceDiff := expectedCoinsOut.AmountOf(ETH).Sub(positionData.Amount0)
		expectedUserFinalUsdcBalanceDiff := expectedCoinsOut.AmountOf(USDC).Sub(positionData.Amount1)
		s.Require().Equal(userBalancesBeforeMigration.AmountOf(ETH).Add(expectedUserFinalEthBalanceDiff).String(), userBalancesAfterMigration.AmountOf(ETH).String())
		s.Require().Equal(userBalancesBeforeMigration.AmountOf(USDC).Add(expectedUserFinalUsdcBalanceDiff).String(), userBalancesAfterMigration.AmountOf(USDC).String())

		// Assure the expected position was created.
		// TODO: When we implement lock breaking, we need to change time.Time{} to the lock's end time.
		position, err := s.App.ConcentratedLiquidityKeeper.GetPositionLiquidity(s.Ctx, positionData.ID)
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
		osmoassert.Equal(s.T(), test.errTolerance, userEthBalanceTransferredToClPool.Amount, clPoolEthBalanceAfterMigration.Amount)
		osmoassert.Equal(s.T(), test.errTolerance, userUsdcBalanceTransferredToClPool.Amount, clPoolUsdcBalanceAfterMigration.Amount)

		// Assert user amount transferred to cl pool from gamm pool should be equal to the amount we migrated from the migrate message.
		// This test is within 100 shares due to rounding that occurs from utilizing .000000000000000001 instead of 0.
		osmoassert.Equal(s.T(), test.errTolerance, userEthBalanceTransferredToClPool.Amount, positionData.Amount0)
		osmoassert.Equal(s.T(), test.errTolerance, userUsdcBalanceTransferredToClPool.Amount, positionData.Amount0)
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
			overwriteBalancerDenom0: appparams.BaseCoinUnit,
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
			overwriteBalancerDenom1: appparams.BaseCoinUnit,
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

			defaultBalancerCoin0 := sdk.NewCoin(ETH, osmomath.NewInt(1000000000))
			defaultBalancerCoin1 := sdk.NewCoin(USDC, osmomath.NewInt(1000000000))

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
				s.PrepareCustomConcentratedPool(s.TestAccs[0], ETH, USDC, defaultTickSpacing, osmomath.ZeroDec())
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

			defaultBalancerCoin0 := sdk.NewCoin(ETH, osmomath.NewInt(1000000000))
			defaultBalancerCoin1 := sdk.NewCoin(USDC, osmomath.NewInt(1000000000))

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
				s.PrepareCustomConcentratedPool(s.TestAccs[0], ETH, USDC, defaultTickSpacing, osmomath.ZeroDec())
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

			keeper.OverwriteMigrationRecords(s.Ctx, DefaultMigrationRecords)

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
				keeper.OverwriteMigrationRecords(s.Ctx, DefaultMigrationRecords)
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
				keeper.OverwriteMigrationRecords(s.Ctx, DefaultMigrationRecords)
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

func (s *KeeperTestSuite) TestCreateConcentratedPoolFromCFMM() {
	tests := map[string]struct {
		poolLiquidity sdk.Coins

		cfmmPoolIdToLinkWith uint64
		desiredDenom0        string
		expectedDenoms       []string
		expectError          error
	}{
		"success": {
			poolLiquidity:        sdk.NewCoins(desiredDenom0Coin, daiCoin),
			cfmmPoolIdToLinkWith: validPoolId,
			desiredDenom0:        desiredDenom0,
			expectedDenoms:       []string{desiredDenom0, daiCoin.Denom},
		},
		"error: invalid denom 0": {
			poolLiquidity:        sdk.NewCoins(desiredDenom0Coin, daiCoin),
			cfmmPoolIdToLinkWith: validPoolId,
			desiredDenom0:        USDCIBCDenom,
			expectError:          types.NoDesiredDenomInPoolError{DesiredDenom: USDCIBCDenom},
		},
		"error: pool with 3 assets, must have two": {
			poolLiquidity:        sdk.NewCoins(desiredDenom0Coin, daiCoin, usdcCoin),
			cfmmPoolIdToLinkWith: validPoolId,
			desiredDenom0:        USDCIBCDenom,
			expectError:          types.ErrMustHaveTwoDenoms,
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			s.SetupTest()

			balancerId := s.PrepareBalancerPoolWithCoins(tc.poolLiquidity...)

			balancerPool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, balancerId)
			s.Require().NoError(err)

			clPoolReturned, err := s.App.GAMMKeeper.CreateConcentratedPoolFromCFMM(s.Ctx, tc.cfmmPoolIdToLinkWith, tc.desiredDenom0, osmomath.ZeroDec(), defaultTickSpacing)

			if tc.expectError != nil {
				s.Require().Error(err)
				s.Require().Nil(clPoolReturned)
				return
			}
			s.Require().NoError(err)

			// Validate that pool saved in state is the same as the one returned
			clPoolInState, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, clPoolReturned.GetId())
			s.Require().NoError(err)
			s.Require().Equal(clPoolReturned, clPoolInState)

			// Validate CL and balancer pools have the same spread factor.
			s.Require().Equal(balancerPool.GetSpreadFactor(s.Ctx), clPoolReturned.GetSpreadFactor(s.Ctx))

			// Validate that CL and balancer pools have the same denoms
			balancerDenoms, err := s.App.PoolManagerKeeper.RouteGetPoolDenoms(s.Ctx, balancerPool.GetId())
			s.Require().NoError(err)

			concentratedDenoms, err := s.App.PoolManagerKeeper.RouteGetPoolDenoms(s.Ctx, clPoolReturned.GetId())
			s.Require().NoError(err)

			// Order between balancer and concentrated might differ
			// because balancer lexicographically orders denoms but CL does not.
			s.Require().ElementsMatch(balancerDenoms, concentratedDenoms)
			s.Require().Equal(tc.expectedDenoms, concentratedDenoms)
		})
	}
}

func (s *KeeperTestSuite) TestCreateCanonicalConcentratedLiquidityPoolAndMigrationLink() {
	s.Setup()

	longestLockableDuration, err := s.App.PoolIncentivesKeeper.GetLongestLockableDuration(s.Ctx)
	s.Require().NoError(err)

	tests := map[string]struct {
		poolLiquidity              sdk.Coins
		cfmmPoolIdToLinkWith       uint64
		desiredDenom0              string
		expectedBalancerDenoms     []string
		expectedConcentratedDenoms []string
		expectError                error
	}{
		"success - denoms reordered relative to balancer": {
			poolLiquidity:        sdk.NewCoins(desiredDenom0Coin, daiCoin),
			cfmmPoolIdToLinkWith: validPoolId,
			// lexicographically ordered
			expectedBalancerDenoms: []string{daiCoin.Denom, desiredDenom0Coin.Denom},
			// determined by desired denom 0
			expectedConcentratedDenoms: []string{desiredDenom0Coin.Denom, daiCoin.Denom},
			desiredDenom0:              desiredDenom0,
		},
		"success - denoms are not reordered relative to balancer": {
			poolLiquidity:        sdk.NewCoins(desiredDenom0Coin, daiCoin),
			cfmmPoolIdToLinkWith: validPoolId,
			// lexicographically ordered
			expectedBalancerDenoms: []string{daiCoin.Denom, desiredDenom0Coin.Denom},
			// determined by desired denom 0
			expectedConcentratedDenoms: []string{daiCoin.Denom, desiredDenom0Coin.Denom},
			desiredDenom0:              daiCoin.Denom,
		},
		"error: invalid denom 0": {
			poolLiquidity:        sdk.NewCoins(desiredDenom0Coin, daiCoin),
			cfmmPoolIdToLinkWith: validPoolId,
			desiredDenom0:        USDCIBCDenom,
			expectError:          types.NoDesiredDenomInPoolError{DesiredDenom: USDCIBCDenom},
		},
		"error: pool with 3 assets, must have two": {
			poolLiquidity:        sdk.NewCoins(desiredDenom0Coin, daiCoin, usdcCoin),
			cfmmPoolIdToLinkWith: validPoolId,
			desiredDenom0:        USDCIBCDenom,
			expectError:          types.ErrMustHaveTwoDenoms,
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			s.SetupTest()

			balancerId := s.PrepareBalancerPoolWithCoins(tc.poolLiquidity...)

			// Another pool for testing that its gauge and migration links are unchanged.
			balancerId2 := s.PrepareBalancerPoolWithCoins(tc.poolLiquidity...)

			// Another pool for testing that previously existing migration links don't get overwritten.
			balancerId3 := s.PrepareBalancerPoolWithCoins(tc.poolLiquidity...)

			clPoolOld, err := s.App.GAMMKeeper.CreateCanonicalConcentratedLiquidityPoolAndMigrationLink(s.Ctx, balancerId3, tc.desiredDenom0, osmomath.ZeroDec(), defaultTickSpacing)

			balancerPool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, balancerId)
			s.Require().NoError(err)

			// Get balancer gauges.
			gaugeToRedirect, _ := s.App.PoolIncentivesKeeper.GetPoolGaugeId(s.Ctx, balancerPool.GetId(), longestLockableDuration)

			gaugeToNotRedeirect, _ := s.App.PoolIncentivesKeeper.GetPoolGaugeId(s.Ctx, balancerId2, longestLockableDuration)

			originalDistrInfo := poolincentivestypes.DistrInfo{
				TotalWeight: osmomath.NewInt(100),
				Records: []poolincentivestypes.DistrRecord{
					{
						GaugeId: gaugeToRedirect,
						Weight:  osmomath.NewInt(50),
					},
					{
						GaugeId: gaugeToNotRedeirect,
						Weight:  osmomath.NewInt(50),
					},
				},
			}
			s.App.PoolIncentivesKeeper.SetDistrInfo(s.Ctx, originalDistrInfo)

			// CreateCanonicalConcentratedLiquidityPoolAndMigration is used to change the distribution records and now no longer does.
			// We take the distribution records before execution to ensure it is not changed.
			distrInfoPre := s.App.PoolIncentivesKeeper.GetDistrInfo(s.Ctx)

			clPool, err := s.App.GAMMKeeper.CreateCanonicalConcentratedLiquidityPoolAndMigrationLink(s.Ctx, tc.cfmmPoolIdToLinkWith, tc.desiredDenom0, osmomath.ZeroDec(), defaultTickSpacing)

			if tc.expectError != nil {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)

			// Get the new concentrated pool.
			// Note, +4 because we create 3 balancer pools and 1 cl pool during test setup, and 1 concentrated pool during migration.
			clPoolInState, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, validPoolId+4)
			s.Require().NoError(err)
			s.Require().Equal(clPool, clPoolInState)

			// Validate that CL and balancer pools have the same denoms
			balancerDenoms, err := s.App.PoolManagerKeeper.RouteGetPoolDenoms(s.Ctx, balancerPool.GetId())
			s.Require().NoError(err)

			concentratedDenoms, err := s.App.PoolManagerKeeper.RouteGetPoolDenoms(s.Ctx, clPoolInState.GetId())
			s.Require().NoError(err)

			// This check does not guarantee order.
			s.Require().ElementsMatch(balancerDenoms, concentratedDenoms)

			// Validate order of balancer denoms is lexicographically sorted.
			s.Require().Equal(tc.expectedBalancerDenoms, balancerDenoms)

			// Validate order of concentrated pool denoms which might be different from balancer.
			s.Require().Equal(tc.expectedConcentratedDenoms, concentratedDenoms)

			// Validate that the new concentrated pool has a gauge.
			_, err = s.App.PoolIncentivesKeeper.GetPoolGaugeId(s.Ctx, clPoolInState.GetId(), s.App.IncentivesKeeper.GetEpochInfo(s.Ctx).Duration)
			s.Require().NoError(err)

			// Ensure the distribution records are unchanged.
			distrInfoPost := s.App.PoolIncentivesKeeper.GetDistrInfo(s.Ctx)
			s.Require().Equal(distrInfoPre, distrInfoPost)

			// Validate migration record.
			migrationInfo, err := s.App.GAMMKeeper.GetAllMigrationInfo(s.Ctx)
			s.Require().NoError(err)
			s.Require().Equal(migrationInfo, gammmigration.MigrationRecords{
				BalancerToConcentratedPoolLinks: []gammmigration.BalancerToConcentratedPoolLink{
					{
						BalancerPoolId: balancerId,
						ClPoolId:       clPoolInState.GetId(),
					},
					{
						BalancerPoolId: balancerId3,
						ClPoolId:       clPoolOld.GetId(),
					},
				},
			})

			// Validate that old gauge still exist
			_, err = s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, gaugeToRedirect)
			s.Require().NoError(err)
		})
	}
}
