package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	epochstypes "github.com/osmosis-labs/osmosis/v7/x/epochs/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v7/x/lockup/types"
	minttypes "github.com/osmosis-labs/osmosis/v7/x/mint/types"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
)

type superfluidDelegation struct {
	delIndex int64
	valIndex int64
	lpDenom  string
	lpAmount int64
}
type superfluidRedelegation struct {
	lockId      uint64
	oldValIndex int64
	newValIndex int64
}

type assetTwap struct {
	denom string
	price sdk.Dec
}

type osmoEquivalentMultiplier struct {
	denom string
	price sdk.Dec
}

func (suite *KeeperTestSuite) SetupSuperfluidDelegations(delAddrs []sdk.AccAddress, valAddrs []sdk.ValAddress, superDelegations []superfluidDelegation) ([]types.SuperfluidIntermediaryAccount, []lockuptypes.PeriodLock) {
	flagIntermediaryAcc := make(map[string]bool)
	intermediaryAccs := []types.SuperfluidIntermediaryAccount{}
	locks := []lockuptypes.PeriodLock{}

	// setup superfluid delegations

	// we do sanity check on the test cases.
	// if superfluid staking is happening with single val and multiple superfluid delegations,
	// we should be running `AddTokensToLockByID`, instead of creating new locks
	for _, del := range superDelegations {
		delAddr := delAddrs[del.delIndex]
		valAddr := valAddrs[del.valIndex]
		lock := suite.SetupSuperfluidDelegate(delAddr, valAddr, del.lpDenom, del.lpAmount)
		expAcc := types.NewSuperfluidIntermediaryAccount(lock.Coins[0].Denom, valAddr.String(), 0)

		// save accounts for future use
		if flagIntermediaryAcc[expAcc.String()] == false {
			flagIntermediaryAcc[expAcc.String()] = true
			intermediaryAccs = append(intermediaryAccs, expAcc)
		}
		// save locks for future use
		locks = append(locks, lock)
	}
	return intermediaryAccs, locks
}

func (suite *KeeperTestSuite) checkIntermediaryAccountDelegations(intermediaryAccs []types.SuperfluidIntermediaryAccount) {
	for _, acc := range intermediaryAccs {
		valAddr, err := sdk.ValAddressFromBech32(acc.ValAddr)
		suite.Require().NoError(err)

		// check delegation from intermediary account to validator
		delegation, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, acc.GetAccAddress(), valAddr)
		suite.Require().True(found)
		suite.Require().True(delegation.Shares.GTE(sdk.NewDec(10000000)))

		// check delegated tokens
		validator, found := suite.app.StakingKeeper.GetValidator(suite.ctx, valAddr)
		suite.Require().True(found)
		delegatedTokens := validator.TokensFromShares(delegation.Shares).TruncateInt()
		suite.Require().True(delegatedTokens.GTE(sdk.NewInt(10000000)))
	}
}

func (suite *KeeperTestSuite) SetupSuperfluidDelegate(delAddr sdk.AccAddress, valAddr sdk.ValAddress, denom string, amount int64) lockuptypes.PeriodLock {
	unbondingDuration := suite.app.StakingKeeper.GetParams(suite.ctx).UnbondingTime

	// register a LP token as a superfluid asset
	suite.app.SuperfluidKeeper.SetSuperfluidAsset(suite.ctx, types.SuperfluidAsset{
		Denom:     denom,
		AssetType: types.SuperfluidAssetTypeLPShare,
	})

	// set OSMO TWAP price for LP token
	suite.app.SuperfluidKeeper.SetOsmoEquivalentMultiplier(suite.ctx, 1, denom, sdk.NewDec(20))
	epochIdentifier := suite.app.SuperfluidKeeper.GetEpochIdentifier(suite.ctx)
	suite.app.EpochsKeeper.SetEpochInfo(suite.ctx, epochstypes.EpochInfo{
		Identifier:   epochIdentifier,
		CurrentEpoch: 2,
	})

	// create lockup of LP token
	coins := sdk.Coins{sdk.NewInt64Coin(denom, amount)}
	lastLockID := suite.app.LockupKeeper.GetLastLockID(suite.ctx)

	lockID := suite.LockTokens(delAddr, coins, unbondingDuration)
	lock, err := suite.app.LockupKeeper.GetLockByID(suite.ctx, lockID)
	suite.Require().NoError(err)

	// here we check if check `LockTokens` added to existing locks or created a new lock.
	// if `LockTokens` created a new lock, we continue SuperfluidDelegate
	// if lock has been existing before, we wouldn't have to call SuperfluidDelegate separately, as hooks on LockTokens would have automatically called SuperfluidDelegateMore
	if lastLockID != lockID {
		err = suite.app.SuperfluidKeeper.SuperfluidDelegate(suite.ctx, lock.Owner, lock.ID, valAddr.String())
		suite.Require().NoError(err)
	} else {
		// here we handle two cases.
		// 1. the lock has existed before but has not been superflud staking
		// 2. the lock has existed before and has been superfluid staking

		// we check if synth lock that has existed before, if it did, it means that the lock has been superfluid staked
		// we do not care about unbonding synthlocks, as superfluid delegation has no effect

		_, err := suite.app.LockupKeeper.GetSyntheticLockup(suite.ctx, lockID, keeper.StakingSyntheticDenom(lock.Coins[0].Denom, valAddr.String()))
		// if lock has existed before but has not been superfluid staked, we do initial superfluid staking
		if err != nil {
			err = suite.app.SuperfluidKeeper.SuperfluidDelegate(suite.ctx, lock.Owner, lock.ID, valAddr.String())
			suite.Require().NoError(err)
		}
	}

	return *lock
}

func (suite *KeeperTestSuite) TestSuperfluidDelegate() {
	testCases := []struct {
		name               string
		validatorStats     []stakingtypes.BondStatus
		delegatorNumber    int
		superDelegations   []superfluidDelegation
		expInterDelegation []sdk.Dec
	}{
		{
			"with single validator and single superfluid delegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			1,
			[]superfluidDelegation{{0, 0, "gamm/pool/1", 1000000}},
			[]sdk.Dec{sdk.NewDec(10000000)}, // 50% x 20 x 1000000
		},
		{
			"with single validator and additional superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			1,
			[]superfluidDelegation{{0, 0, "gamm/pool/1", 1000000}, {0, 0, "gamm/pool/1", 1000000}},
			[]sdk.Dec{sdk.NewDec(20000000)}, // 50% x 20 x 1000000 x 2
		},
		{
			"with multiples validator and multiple superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			2,
			[]superfluidDelegation{{0, 0, "gamm/pool/1", 1000000}, {1, 1, "gamm/pool/1", 1000000}},
			[]sdk.Dec{sdk.NewDec(10000000), sdk.NewDec(10000000)}, // 50% x 20 x 1000000
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()
			bondDenom := suite.app.StakingKeeper.BondDenom(suite.ctx)

			// Generate delegator addresses
			delAddrs := CreateRandomAccounts(tc.delegatorNumber)

			// setup validators
			valAddrs := suite.SetupValidators(tc.validatorStats)

			// get pre-superfluid delgations osmo supply and supplyWithOffset
			presupply := suite.app.BankKeeper.GetSupply(suite.ctx, bondDenom)
			presupplyWithOffset := suite.app.BankKeeper.GetSupplyWithOffset(suite.ctx, bondDenom)

			// setup superfluid delegations
			intermediaryAccs, locks := suite.SetupSuperfluidDelegations(delAddrs, valAddrs, tc.superDelegations)

			// ensure post-superfluid delegations osmo supplywithoffset is the same while supply is not
			postsupply := suite.app.BankKeeper.GetSupply(suite.ctx, bondDenom)
			postsupplyWithOffset := suite.app.BankKeeper.GetSupplyWithOffset(suite.ctx, bondDenom)
			suite.Require().False(postsupply.IsEqual(presupply), "presupply: %s   postsupply: %s", presupply, postsupply)
			suite.Require().True(postsupplyWithOffset.IsEqual(presupplyWithOffset))

			unbondingDuration := suite.app.StakingKeeper.GetParams(suite.ctx).UnbondingTime

			for index, del := range tc.superDelegations {
				lock := locks[index]
				valAddr := valAddrs[del.valIndex]

				// check synthetic lockup creation
				synthLock, err := suite.app.LockupKeeper.GetSyntheticLockup(suite.ctx, lock.ID, keeper.StakingSyntheticDenom(lock.Coins[0].Denom, valAddr.String()))
				suite.Require().NoError(err)
				suite.Require().Equal(synthLock.UnderlyingLockId, lock.ID)
				suite.Require().Equal(synthLock.SynthDenom, keeper.StakingSyntheticDenom(lock.Coins[0].Denom, valAddr.String()))
				suite.Require().Equal(synthLock.EndTime, time.Time{})

				expAcc := types.NewSuperfluidIntermediaryAccount(lock.Coins[0].Denom, valAddr.String(), 0)

				// Check lockID connection with intermediary account
				intAcc := suite.app.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(suite.ctx, lock.ID)
				suite.Require().Equal(intAcc.String(), expAcc.GetAccAddress().String())
			}

			for index, expAcc := range intermediaryAccs {
				// check intermediary account creation
				gotAcc := suite.app.SuperfluidKeeper.GetIntermediaryAccount(suite.ctx, expAcc.GetAccAddress())
				suite.Require().Equal(gotAcc.Denom, expAcc.Denom)
				suite.Require().Equal(gotAcc.ValAddr, expAcc.ValAddr)
				suite.Require().GreaterOrEqual(gotAcc.GaugeId, uint64(1))

				valAddr, err := sdk.ValAddressFromBech32(expAcc.ValAddr)
				suite.Require().NoError(err)

				// check gauge creation
				gauge, err := suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gotAcc.GaugeId)
				suite.Require().NoError(err)
				suite.Require().Equal(gauge.Id, gotAcc.GaugeId)
				suite.Require().Equal(gauge.IsPerpetual, true)
				suite.Require().Equal(gauge.DistributeTo, lockuptypes.QueryCondition{
					LockQueryType: lockuptypes.ByDuration,
					Denom:         keeper.StakingSyntheticDenom(expAcc.Denom, valAddr.String()),
					Duration:      unbondingDuration,
				})
				suite.Require().Equal(gauge.Coins, sdk.Coins(nil))
				suite.Require().Equal(gauge.StartTime, suite.ctx.BlockTime())
				suite.Require().Equal(gauge.NumEpochsPaidOver, uint64(1))
				suite.Require().Equal(gauge.FilledEpochs, uint64(0))
				suite.Require().Equal(gauge.DistributedCoins, sdk.Coins(nil))

				// check delegation from intermediary account to validator
				delegation, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, expAcc.GetAccAddress(), valAddr)
				suite.Require().True(found)
				suite.Require().Equal(tc.expInterDelegation[index], delegation.Shares)
			}

			// try delegating twice with same lockup
			for _, lock := range locks {
				err := suite.app.SuperfluidKeeper.SuperfluidDelegate(suite.ctx, lock.Owner, lock.ID, valAddrs[0].String())
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestSuperfluidUndelegate() {
	testCases := []struct {
		name                  string
		validatorStats        []stakingtypes.BondStatus
		delegatorNumber       int
		superDelegations      []superfluidDelegation
		addMoreTokensLockIds  []uint64
		superUnbondingLockIds []uint64
		expSuperUnbondingErr  []bool
		// expected amount of delegation to intermediary account
		expInterDelegation []sdk.Dec
	}{
		{
			"with single validator and single superfluid delegation and single undelegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			1,
			[]superfluidDelegation{{0, 0, "gamm/pool/1", 1000000}},
			[]uint64{},
			[]uint64{1},
			[]bool{false},
			[]sdk.Dec{sdk.ZeroDec()},
		},
		// {
		// 	"with single validator, single superfluid delegation, add more tokens to the lock, and single undelegation",
		// 	[]stakingtypes.BondStatus{stakingtypes.Bonded},
		// 	[]superfluidDelegation{{0, "gamm/pool/1", 1000000}},
		// 	[]uint64{1},
		// 	[]uint64{1},
		// 	[]bool{false},
		// 	[]sdk.Dec{sdk.ZeroDec()},
		// },
		{
			"with single validator and additional superfluid delegations and single undelegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			1,
			[]superfluidDelegation{{0, 0, "gamm/pool/1", 1000000}, {0, 0, "gamm/pool/1", 1000000}},
			[]uint64{},
			[]uint64{1},
			[]bool{false},
			[]sdk.Dec{sdk.ZeroDec()},
		},
		{
			"with multiple validators and multiple superfluid delegations and multiple undelegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			2,
			[]superfluidDelegation{{0, 0, "gamm/pool/1", 1000000}, {1, 1, "gamm/pool/1", 1000000}},
			[]uint64{},
			[]uint64{1, 2},
			[]bool{false, false},
			[]sdk.Dec{sdk.ZeroDec()},
		},
		{
			"undelegating not available lock id",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			1,
			[]superfluidDelegation{{0, 0, "gamm/pool/1", 1000000}},
			[]uint64{},
			[]uint64{2},
			[]bool{true},
			[]sdk.Dec{sdk.NewDec(10000000)},
		},
		{
			"try undelegating twice for same lock id",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			1,
			[]superfluidDelegation{{0, 0, "gamm/pool/1", 1000000}},
			[]uint64{},
			[]uint64{1, 1},
			[]bool{false, true},
			[]sdk.Dec{sdk.ZeroDec()},
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()

			bondDenom := suite.app.StakingKeeper.GetParams(suite.ctx).BondDenom

			// Generate delegator addresses
			delAddrs := CreateRandomAccounts(tc.delegatorNumber)

			// setup validators
			valAddrs := suite.SetupValidators(tc.validatorStats)

			// setup superfluid delegations
			intermediaryAccs, _ := suite.SetupSuperfluidDelegations(delAddrs, valAddrs, tc.superDelegations)
			suite.checkIntermediaryAccountDelegations(intermediaryAccs)

			for _, lockId := range tc.addMoreTokensLockIds {
				lock, err := suite.app.LockupKeeper.GetLockByID(suite.ctx, lockId)
				suite.Require().NoError(err)
				lockOwner, err := sdk.AccAddressFromBech32(lock.Owner)
				suite.Require().NoError(err)
				coins := sdk.Coins{sdk.NewInt64Coin("gamm/pool/1", 1000000)}
				suite.app.BankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, coins)
				suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, minttypes.ModuleName, lockOwner, coins)
				_, err = suite.app.LockupKeeper.AddTokensToLockByID(suite.ctx, lockId, coins)
				suite.Require().NoError(err)
			}

			for index, lockId := range tc.superUnbondingLockIds {
				// get intermediary account
				accAddr := suite.app.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(suite.ctx, lockId)
				intermediaryAcc := suite.app.SuperfluidKeeper.GetIntermediaryAccount(suite.ctx, accAddr)
				valAddr := intermediaryAcc.ValAddr

				lock, err := suite.app.LockupKeeper.GetLockByID(suite.ctx, lockId)
				if err != nil {
					lock = &lockuptypes.PeriodLock{}
				}

				// get pre-superfluid delgations osmo supply and supplyWithOffset
				presupply := suite.app.BankKeeper.GetSupply(suite.ctx, bondDenom)
				presupplyWithOffset := suite.app.BankKeeper.GetSupplyWithOffset(suite.ctx, bondDenom)

				// superfluid undelegate
				err = suite.app.SuperfluidKeeper.SuperfluidUndelegate(suite.ctx, lock.Owner, lockId)
				if tc.expSuperUnbondingErr[index] {
					suite.Require().Error(err)
					continue
				}
				suite.Require().NoError(err)

				// ensure post-superfluid delegations osmo supplywithoffset is the same while supply is not
				postsupply := suite.app.BankKeeper.GetSupply(suite.ctx, bondDenom)
				postsupplyWithOffset := suite.app.BankKeeper.GetSupplyWithOffset(suite.ctx, bondDenom)
				suite.Require().False(postsupply.IsEqual(presupply), "presupply: %s   postsupply: %s", presupply, postsupply)
				suite.Require().True(postsupplyWithOffset.IsEqual(presupplyWithOffset))

				// check lockId and intermediary account connection deletion
				addr := suite.app.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(suite.ctx, lockId)
				suite.Require().Equal(addr.String(), "")

				// check bonding synthetic lockup deletion
				_, err = suite.app.LockupKeeper.GetSyntheticLockup(suite.ctx, lockId, keeper.StakingSyntheticDenom(lock.Coins[0].Denom, valAddr))
				suite.Require().Error(err)

				// check unbonding synthetic lockup creation
				unbondingDuration := suite.app.StakingKeeper.GetParams(suite.ctx).UnbondingTime
				synthLock, err := suite.app.LockupKeeper.GetSyntheticLockup(suite.ctx, lockId, keeper.UnstakingSyntheticDenom(lock.Coins[0].Denom, valAddr))
				suite.Require().NoError(err)
				suite.Require().Equal(synthLock.UnderlyingLockId, lockId)
				suite.Require().Equal(synthLock.SynthDenom, keeper.UnstakingSyntheticDenom(lock.Coins[0].Denom, valAddr))
				suite.Require().Equal(synthLock.EndTime, suite.ctx.BlockTime().Add(unbondingDuration))
			}

			// check remaining intermediary account delegation
			for index, expDelegation := range tc.expInterDelegation {
				acc := intermediaryAccs[index]
				valAddr, err := sdk.ValAddressFromBech32(acc.ValAddr)
				suite.Require().NoError(err)
				delegation, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, acc.GetAccAddress(), valAddr)
				if expDelegation.IsZero() {
					suite.Require().False(found, "expected no delegation, found delegation w/ %d shares", delegation.Shares)
				} else {
					suite.Require().True(found)
					suite.Require().Equal(expDelegation, delegation.Shares)
				}
			}

			// try undelegating twice
			for index, lockId := range tc.superUnbondingLockIds {
				if tc.expSuperUnbondingErr[index] {
					continue
				}

				lock, err := suite.app.LockupKeeper.GetLockByID(suite.ctx, lockId)
				suite.Require().NoError(err)

				err = suite.app.SuperfluidKeeper.SuperfluidUndelegate(suite.ctx, lock.Owner, lockId)
				suite.Require().Error(err)
			}
		})
	}
}

// TestSuperfluidUnbondLock tests the following.
// 		1. test SuperfluidUnbondLock does not work before undelegation
// 		2. test SuperfluidUnbondLock makes underlying lock start unlocking
// 		3. test that synthetic lockup being finished does not mean underlying lock is finished
//      4. test after SuperfluidUnbondLock + lockup time, the underlying lock is finished
func (suite *KeeperTestSuite) TestSuperfluidUnbondLock() {
	suite.SetupTest()

	// Generate delegator addresses
	delAddrs := CreateRandomAccounts(1)

	// setup validators
	valAddrs := suite.SetupValidators([]stakingtypes.BondStatus{stakingtypes.Bonded})

	// setup superfluid delegations
	intermediaryAccs, locks := suite.SetupSuperfluidDelegations(delAddrs, valAddrs, []superfluidDelegation{{0, 0, "gamm/pool/1", 1000000}})
	suite.checkIntermediaryAccountDelegations(intermediaryAccs)

	for _, lock := range locks {
		startTime := time.Now()
		suite.ctx = suite.ctx.WithBlockTime(startTime)
		accAddr := suite.app.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(suite.ctx, lock.ID)
		intermediaryAcc := suite.app.SuperfluidKeeper.GetIntermediaryAccount(suite.ctx, accAddr)
		valAddr := intermediaryAcc.ValAddr

		// first we test that SuperfluidUnbondLock would cause error before undelegating
		err := suite.app.SuperfluidKeeper.SuperfluidUnbondLock(suite.ctx, lock.ID, lock.GetOwner())
		suite.Require().Error(err)

		// undelegation needs to happen prior to SuperfluidUnbondLock
		err = suite.app.SuperfluidKeeper.SuperfluidUndelegate(suite.ctx, lock.Owner, lock.ID)
		suite.Require().NoError(err)
		balances := suite.app.BankKeeper.GetAllBalances(suite.ctx, lock.OwnerAddress())
		suite.Require().Equal(0, balances.Len())

		// check that unbonding synth has been created correctly after undelegation
		unbondingDuration := suite.app.StakingKeeper.GetParams(suite.ctx).UnbondingTime
		synthLock, err := suite.app.LockupKeeper.GetSyntheticLockup(suite.ctx, lock.ID, keeper.UnstakingSyntheticDenom(lock.Coins[0].Denom, valAddr))
		suite.Require().NoError(err)
		suite.Require().Equal(synthLock.UnderlyingLockId, lock.ID)
		suite.Require().Equal(synthLock.SynthDenom, keeper.UnstakingSyntheticDenom(lock.Coins[0].Denom, valAddr))
		suite.Require().Equal(synthLock.EndTime, suite.ctx.BlockTime().Add(unbondingDuration))

		// test SuperfluidUnbondLock
		unbondLockStartTime := startTime.Add(time.Hour)
		suite.ctx = suite.ctx.WithBlockTime(unbondLockStartTime)
		err = suite.app.SuperfluidKeeper.SuperfluidUnbondLock(suite.ctx, lock.ID, lock.GetOwner())
		suite.Require().NoError(err)

		// check that SuperfluidUnbondLock makes underlying lock start unlocking
		// we run WithdrawAllMaturedLocks to ensure that lock isn't getting finished immediately
		suite.app.LockupKeeper.WithdrawAllMaturedLocks(suite.ctx)
		updatedLock, err := suite.app.LockupKeeper.GetLockByID(suite.ctx, lock.ID)
		suite.Require().NoError(err)
		suite.Require().True(updatedLock.IsUnlocking())

		// check if finsihed unlocking synth lock did not increase balance
		balances = suite.app.BankKeeper.GetAllBalances(suite.ctx, lock.OwnerAddress())
		suite.Require().Equal(0, balances.Len())

		// test that synth lock finish does not mean underlying lock is finished
		suite.ctx = suite.ctx.WithBlockTime((startTime.Add(unbondingDuration)))
		suite.app.LockupKeeper.DeleteAllMaturedSyntheticLocks(suite.ctx)
		suite.app.LockupKeeper.WithdrawAllMaturedLocks(suite.ctx)
		_, err = suite.app.LockupKeeper.GetSyntheticLockup(suite.ctx, lock.ID, keeper.UnstakingSyntheticDenom(lock.Coins[0].Denom, valAddr))
		suite.Require().Error(err)
		updatedLock, err = suite.app.LockupKeeper.GetLockByID(suite.ctx, lock.ID)
		suite.Require().NoError(err)
		suite.Require().True(updatedLock.IsUnlocking())

		// test after SuperfluidUnbondLock + lockup unbonding duration, lock is finished and does not exist
		suite.ctx = suite.ctx.WithBlockTime(unbondLockStartTime.Add(unbondingDuration))
		suite.app.LockupKeeper.WithdrawAllMaturedLocks(suite.ctx)
		_, err = suite.app.LockupKeeper.GetLockByID(suite.ctx, lock.ID)
		suite.Require().Error(err)

		// check if finished unlocking succesfully increased balance
		balances = suite.app.BankKeeper.GetAllBalances(suite.ctx, lock.OwnerAddress())
		suite.Require().Equal(1, balances.Len())
		suite.Require().Equal("gamm/pool/1", balances[0].Denom)
		suite.Require().Equal(sdk.NewInt(1000000), balances[0].Amount)

	}
}

// func (suite *KeeperTestSuite) TestSuperfluidRedelegate() {
// 	testCases := []struct {
// 		name                    string
// 		validatorStats          []stakingtypes.BondStatus
// 		superDelegations        []superfluidDelegation
// 		superRedelegations      []superfluidRedelegation
// 		expSuperRedelegationErr []bool
// 	}{
// 		{
// 			"with single validator and single superfluid delegation with single redelegation",
// 			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
// 			[]superfluidDelegation{{0, "gamm/pool/1", 1000000}},
// 			[]superfluidRedelegation{{1, 0, 1}}, // lock1 => val0 -> val1
// 			[]bool{false},
// 		},
// 		{
// 			"with multiple superfluid delegations with single redelegation",
// 			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
// 			[]superfluidDelegation{{0, "gamm/pool/1", 1000000}, {0, "gamm/pool/1", 1000000}},
// 			[]superfluidRedelegation{{1, 0, 1}}, // lock1 => val0 -> val1
// 			[]bool{false},
// 		},
// 		{
// 			"with multiple superfluid delegations with multiple redelegations",
// 			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
// 			[]superfluidDelegation{{0, "gamm/pool/1", 1000000}, {0, "gamm/pool/1", 1000000}},
// 			[]superfluidRedelegation{{1, 0, 1}, {2, 0, 1}}, // lock1 => val0 -> val1, lock2 => val0 -> val1
// 			[]bool{false, false},
// 		},
// 		{
// 			"try redelegating back from new validator to original validator",
// 			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
// 			[]superfluidDelegation{{0, "gamm/pool/1", 1000000}, {0, "gamm/pool/1", 1000000}},
// 			[]superfluidRedelegation{{1, 0, 1}, {1, 1, 0}}, // lock1 => val0 -> val1, lock1 => val1 -> val0
// 			[]bool{false, true},
// 		},
// 		{
// 			"not available lock id redelegation",
// 			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
// 			[]superfluidDelegation{{0, "gamm/pool/1", 1000000}},
// 			[]superfluidRedelegation{{2, 0, 1}}, // lock1 => val0 -> val1
// 			[]bool{true},
// 		},
// 		{
// 			"redelegation for same validator",
// 			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
// 			[]superfluidDelegation{{0, "gamm/pool/1", 1000000}},
// 			[]superfluidRedelegation{{1, 0, 0}}, // lock1 => val0 -> val0
// 			[]bool{true},
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc
// 		suite.Run(tc.name, func() {
// 			suite.SetupTest()

// 			poolId := suite.createGammPool([]string{appparams.BaseCoinUnit, "foo"})
// 			suite.Require().Equal(poolId, uint64(1))

// 			// setup validators
// 			valAddrs := suite.SetupValidators(tc.validatorStats)

// 			// setup superfluid delegations
// 			intermediaryAccs, _ := suite.SetupSuperfluidDelegations(valAddrs, tc.superDelegations)
// 			suite.checkIntermediaryAccountDelegations(intermediaryAccs)

// 			// execute redelegation and check changes on store
// 			for index, srd := range tc.superRedelegations {
// 				lock, err := suite.app.LockupKeeper.GetLockByID(suite.ctx, srd.lockId)
// 				if err != nil {
// 					lock = &lockuptypes.PeriodLock{}
// 				}

// 				// superfluid redelegate
// 				err = suite.app.SuperfluidKeeper.SuperfluidRedelegate(suite.ctx, lock.Owner, srd.lockId, valAddrs[srd.newValIndex].String())
// 				if tc.expSuperRedelegationErr[index] {
// 					suite.Require().Error(err)
// 					continue
// 				}
// 				suite.Require().NoError(err)

// 				// check previous validator bonding synthetic lockup deletion
// 				_, err = suite.app.LockupKeeper.GetSyntheticLockup(suite.ctx, srd.lockId, keeper.StakingSuffix(valAddrs[srd.oldValIndex].String()))
// 				suite.Require().Error(err)

// 				// check unbonding synthetic lockup creation
// 				params := suite.app.SuperfluidKeeper.GetParams(suite.ctx)
// 				synthLock, err := suite.app.LockupKeeper.GetSyntheticLockup(suite.ctx, srd.lockId, keeper.UnstakingSuffix(valAddrs[srd.oldValIndex].String()))
// 				suite.Require().NoError(err)
// 				suite.Require().Equal(synthLock.UnderlyingLockId, srd.lockId)
// 				suite.Require().Equal(synthLock.Suffix, keeper.UnstakingSuffix(valAddrs[srd.oldValIndex].String()))
// 				suite.Require().Equal(synthLock.EndTime, suite.ctx.BlockTime().Add(params.UnbondingDuration))

// 				// check synthetic lockup creation
// 				synthLock2, err := suite.app.LockupKeeper.GetSyntheticLockup(suite.ctx, srd.lockId, keeper.StakingSuffix(valAddrs[srd.newValIndex].String()))
// 				suite.Require().NoError(err)
// 				suite.Require().Equal(synthLock2.UnderlyingLockId, srd.lockId)
// 				suite.Require().Equal(synthLock2.Suffix, keeper.StakingSuffix(valAddrs[srd.newValIndex].String()))
// 				suite.Require().Equal(synthLock2.EndTime, time.Time{})

// 				// check intermediary account creation
// 				lock, err = suite.app.LockupKeeper.GetLockByID(suite.ctx, srd.lockId)
// 				suite.Require().NoError(err)

// 				expAcc := types.NewSuperfluidIntermediaryAccount(lock.Coins[0].Denom, valAddrs[srd.newValIndex].String(), 1)
// 				gotAcc := suite.app.SuperfluidKeeper.GetIntermediaryAccount(suite.ctx, expAcc.GetAccAddress())
// 				suite.Require().Equal(gotAcc.Denom, expAcc.Denom)
// 				suite.Require().Equal(gotAcc.ValAddr, expAcc.ValAddr)

// 				// check gauge creation
// 				gauge, err := suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gotAcc.GaugeId)
// 				suite.Require().NoError(err)
// 				suite.Require().Equal(gauge.Id, gotAcc.GaugeId)
// 				suite.Require().Equal(gauge.IsPerpetual, true)
// 				suite.Require().Equal(gauge.DistributeTo, lockuptypes.QueryCondition{
// 					LockQueryType: lockuptypes.ByDuration,
// 					Denom:         expAcc.Denom + keeper.StakingSuffix(valAddrs[srd.newValIndex].String()),
// 					Duration:      params.UnbondingDuration,
// 				})
// 				suite.Require().Equal(gauge.Coins, sdk.Coins(nil))
// 				suite.Require().Equal(gauge.StartTime, suite.ctx.BlockTime())
// 				suite.Require().Equal(gauge.NumEpochsPaidOver, uint64(1))
// 				suite.Require().Equal(gauge.FilledEpochs, uint64(0))
// 				suite.Require().Equal(gauge.DistributedCoins, sdk.Coins(nil))

// 				// Check lockID connection with intermediary account
// 				intAcc := suite.app.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(suite.ctx, srd.lockId)
// 				suite.Require().Equal(intAcc.String(), expAcc.GetAccAddress().String())

// 				// check delegation from intermediary account to validator
// 				_, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, expAcc.GetAccAddress(), valAddrs[srd.newValIndex])
// 				suite.Require().True(found)
// 			}

// 			// try redelegating twice
// 			for index, srd := range tc.superRedelegations {
// 				if tc.expSuperRedelegationErr[index] {
// 					continue
// 				}
// 				cacheCtx, _ := suite.ctx.CacheContext()
// 				lock, err := suite.app.LockupKeeper.GetLockByID(suite.ctx, srd.lockId)
// 				suite.Require().NoError(err)
// 				err = suite.app.SuperfluidKeeper.SuperfluidRedelegate(cacheCtx, lock.Owner, srd.lockId, valAddrs[srd.newValIndex].String())
// 				suite.Require().Error(err)
// 			}
// 		})
// 	}
// }

// func (suite *KeeperTestSuite) TestRefreshIntermediaryDelegationAmounts1() {
// 	testCases := []struct {
// 		name             string
// 		validatorStats   []stakingtypes.BondStatus
// 		delegatorNumber  int
// 		superDelegations []superfluidDelegation
// 		roundOneTwaps    []assetTwap
// 		roundTwoTwaps    []assetTwap
// 		checkAccIndexes  []int64
// 	}{
// 		{
// 			"with single validator and single delegation",
// 			[]stakingtypes.BondStatus{stakingtypes.Bonded},
// 			1,
// 			[]superfluidDelegation{{0, 0, "gamm/pool/1", 1000000}},
// 			[]assetTwap{{"gamm/pool/1", sdk.NewDec(10)}},
// 			[]assetTwap{},
// 			[]int64{0},
// 		},
// 		{
// 			"with single validator and additional delegations",
// 			[]stakingtypes.BondStatus{stakingtypes.Bonded},
// 			2,
// 			[]superfluidDelegation{{0, 0, "gamm/pool/1", 1000000}, {0, 0, "gamm/pool/1", 1000000}},
// 			[]assetTwap{{"gamm/pool/1", sdk.NewDec(20)}},
// 			[]assetTwap{},
// 			[]int64{0},
// 		},
// 		{
// 			"with multiple validator and multiple superfluid delegations",
// 			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
// 			[]superfluidDelegation{{0, 0, "gamm/pool/1", 1000000}, {0, 1, "gamm/pool/1", 1000000}},
// 			[]assetTwap{{"gamm/pool/1", sdk.NewDec(10)}},
// 			[]assetTwap{},
// 			[]int64{0, 1},
// 		},
// 		{
// 			"with single validator and multiple denom superfluid delegations",
// 			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
// 			[]superfluidDelegation{{0, 0, "gamm/pool/1", 1000000}, {0, 0, "gamm/pool/2", 1000000}},
// 			[]assetTwap{{"gamm/pool/1", sdk.NewDec(10)}, {"gamm/pool/2", sdk.NewDec(10)}},
// 			[]assetTwap{},
// 			[]int64{0, 1},
// 		},
// 		{
// 			"with multiple validators and multiple denom superfluid delegations",
// 			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
// 			[]superfluidDelegation{{0, 0, "gamm/pool/1", 1000000}, {0, 1, "gamm/pool/2", 1000000}},
// 			[]assetTwap{{"gamm/pool/1", sdk.NewDec(10)}, {"gamm/pool/2", sdk.NewDec(10)}},
// 			[]assetTwap{},
// 			[]int64{0, 1},
// 		},
// 		{
// 			"zero price twap check",
// 			[]stakingtypes.BondStatus{stakingtypes.Bonded},
// 			[]superfluidDelegation{{0, 0, "gamm/pool/1", 1000000}},
// 			[]assetTwap{{"gamm/pool/1", sdk.NewDec(0)}},
// 			[]assetTwap{},
// 			[]int64{0},
// 		},
// 		{
// 			"refresh case from zero to non-zero",
// 			[]stakingtypes.BondStatus{stakingtypes.Bonded},
// 			[]superfluidDelegation{{0, 0, "gamm/pool/1", 1000000}},
// 			[]assetTwap{{"gamm/pool/1", sdk.NewDec(0)}},
// 			[]assetTwap{{"gamm/pool/1", sdk.NewDec(10)}},
// 			[]int64{0},
// 		},
// 		{
// 			"dust price twap check",
// 			[]stakingtypes.BondStatus{stakingtypes.Bonded},
// 			[]superfluidDelegation{{0, 0, "gamm/pool/1", 1000000}},
// 			[]assetTwap{{"gamm/pool/1", sdk.NewDecWithPrec(1, 10)}}, // 10^-10
// 			[]assetTwap{},
// 			[]int64{0},
// 		},
// 		{
// 			"refresh case from dust to non-dust",
// 			[]stakingtypes.BondStatus{stakingtypes.Bonded},
// 			[]superfluidDelegation{{0, 0, "gamm/pool/1", 1000000}},
// 			[]assetTwap{{"gamm/pool/1", sdk.NewDecWithPrec(1, 10)}}, // 10^-10
// 			[]assetTwap{{"gamm/pool/1", sdk.NewDec(10)}},
// 			[]int64{0},
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc
// 		suite.Run(tc.name, func() {
// 			suite.SetupTest()

// params := suite.app.SuperfluidKeeper.GetParams(suite.ctx)
// bondDenom := suite.app.StakingKeeper.BondDenom(suite.ctx)

// 			// Generate delegator addresses
// 			delAddrs := CreateRandomAccounts(tc.delegatorNumber)

// 			// setup validators
// 			valAddrs := suite.SetupValidators(tc.validatorStats)
// 			// setup superfluid delegations
// 			intermediaryAccs, locks := suite.SetupSuperfluidDelegations(delAddrs, valAddrs, tc.superDelegations)
// 			suite.checkIntermediaryAccountDelegations(intermediaryAccs)
// 			intermediaryDels := []sdk.Dec{}

// 			for _, intAccIndex := range tc.checkAccIndexes {
// 				expAcc := intermediaryAccs[intAccIndex]
// 				valAddr, err := sdk.ValAddressFromBech32(expAcc.ValAddr)
// 				suite.Require().NoError(err)

// 				// check delegation from intermediary account to validator
// 				delegation, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, expAcc.GetAccAddress(), valAddr)
// 				fmt.Println("====delegation share")
// 				fmt.Println(delegation.Shares)
// 				suite.Require().True(found)
// 				intermediaryDels = append(intermediaryDels, delegation.Shares)
// 			}

// 			// twap price change before refresh
// 			twapByDenom := make(map[string]sdk.Dec)
// 			for _, twap := range tc.roundOneTwaps {
// 				twapByDenom[twap.denom] = twap.price
// 				suite.app.SuperfluidKeeper.SetOsmoEquivalentMultiplier(suite.ctx, 2, twap.denom, twap.price)
// 			}

// 			suite.app.EpochsKeeper.SetEpochInfo(suite.ctx, epochstypes.EpochInfo{
// 				Identifier:   params.RefreshEpochIdentifier,
// 				CurrentEpoch: 2,
// 			})

// 			// get pre-superfluid delgations osmo supply and supplyWithOffset
// 			presupply := suite.app.BankKeeper.GetSupply(suite.ctx, bondDenom)
// 			presupplyWithOffset := suite.app.BankKeeper.GetSupplyWithOffset(suite.ctx, bondDenom)

// 			// refresh intermediary account delegations
// 			suite.NotPanics(func() {
// 				suite.app.SuperfluidKeeper.RefreshIntermediaryDelegationAmounts(suite.ctx)
// 			})

// 			// ensure post-superfluid delegations osmo supplywithoffset is the same while supply is not
// 			postsupply := suite.app.BankKeeper.GetSupply(suite.ctx, bondDenom)
// 			postsupplyWithOffset := suite.app.BankKeeper.GetSupplyWithOffset(suite.ctx, bondDenom)
// 			suite.Require().False(postsupply.IsEqual(presupply), "presupply: %s   postsupply: %s", presupply, postsupply)
// 			suite.Require().True(postsupplyWithOffset.IsEqual(presupplyWithOffset))

// 			originTwap := sdk.NewDec(20)
// 			targetDelegations := []sdk.Dec{}
// 			targetAmounts := []sdk.Int{}
// 			for index, intAccIndex := range tc.checkAccIndexes {
// 				expAcc := intermediaryAccs[intAccIndex]
// 				twap, ok := twapByDenom[expAcc.Denom]
// 				if !ok {
// 					twap = originTwap
// 				}

// 				targetDelegation := intermediaryDels[index].Mul(twap).Quo(originTwap)
// 				fmt.Println("===targetDelegation")
// 				fmt.Println(targetDelegation)
// 				lpTokenAmount := sdk.NewInt(1000000)
// 				decAmt := twap.Mul(lpTokenAmount.ToDec())
// 				asset := suite.app.SuperfluidKeeper.GetSuperfluidAsset(suite.ctx, expAcc.Denom)
// 				targetAmount := suite.app.SuperfluidKeeper.GetRiskAdjustedOsmoValue(suite.ctx, asset, decAmt.RoundInt())
// 				fmt.Println("=====targetAmount")
// 				fmt.Println(targetAmount)
// 				targetDelegations = append(targetDelegations, targetDelegation)
// 				targetAmounts = append(targetAmounts, targetAmount)
// 			}

// 			for index, intAccIndex := range tc.checkAccIndexes {
// 				expAcc := intermediaryAccs[intAccIndex]
// 				valAddr, err := sdk.ValAddressFromBech32(expAcc.ValAddr)
// 				suite.Require().NoError(err)

// 				targetAmount := targetAmounts[index]
// 				targetDelegation := targetDelegations[index]

// 				// check delegation changes
// 				delegation, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, expAcc.GetAccAddress(), valAddr)
// 				fmt.Println("===changed delegation")
// 				fmt.Println(delegation)
// 				if targetAmount.IsPositive() {
// 					suite.Require().True(found)
// 					suite.Require().Equal(delegation.Shares, targetDelegation)
// 				} else {
// 					suite.Require().False(found)
// 				}
// 			}

// 			// start new epoch
// 			suite.app.EpochsKeeper.SetEpochInfo(suite.ctx, epochstypes.EpochInfo{
// 				Identifier:   params.RefreshEpochIdentifier,
// 				CurrentEpoch: 3,
// 			})

// 			// if roundTwo twaps exists, execute round two twaps and finish tests
// 			if len(tc.roundTwoTwaps) > 0 {
// 				twap2ByDenom := make(map[string]sdk.Dec)
// 				for _, twap := range tc.roundTwoTwaps {
// 					twap2ByDenom[twap.denom] = twap.price
// 					suite.app.SuperfluidKeeper.SetOsmoEquivalentMultiplier(suite.ctx, 3, twap.denom, twap.price)
// 				}
// 				// refresh intermediary account delegations
// 				suite.NotPanics(func() {
// 					suite.app.SuperfluidKeeper.RefreshIntermediaryDelegationAmounts(suite.ctx)
// 				})

// 				for index, intAccIndex := range tc.checkAccIndexes {
// 					expAcc := intermediaryAccs[intAccIndex]
// 					valAddr, err := sdk.ValAddressFromBech32(expAcc.ValAddr)
// 					suite.Require().NoError(err)

// 					targetDelegation := intermediaryDels[index].Mul(twap2ByDenom[expAcc.Denom]).Quo(originTwap)

// 					// check delegation changes
// 					delegation, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, expAcc.GetAccAddress(), valAddr)

// 					suite.Require().True(found)
// 					suite.Require().Equal(delegation.Shares, targetDelegation)
// 				}
// 				return
// 			}

// 			// unbond all lockups
// 			for _, lock := range locks {
// 				// superfluid undelegate
// 				err := suite.app.SuperfluidKeeper.SuperfluidUndelegate(suite.ctx, lock.Owner, lock.ID)
// 				suite.Require().NoError(err)
// 			}
// 			unbondingDuration := suite.app.StakingKeeper.GetParams(suite.ctx).UnbondingTime

// 			// check intermediary account changes after unbonding operations
// 			for _, intAccIndex := range tc.checkAccIndexes {
// 				expAcc := intermediaryAccs[intAccIndex]
// 				suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(unbondingDuration + time.Second))
// 				suite.app.EndBlocker(suite.ctx, abci.RequestEndBlock{Height: suite.ctx.BlockHeight()})

// 				unbonded := suite.app.BankKeeper.GetBalance(suite.ctx, expAcc.GetAccAddress(), sdk.DefaultBondDenom)
// 				suite.Require().True(unbonded.IsZero())
// 			}

// 			// refresh intermediary account delegations
// 			suite.NotPanics(func() {
// 				suite.app.SuperfluidKeeper.RefreshIntermediaryDelegationAmounts(suite.ctx)
// 			})

// 			// check changes after refresh operation
// 			for _, intAccIndex := range tc.checkAccIndexes {
// 				expAcc := intermediaryAccs[intAccIndex]
// 				// check unbonded amount is removed after refresh operation
// 				refreshed := suite.app.BankKeeper.GetBalance(suite.ctx, expAcc.GetAccAddress(), sdk.DefaultBondDenom)
// 				suite.Require().True(refreshed.IsZero())
// 			}
// 		})
// 	}
// }

// func (suite *KeeperTestSuite) TestRefreshIntermediaryDelegationAmounts() {
// 	testCases := []struct {
// 		name                     string
// 		validatorStats           []stakingtypes.BondStatus
// 		delegatorNumber          int
// 		superDelegations         []superfluidDelegation
// 		osmoEquivalentMultipliers []osmoEquivalentMultiplier
// 		checkAccIndexes          []int64
// 	}{
// 		{
// 			"with single validator and single delegation",
// 			[]stakingtypes.BondStatus{stakingtypes.Bonded},
// 			1,
// 			[]superfluidDelegation{{0, 0, "gamm/pool/1", 1000000}},
// 			[]osmoEquivalentMultiplier{{"gamm/pool/1", sdk.NewDec(10)}},
// 			[]int64{0},
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc
// 		suite.Run(tc.name, func() {
// 			suite.SetupTest()

// 			params := suite.app.SuperfluidKeeper.GetParams(suite.ctx)
// 			poolId := suite.createGammPool([]string{appparams.BaseCoinUnit, "foo"})
// 			suite.Require().Equal(poolId, uint64(1))
// 			// bondDenom := suite.app.StakingKeeper.BondDenom(suite.ctx)

// 			// Generate delegator addresses
// 			delAddrs := CreateRandomAccounts(tc.delegatorNumber)

// 			// setup validators
// 			valAddrs := suite.SetupValidators(tc.validatorStats)
// 			// setup superfluid delegations
// 			intermediaryAccs, _ := suite.SetupSuperfluidDelegations(delAddrs, valAddrs, tc.superDelegations)
// 			suite.checkIntermediaryAccountDelegations(intermediaryAccs)
// 			intermediaryDels := []sdk.Dec{}

// 			for _, intAccIndex := range tc.checkAccIndexes {
// 				expAcc := intermediaryAccs[intAccIndex]
// 				valAddr, err := sdk.ValAddressFromBech32(expAcc.ValAddr)
// 				suite.Require().NoError(err)

// 				// check delegation from intermediary account to validator
// 				delegation, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, expAcc.GetAccAddress(), valAddr)
// 				suite.Require().True(found)
// 				intermediaryDels = append(intermediaryDels, delegation.Shares)
// 			}

// 			for _, osmoEquivilentMultiplier := range tc.osmoEquivalentMultipliers {
// 				suite.app.SuperfluidKeeper.SetOsmoEquivalentMultiplier(suite.ctx, 2, osmoEquivilentMultiplier.denom, osmoEquivilentMultiplier.price)
// 			}
// 			suite.app.EpochsKeeper.SetEpochInfo(suite.ctx, epochstypes.EpochInfo{
// 				Identifier:   params.RefreshEpochIdentifier,
// 				CurrentEpoch: 2,
// 			})

// 			suite.NotPanics(func() {
// 				suite.app.SuperfluidKeeper.RefreshIntermediaryDelegationAmounts(suite.ctx)
// 			})

// 			for index, intAccIndex := range tc.checkAccIndexes {
// 				expAcc := intermediaryAccs[intAccIndex]
// 				valAddr, err := sdk.ValAddressFromBech32(expAcc.ValAddr)
// 				suite.Require().NoError(err)

// 				// get the updated delegtaion
// 				delegation, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, expAcc.GetAccAddress(), valAddr)

// 			}

// 			panic("e")
// 		})
// 	}
// }
