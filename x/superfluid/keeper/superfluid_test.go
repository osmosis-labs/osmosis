package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	appparams "github.com/osmosis-labs/osmosis/v7/app/params"
	"github.com/osmosis-labs/osmosis/v7/x/epochs"
	epochstypes "github.com/osmosis-labs/osmosis/v7/x/epochs/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v7/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
)

// TODO: add more test cases
func (suite *KeeperTestSuite) TestSuperfluidFlow() {
	testCases := []struct {
		name               string
		validatorStats     []stakingtypes.BondStatus
		superDelegation    superfluidDelegation
		expInterDelegation sdk.Dec
	}{
		{
			"with single validator and single superfluid delegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			superfluidDelegation{0, "gamm/pool/1"},
			sdk.NewDec(19000000), // 95% x 2 x 1000000
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()

			// we set suite.ctx so that other suite methods are ran in the correct block height and block time
			now := time.Now()
			suite.ctx = suite.ctx.WithBlockHeight(1).WithBlockTime(now)
			// run initial setup for epoch, starting with deleting old epoch information
			epochInfos := suite.app.EpochsKeeper.AllEpochInfos(suite.ctx)
			for _, epochInfo := range epochInfos {
				suite.app.EpochsKeeper.DeleteEpochInfo(suite.ctx, epochInfo.Identifier)
			}

			epochInfo := epochstypes.EpochInfo{
				Identifier:            "day",
				StartTime:             suite.ctx.BlockTime(),
				Duration:              time.Hour * 24,
				CurrentEpoch:          0,
				CurrentEpochStartTime: time.Time{},
				EpochCountingStarted:  false,
			}
			suite.app.EpochsKeeper.SetEpochInfo(suite.ctx, epochInfo)

			epochs.BeginBlocker(suite.ctx, *suite.app.EpochsKeeper)

			poolId := suite.createGammPool([]string{appparams.BaseCoinUnit, "foo"})
			suite.Require().Equal(poolId, uint64(1))

			// setup validators
			valAddrs := suite.SetupValidators(tc.validatorStats)

			// setup superfluid delegations
			//flagIntermediaryAcc := make(map[string]bool)
			//intermediaryAccs := []types.SuperfluidIntermediaryAccount{}
			locks := []lockuptypes.PeriodLock{}

			valAddr := valAddrs[tc.superDelegation.valIndex]

			params := suite.app.SuperfluidKeeper.GetParams(suite.ctx)
			suite.app.IncentivesKeeper.SetLockableDurations(suite.ctx, []time.Duration{
				time.Hour * 24 * 14,
				time.Hour,
				time.Hour * 3,
				time.Hour * 7,
				params.UnbondingDuration,
			})

			// register a LP token as a superfluid asset
			suite.app.SuperfluidKeeper.SetSuperfluidAsset(suite.ctx, types.SuperfluidAsset{
				Denom:     tc.superDelegation.lpDenom,
				AssetType: types.SuperfluidAssetTypeLPShare,
			})

			suite.app.SuperfluidKeeper.SetEpochOsmoEquivalentTWAP(suite.ctx, 1, tc.superDelegation.lpDenom, sdk.NewDec(20))
			// create lockup of LP token
			addr1 := sdk.AccAddress([]byte("addr1---------------"))
			coins := sdk.Coins{sdk.NewInt64Coin(tc.superDelegation.lpDenom, 1000000)}
			lock := suite.LockTokens(addr1, coins, params.UnbondingDuration)

			// call SuperfluidDelegate and check response
			err := suite.app.SuperfluidKeeper.SuperfluidDelegate(suite.ctx, lock.Owner, lock.ID, valAddr.String())
			suite.Require().NoError(err)

			expAcc := types.NewSuperfluidIntermediaryAccount(lock.Coins[0].Denom, valAddr.String(), 0)

			// save locks for future use
			locks = append(locks, lock)
			lock = locks[0]
			valAddr = valAddrs[tc.superDelegation.valIndex]

			// check synthetic lockup creation
			synthLock, err := suite.app.LockupKeeper.GetSyntheticLockup(suite.ctx, lock.ID, keeper.StakingSuffix(valAddr.String()))
			suite.Require().NoError(err)
			suite.Require().Equal(synthLock.UnderlyingLockId, lock.ID)
			suite.Require().Equal(synthLock.Suffix, keeper.StakingSuffix(valAddr.String()))
			suite.Require().Equal(synthLock.EndTime, time.Time{})

			expAcc = types.NewSuperfluidIntermediaryAccount(lock.Coins[0].Denom, valAddr.String(), 0)

			// Check lockID connection with intermediary account
			intAcc := suite.app.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(suite.ctx, lock.ID)
			suite.Require().Equal(intAcc.String(), expAcc.GetAccAddress().String())

			// check intermediary account creation
			gotAcc := suite.app.SuperfluidKeeper.GetIntermediaryAccount(suite.ctx, expAcc.GetAccAddress())
			suite.Require().Equal(gotAcc.Denom, expAcc.Denom)
			suite.Require().Equal(gotAcc.ValAddr, expAcc.ValAddr)
			suite.Require().GreaterOrEqual(gotAcc.GaugeId, uint64(1))

			valAddr, err = sdk.ValAddressFromBech32(expAcc.ValAddr)
			suite.Require().NoError(err)

			// check gauge creation
			gauge, err := suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gotAcc.GaugeId)
			suite.Require().NoError(err)
			suite.Require().Equal(gauge.Id, gotAcc.GaugeId)
			suite.Require().Equal(gauge.IsPerpetual, true)
			suite.Require().Equal(gauge.DistributeTo, lockuptypes.QueryCondition{
				LockQueryType: lockuptypes.ByDuration,
				Denom:         expAcc.Denom + keeper.StakingSuffix(valAddr.String()),
				Duration:      params.UnbondingDuration,
			})
			suite.Require().Equal(gauge.Coins, sdk.Coins(nil))
			suite.Require().Equal(gauge.StartTime, suite.ctx.BlockTime())
			suite.Require().Equal(gauge.NumEpochsPaidOver, uint64(1))
			suite.Require().Equal(gauge.FilledEpochs, uint64(0))
			suite.Require().Equal(gauge.DistributedCoins, sdk.Coins(nil))

			// check delegation from intermediary account to validator
			delegation, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, expAcc.GetAccAddress(), valAddr)
			suite.Require().True(found)
			suite.Require().Equal(delegation.Shares, tc.expInterDelegation)

			// try delegating twice with same lockup
			for _, lock := range locks {
				err := suite.app.SuperfluidKeeper.SuperfluidDelegate(suite.ctx, lock.Owner, lock.ID, valAddrs[0].String())
				suite.Require().Error(err)
			}

			// ensure that we do not have dangling rewards initially
			totalReward := suite.app.DistrKeeper.GetTotalRewards(suite.ctx)
			suite.Require().Equal(0, totalReward.Len())

			// allocate 20_000 stake to rewards to validators,
			// this is done manually per testing environment, should be automatically happening
			suite.allocateRewardsToValidator(valAddrs[0])
			totalReward = suite.app.DistrKeeper.GetTotalRewards(suite.ctx)

			// now we test if `AfterEpochEnd` shpws normal behaviour
			// we first trigger AfterEpochEnd by incrementing block height and time
			suite.ctx = suite.ctx.WithBlockHeight(2).WithBlockTime(now.Add(time.Hour * 25))
			epochs.BeginBlocker(suite.ctx, *suite.app.EpochsKeeper)

			gauge, err = suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gotAcc.GaugeId)
			suite.Require().NoError(err)
			suite.Require().Equal(gauge.Id, gotAcc.GaugeId)
			suite.Require().Equal(gauge.IsPerpetual, true)
			suite.Require().Equal(gauge.DistributeTo, lockuptypes.QueryCondition{
				LockQueryType: lockuptypes.ByDuration,
				Denom:         "gamm/pool/1" + keeper.StakingSuffix(valAddrs[0].String()),
				Duration:      params.UnbondingDuration,
			})

			// check if staking rewards has been passed to gauges
			distributedReward := gauge.Coins.AmountOf(sdk.DefaultBondDenom)
			// // 190_00 = 95% x 20_000
			suite.Require().Equal(sdk.NewInt(19_000), distributedReward)
		})
	}
}
