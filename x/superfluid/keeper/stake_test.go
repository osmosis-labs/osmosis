package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	appparams "github.com/osmosis-labs/osmosis/app/params"
	epochstypes "github.com/osmosis-labs/osmosis/x/epochs/types"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
	"github.com/osmosis-labs/osmosis/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/x/superfluid/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

func (suite *KeeperTestSuite) LockTokens(addr sdk.AccAddress, coins sdk.Coins, duration time.Duration) lockuptypes.PeriodLock {
	err := suite.app.BankKeeper.SetBalances(suite.ctx, addr, coins)
	suite.Require().NoError(err)
	lock, err := suite.app.LockupKeeper.LockTokens(suite.ctx, addr, coins, duration)
	suite.Require().NoError(err)
	return lock
}

func (suite *KeeperTestSuite) SetupValidator(bondStatus stakingtypes.BondStatus) sdk.ValAddress {
	valPub := secp256k1.GenPrivKey().PubKey()
	valAddr := sdk.ValAddress(valPub.Address())

	validator, err := stakingtypes.NewValidator(valAddr, valPub, stakingtypes.NewDescription("moniker", "", "", "", ""))
	suite.Require().NoError(err)

	amount := sdk.TokensFromConsensusPower(1)
	issuedShares := amount.ToDec()
	validator.Status = bondStatus
	validator.Tokens = validator.Tokens.Add(amount)
	validator.DelegatorShares = validator.DelegatorShares.Add(issuedShares)

	suite.app.StakingKeeper.SetValidator(suite.ctx, validator)
	suite.app.StakingKeeper.SetValidatorByConsAddr(suite.ctx, validator)
	suite.app.StakingKeeper.SetValidatorByPowerIndex(suite.ctx, validator)
	suite.app.StakingKeeper.AfterValidatorCreated(suite.ctx, validator.GetOperator())
	return valAddr
}

func (suite *KeeperTestSuite) SetupSuperfluidDelegate(valAddr sdk.ValAddress, denom string) lockuptypes.PeriodLock {
	suite.app.IncentivesKeeper.SetLockableDurations(suite.ctx, []time.Duration{
		time.Hour * 24 * 14,
		time.Hour,
		time.Hour * 3,
		time.Hour * 7,
	})

	// register a LP token as a superfluid asset
	suite.app.SuperfluidKeeper.SetSuperfluidAsset(suite.ctx, types.SuperfluidAsset{
		Denom:     denom,
		AssetType: types.SuperfluidAssetTypeLPShare,
	})

	// set OSMO TWAP price for LP token
	suite.app.SuperfluidKeeper.SetEpochOsmoEquivalentTWAP(suite.ctx, 1, denom, sdk.NewDec(20))
	params := suite.app.SuperfluidKeeper.GetParams(suite.ctx)
	suite.app.EpochsKeeper.SetEpochInfo(suite.ctx, epochstypes.EpochInfo{
		Identifier:   params.RefreshEpochIdentifier,
		CurrentEpoch: 2,
	})

	// create lockup of LP token
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin(denom, 1000000)}
	lock := suite.LockTokens(addr1, coins, time.Hour*24*21)

	// call SuperfluidDelegate and check response
	err := suite.app.SuperfluidKeeper.SuperfluidDelegate(suite.ctx, lock.ID, valAddr.String())
	suite.Require().NoError(err)

	return lock
}

func (suite *KeeperTestSuite) TestSuperfluidDelegate() {
	type superfluidDelegation struct {
		valIndex int64
		lpDenom  string
	}
	testCases := []struct {
		name             string
		validatorStats   []stakingtypes.BondStatus
		superDelegations []superfluidDelegation
	}{
		{
			"happy path with single validator and delegator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}},
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()

			poolId := suite.createGammPool([]string{appparams.BaseCoinUnit, "foo"})
			suite.Require().Equal(poolId, uint64(1))

			// setup validators
			valAddrs := []sdk.ValAddress{}
			for _, status := range tc.validatorStats {
				valAddr := suite.SetupValidator(status)
				valAddrs = append(valAddrs, valAddr)
			}

			// setup superfluid delegations
			for _, del := range tc.superDelegations {
				valAddr := valAddrs[del.valIndex]
				lock := suite.SetupSuperfluidDelegate(valAddr, del.lpDenom)

				// check synthetic lockup creation
				synthLock, err := suite.app.LockupKeeper.GetSyntheticLockup(suite.ctx, lock.ID, keeper.StakingSuffix(valAddr.String()))
				suite.Require().NoError(err)
				suite.Require().Equal(synthLock.UnderlyingLockId, lock.ID)
				suite.Require().Equal(synthLock.Suffix, keeper.StakingSuffix(valAddr.String()))
				suite.Require().Equal(synthLock.EndTime, time.Time{})

				// check intermediary account creation
				expAcc := types.SuperfluidIntermediaryAccount{
					Denom:   lock.Coins[0].Denom,
					ValAddr: valAddr.String(),
				}
				gotAcc := suite.app.SuperfluidKeeper.GetIntermediaryAccount(suite.ctx, expAcc.GetAddress())
				suite.Require().Equal(gotAcc.Denom, expAcc.Denom)
				suite.Require().Equal(gotAcc.ValAddr, expAcc.ValAddr)
				suite.Require().GreaterOrEqual(gotAcc.GaugeId, uint64(1))

				// check gauge creation
				gauge, err := suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gotAcc.GaugeId)
				suite.Require().NoError(err)
				suite.Require().Equal(gauge.Id, gotAcc.GaugeId)
				suite.Require().Equal(gauge.IsPerpetual, true)
				suite.Require().Equal(gauge.DistributeTo, lockuptypes.QueryCondition{
					LockQueryType: lockuptypes.ByDuration,
					Denom:         expAcc.Denom + keeper.StakingSuffix(valAddr.String()),
					Duration:      time.Hour * 24 * 14,
				})
				suite.Require().Equal(gauge.Coins, sdk.Coins(nil))
				suite.Require().Equal(gauge.StartTime, suite.ctx.BlockTime())
				suite.Require().Equal(gauge.NumEpochsPaidOver, uint64(1))
				suite.Require().Equal(gauge.FilledEpochs, uint64(0))
				suite.Require().Equal(gauge.DistributedCoins, sdk.Coins(nil))

				// Check lockID connection with intermediary account
				intAcc := suite.app.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(suite.ctx, lock.ID)
				suite.Require().Equal(intAcc.String(), expAcc.GetAddress().String())

				// check delegation from intermediary account to validator
				delegation, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, expAcc.GetAddress(), valAddr)
				suite.Require().True(found)
				suite.Require().Equal(delegation.Shares, sdk.NewDec(19000000)) // 95% x 2 x 1000000
			}
		})
	}
}

func (suite *KeeperTestSuite) TestSuperfluidUndelegate() {
	type superfluidDelegation struct {
		valIndex int64
		lpDenom  string
	}
	testCases := []struct {
		name                  string
		validatorStats        []stakingtypes.BondStatus
		superDelegations      []superfluidDelegation
		superUnbondingLockIds []uint64
	}{
		{
			"happy path with single validator and delegator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}},
			[]uint64{1},
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()

			poolId := suite.createGammPool([]string{appparams.BaseCoinUnit, "foo"})
			suite.Require().Equal(poolId, uint64(1))

			// setup validators
			valAddrs := []sdk.ValAddress{}
			for _, status := range tc.validatorStats {
				valAddr := suite.SetupValidator(status)
				valAddrs = append(valAddrs, valAddr)
			}

			// setup superfluid delegations
			for _, del := range tc.superDelegations {
				valAddr := valAddrs[del.valIndex]
				suite.SetupSuperfluidDelegate(valAddr, del.lpDenom)
			}

			for _, lockId := range tc.superUnbondingLockIds {
				// superfluid undelegate
				err := suite.app.SuperfluidKeeper.SuperfluidUndelegate(suite.ctx, lockId)
				suite.Require().NoError(err)

				// get intermediary account
				accAddr := suite.app.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(suite.ctx, lockId)
				intermediaryAcc := suite.app.SuperfluidKeeper.GetIntermediaryAccount(suite.ctx, accAddr)
				valAddr := intermediaryAcc.ValAddr

				// check bonding synthetic lockup deletion
				_, err = suite.app.LockupKeeper.GetSyntheticLockup(suite.ctx, lockId, keeper.StakingSuffix(valAddr))
				suite.Require().Error(err)

				// check unbonding synthetic lockup creation
				synthLock, err := suite.app.LockupKeeper.GetSyntheticLockup(suite.ctx, lockId, keeper.UntakingSuffix(valAddr))
				suite.Require().NoError(err)
				suite.Require().Equal(synthLock.UnderlyingLockId, lockId)
				suite.Require().Equal(synthLock.Suffix, keeper.UntakingSuffix(valAddr))
				suite.Require().Equal(synthLock.EndTime, suite.ctx.BlockTime().Add(time.Hour*24*21))
			}
		})
	}
}

func (suite *KeeperTestSuite) TestSuperfluidRedelegate() {
	type superfluidDelegation struct {
		valIndex int64
		lpDenom  string
	}
	type superfluidRedelegation struct {
		lockId      uint64
		oldValIndex int64
		newValIndex int64
	}
	testCases := []struct {
		name               string
		validatorStats     []stakingtypes.BondStatus
		superDelegations   []superfluidDelegation
		superRedelegations []superfluidRedelegation
	}{
		{
			"happy path with single validator and delegator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}},
			[]superfluidRedelegation{{1, 0, 1}},
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()

			poolId := suite.createGammPool([]string{appparams.BaseCoinUnit, "foo"})
			suite.Require().Equal(poolId, uint64(1))

			// setup validators
			valAddrs := []sdk.ValAddress{}
			for _, status := range tc.validatorStats {
				valAddr := suite.SetupValidator(status)
				valAddrs = append(valAddrs, valAddr)
			}

			// setup superfluid delegations
			for _, del := range tc.superDelegations {
				valAddr := valAddrs[del.valIndex]
				suite.SetupSuperfluidDelegate(valAddr, del.lpDenom)
			}

			// execute redelegation and check changes on store
			for _, srd := range tc.superRedelegations {
				// superfluid redelegate
				err := suite.app.SuperfluidKeeper.SuperfluidRedelegate(suite.ctx, srd.lockId, valAddrs[srd.newValIndex].String())
				suite.Require().NoError(err)

				// check previous validator bonding synthetic lockup deletion
				_, err = suite.app.LockupKeeper.GetSyntheticLockup(suite.ctx, srd.lockId, keeper.StakingSuffix(valAddrs[srd.oldValIndex].String()))
				suite.Require().Error(err)

				// check unbonding synthetic lockup creation
				synthLock, err := suite.app.LockupKeeper.GetSyntheticLockup(suite.ctx, srd.lockId, keeper.UntakingSuffix(valAddrs[srd.oldValIndex].String()))
				suite.Require().NoError(err)
				suite.Require().Equal(synthLock.UnderlyingLockId, srd.lockId)
				suite.Require().Equal(synthLock.Suffix, keeper.UntakingSuffix(valAddrs[srd.oldValIndex].String()))
				suite.Require().Equal(synthLock.EndTime, suite.ctx.BlockTime().Add(time.Hour*24*21))

				// check synthetic lockup creation
				synthLock2, err := suite.app.LockupKeeper.GetSyntheticLockup(suite.ctx, srd.lockId, keeper.StakingSuffix(valAddrs[srd.newValIndex].String()))
				suite.Require().NoError(err)
				suite.Require().Equal(synthLock2.UnderlyingLockId, srd.lockId)
				suite.Require().Equal(synthLock2.Suffix, keeper.StakingSuffix(valAddrs[srd.newValIndex].String()))
				suite.Require().Equal(synthLock2.EndTime, time.Time{})

				// check intermediary account creation
				lock, err := suite.app.LockupKeeper.GetLockByID(suite.ctx, srd.lockId)
				suite.Require().NoError(err)
				expAcc := types.SuperfluidIntermediaryAccount{
					Denom:   lock.Coins[0].Denom,
					ValAddr: valAddrs[srd.newValIndex].String(),
				}
				gotAcc := suite.app.SuperfluidKeeper.GetIntermediaryAccount(suite.ctx, expAcc.GetAddress())
				suite.Require().Equal(gotAcc.Denom, expAcc.Denom)
				suite.Require().Equal(gotAcc.ValAddr, expAcc.ValAddr)
				suite.Require().Greater(gotAcc.GaugeId, uint64(1))

				// check gauge creation
				gauge, err := suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gotAcc.GaugeId)
				suite.Require().NoError(err)
				suite.Require().Equal(gauge.Id, gotAcc.GaugeId)
				suite.Require().Equal(gauge.IsPerpetual, true)
				suite.Require().Equal(gauge.DistributeTo, lockuptypes.QueryCondition{
					LockQueryType: lockuptypes.ByDuration,
					Denom:         expAcc.Denom + keeper.StakingSuffix(valAddrs[srd.newValIndex].String()),
					Duration:      time.Hour * 24 * 14,
				})
				suite.Require().Equal(gauge.Coins, sdk.Coins(nil))
				suite.Require().Equal(gauge.StartTime, suite.ctx.BlockTime())
				suite.Require().Equal(gauge.NumEpochsPaidOver, uint64(1))
				suite.Require().Equal(gauge.FilledEpochs, uint64(0))
				suite.Require().Equal(gauge.DistributedCoins, sdk.Coins(nil))

				// Check lockID connection with intermediary account
				intAcc := suite.app.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(suite.ctx, srd.lockId)
				suite.Require().Equal(intAcc.String(), expAcc.GetAddress().String())

				// check delegation from intermediary account to validator
				_, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, expAcc.GetAddress(), valAddrs[srd.newValIndex])
				suite.Require().True(found)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestRefreshIntermediaryDelegationAmounts() {
	type superfluidDelegation struct {
		valIndex int64
		lpDenom  string
	}

	testCases := []struct {
		name             string
		validatorStats   []stakingtypes.BondStatus
		superDelegations []superfluidDelegation
		checkAccIndexes  []int64
	}{
		{
			"happy path with single validator and delegator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}},
			[]int64{0},
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()

			params := suite.app.SuperfluidKeeper.GetParams(suite.ctx)
			poolId := suite.createGammPool([]string{appparams.BaseCoinUnit, "foo"})
			suite.Require().Equal(poolId, uint64(1))

			// setup validators
			valAddrs := []sdk.ValAddress{}
			for _, status := range tc.validatorStats {
				valAddr := suite.SetupValidator(status)
				valAddrs = append(valAddrs, valAddr)
			}

			// setup superfluid delegations
			intAccs := []types.SuperfluidIntermediaryAccount{}
			locks := []lockuptypes.PeriodLock{}
			for _, del := range tc.superDelegations {
				valAddr := valAddrs[del.valIndex]
				lock := suite.SetupSuperfluidDelegate(valAddr, del.lpDenom)
				locks = append(locks, lock)
				intAccs = append(intAccs, types.SuperfluidIntermediaryAccount{
					Denom:   lock.Coins[0].Denom,
					ValAddr: valAddr.String(),
				})
			}

			for _, intAccIndex := range tc.checkAccIndexes {
				expAcc := intAccs[intAccIndex]
				valAddr, err := sdk.ValAddressFromBech32(expAcc.ValAddr)
				suite.Require().NoError(err)

				// check delegation from intermediary account to validator
				delegation, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, expAcc.GetAddress(), valAddr)
				suite.Require().True(found)
				suite.Require().Equal(delegation.Shares, sdk.NewDec(19000000)) // 95% x 20 x 1000000

				// twap price change before refresh
				suite.app.SuperfluidKeeper.SetEpochOsmoEquivalentTWAP(suite.ctx, 2, "gamm/pool/1", sdk.NewDec(10))
				suite.app.EpochsKeeper.SetEpochInfo(suite.ctx, epochstypes.EpochInfo{
					Identifier:   params.RefreshEpochIdentifier,
					CurrentEpoch: 2,
				})

				// refresh intermediary account delegations
				suite.NotPanics(func() {
					suite.app.SuperfluidKeeper.RefreshIntermediaryDelegationAmounts(suite.ctx)
				})

				// check delegation changes
				delegation, found = suite.app.StakingKeeper.GetDelegation(suite.ctx, expAcc.GetAddress(), valAddr)
				suite.Require().True(found)
				suite.Require().Equal(delegation.Shares, sdk.NewDec(9500000)) // 95% x 10 x 1000000

			}

			// start new epoch
			suite.app.EpochsKeeper.SetEpochInfo(suite.ctx, epochstypes.EpochInfo{
				Identifier:   params.RefreshEpochIdentifier,
				CurrentEpoch: 3,
			})

			// unbond all lockups
			for _, lock := range locks {
				// superfluid undelegate
				err := suite.app.SuperfluidKeeper.SuperfluidUndelegate(suite.ctx, lock.ID)
				suite.Require().NoError(err)
			}

			// check intermediary account changes after unbonding operations
			for _, intAccIndex := range tc.checkAccIndexes {
				expAcc := intAccs[intAccIndex]
				suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Hour*24*21 + time.Second))
				suite.app.EndBlocker(suite.ctx, abci.RequestEndBlock{Height: suite.ctx.BlockHeight()})

				unbonded := suite.app.BankKeeper.GetBalance(suite.ctx, expAcc.GetAddress(), sdk.DefaultBondDenom)
				suite.Require().True(unbonded.IsPositive())
			}

			// refresh intermediary account delegations
			suite.NotPanics(func() {
				suite.app.SuperfluidKeeper.RefreshIntermediaryDelegationAmounts(suite.ctx)
			})

			// check changes after refresh operation
			for _, intAccIndex := range tc.checkAccIndexes {
				expAcc := intAccs[intAccIndex]
				// check unbonded amount is removed after refresh operation
				refreshed := suite.app.BankKeeper.GetBalance(suite.ctx, expAcc.GetAddress(), sdk.DefaultBondDenom)
				suite.Require().True(refreshed.IsZero())
			}
		})
	}
}
