package keeper_test

import (
	"fmt"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"

	lockuptypes "github.com/osmosis-labs/osmosis/v8/x/lockup/types"
	minttypes "github.com/osmosis-labs/osmosis/v8/x/mint/types"
	"github.com/osmosis-labs/osmosis/v8/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/v8/x/superfluid/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type normalDelegation struct {
	delIndex   int64
	valIndex   int64
	coinAmount int64
}

type superfluidDelegation struct {
	delIndex int64
	valIndex int64
	lpIndex  int64
	lpAmount int64
}
type superfluidRedelegation struct {
	lockId      uint64
	oldValIndex int64
	newValIndex int64
}

type osmoEquivalentMultiplier struct {
	lpIndex int64
	price   sdk.Dec
}

func (suite *KeeperTestSuite) SetupNormalDelegation(delAddrs []sdk.AccAddress, valAddrs []sdk.ValAddress, del normalDelegation) error {
	val, found := suite.App.StakingKeeper.GetValidator(suite.Ctx, valAddrs[del.valIndex])
	if !found {
		return fmt.Errorf("validator not found")
	}
	_, err := suite.App.StakingKeeper.Delegate(suite.Ctx, delAddrs[del.delIndex], sdk.NewIntFromUint64(uint64(del.coinAmount)), stakingtypes.Bonded, val, false)
	return err
}

func (suite *KeeperTestSuite) SetupSuperfluidDelegations(delAddrs []sdk.AccAddress, valAddrs []sdk.ValAddress, superDelegations []superfluidDelegation, denoms []string) ([]types.SuperfluidIntermediaryAccount, []lockuptypes.PeriodLock) {
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
		lock := suite.SetupSuperfluidDelegate(delAddr, valAddr, denoms[del.lpIndex], del.lpAmount)
		address := suite.App.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(suite.Ctx, lock.ID)
		gotAcc := suite.App.SuperfluidKeeper.GetIntermediaryAccount(suite.Ctx, address)

		// save accounts for future use
		if flagIntermediaryAcc[gotAcc.String()] == false {
			flagIntermediaryAcc[gotAcc.String()] = true
			intermediaryAccs = append(intermediaryAccs, gotAcc)
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
		delegation, found := suite.App.StakingKeeper.GetDelegation(suite.Ctx, acc.GetAccAddress(), valAddr)
		suite.Require().True(found)
		suite.Require().True(delegation.Shares.GTE(sdk.NewDec(10000000)))

		// check delegated tokens
		validator, found := suite.App.StakingKeeper.GetValidator(suite.Ctx, valAddr)
		suite.Require().True(found)
		delegatedTokens := validator.TokensFromShares(delegation.Shares).TruncateInt()
		suite.Require().True(delegatedTokens.GTE(sdk.NewInt(10000000)))
	}
}

func (suite *KeeperTestSuite) SetupSuperfluidDelegate(delAddr sdk.AccAddress, valAddr sdk.ValAddress, denom string, amount int64) lockuptypes.PeriodLock {
	unbondingDuration := suite.App.StakingKeeper.GetParams(suite.Ctx).UnbondingTime

	// create lockup of LP token
	coins := sdk.Coins{sdk.NewInt64Coin(denom, amount)}
	lastLockID := suite.App.LockupKeeper.GetLastLockID(suite.Ctx)

	lockID := suite.LockTokens(delAddr, coins, unbondingDuration)
	lock, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, lockID)
	suite.Require().NoError(err)

	// here we check if check `LockTokens` added to existing locks or created a new lock.
	// if `LockTokens` created a new lock, we continue SuperfluidDelegate
	// if lock has been existing before, we wouldn't have to call SuperfluidDelegate separately, as hooks on LockTokens would have automatically called IncreaseSuperfluidDelegation
	if lastLockID != lockID {
		err = suite.App.SuperfluidKeeper.SuperfluidDelegate(suite.Ctx, lock.Owner, lock.ID, valAddr.String())
		suite.Require().NoError(err)
	} else {
		// here we handle two cases.
		// 1. the lock has existed before but has not been superflud staking
		// 2. the lock has existed before and has been superfluid staking

		// we check if synth lock that has existed before, if it did, it means that the lock has been superfluid staked
		// we do not care about unbonding synthlocks, as superfluid delegation has no effect

		_, err := suite.App.LockupKeeper.GetSyntheticLockup(suite.Ctx, lockID, keeper.StakingSyntheticDenom(lock.Coins[0].Denom, valAddr.String()))
		// if lock has existed before but has not been superfluid staked, we do initial superfluid staking
		if err != nil {
			err = suite.App.SuperfluidKeeper.SuperfluidDelegate(suite.Ctx, lock.Owner, lock.ID, valAddr.String())
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
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			[]sdk.Dec{sdk.NewDec(10000000)}, // 50% x 20 x 1000000
		},
		{
			"with single validator and additional superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			1,
			[]superfluidDelegation{{0, 0, 0, 1000000}, {0, 0, 0, 1000000}},
			[]sdk.Dec{sdk.NewDec(20000000)}, // 50% x 20 x 1000000 x 2
		},
		{
			"with multiple validators and multiple superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			2,
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
			[]sdk.Dec{sdk.NewDec(10000000), sdk.NewDec(10000000)}, // 50% x 20 x 1000000
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()
			bondDenom := suite.App.StakingKeeper.BondDenom(suite.Ctx)

			// Generate delegator addresses
			delAddrs := CreateRandomAccounts(tc.delegatorNumber)

			// setup validators
			valAddrs := suite.SetupValidators(tc.validatorStats)

			denoms, _ := suite.SetupGammPoolsAndSuperfluidAssets([]sdk.Dec{sdk.NewDec(20), sdk.NewDec(20)})

			// get pre-superfluid delgations osmo supply and supplyWithOffset
			presupply := suite.App.BankKeeper.GetSupply(suite.Ctx, bondDenom)
			presupplyWithOffset := suite.App.BankKeeper.GetSupplyWithOffset(suite.Ctx, bondDenom)

			// setup superfluid delegations
			_, _, _ = delAddrs, valAddrs, denoms
			intermediaryAccs, locks := suite.SetupSuperfluidDelegations(delAddrs, valAddrs, tc.superDelegations, denoms)

			// ensure post-superfluid delegations osmo supplywithoffset is the same while supply is not
			postsupply := suite.App.BankKeeper.GetSupply(suite.Ctx, bondDenom)
			postsupplyWithOffset := suite.App.BankKeeper.GetSupplyWithOffset(suite.Ctx, bondDenom)
			suite.Require().False(postsupply.IsEqual(presupply), "presupply: %s   postsupply: %s", presupply, postsupply)
			suite.Require().Equal(postsupplyWithOffset.String(), presupplyWithOffset.String())

			unbondingDuration := suite.App.StakingKeeper.GetParams(suite.Ctx).UnbondingTime

			for index, del := range tc.superDelegations {
				lock := locks[index]
				valAddr := valAddrs[del.valIndex]

				// check synthetic lockup creation
				synthLock, err := suite.App.LockupKeeper.GetSyntheticLockup(suite.Ctx, lock.ID, keeper.StakingSyntheticDenom(lock.Coins[0].Denom, valAddr.String()))
				suite.Require().NoError(err)
				suite.Require().Equal(synthLock.UnderlyingLockId, lock.ID)
				suite.Require().Equal(synthLock.SynthDenom, keeper.StakingSyntheticDenom(lock.Coins[0].Denom, valAddr.String()))
				suite.Require().Equal(synthLock.EndTime, time.Time{})

				expAcc := types.NewSuperfluidIntermediaryAccount(lock.Coins[0].Denom, valAddr.String(), 0)

				// Check lockID connection with intermediary account
				intAcc := suite.App.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(suite.Ctx, lock.ID)
				suite.Require().Equal(intAcc.String(), expAcc.GetAccAddress().String())
			}

			for index, expAcc := range intermediaryAccs {
				// check intermediary account creation
				gotAcc := suite.App.SuperfluidKeeper.GetIntermediaryAccount(suite.Ctx, expAcc.GetAccAddress())
				suite.Require().Equal(gotAcc.Denom, expAcc.Denom)
				suite.Require().Equal(gotAcc.ValAddr, expAcc.ValAddr)
				suite.Require().GreaterOrEqual(gotAcc.GaugeId, uint64(1))

				valAddr, err := sdk.ValAddressFromBech32(expAcc.ValAddr)
				suite.Require().NoError(err)

				// check gauge creation
				gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gotAcc.GaugeId)
				suite.Require().NoError(err)
				suite.Require().Equal(gauge.Id, gotAcc.GaugeId)
				suite.Require().Equal(gauge.IsPerpetual, true)
				suite.Require().Equal(gauge.DistributeTo, lockuptypes.QueryCondition{
					LockQueryType: lockuptypes.ByDuration,
					Denom:         keeper.StakingSyntheticDenom(expAcc.Denom, valAddr.String()),
					Duration:      unbondingDuration,
				})
				suite.Require().Equal(gauge.Coins, sdk.Coins(nil))
				suite.Require().Equal(gauge.StartTime, suite.Ctx.BlockTime())
				suite.Require().Equal(gauge.NumEpochsPaidOver, uint64(1))
				suite.Require().Equal(gauge.FilledEpochs, uint64(0))
				suite.Require().Equal(gauge.DistributedCoins, sdk.Coins(nil))

				// check delegation from intermediary account to validator
				delegation, found := suite.App.StakingKeeper.GetDelegation(suite.Ctx, expAcc.GetAccAddress(), valAddr)
				suite.Require().True(found)
				suite.Require().Equal(tc.expInterDelegation[index], delegation.Shares)
			}

			// check invariant is fine
			reason, broken := keeper.AllInvariants(*suite.App.SuperfluidKeeper)(suite.Ctx)
			suite.Require().False(broken, reason)

			// try delegating twice with same lockup
			for _, lock := range locks {
				err := suite.App.SuperfluidKeeper.SuperfluidDelegate(suite.Ctx, lock.Owner, lock.ID, valAddrs[0].String())
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
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			[]uint64{},
			[]uint64{1},
			[]bool{false},
			[]sdk.Dec{sdk.ZeroDec()},
		},
		// {
		// 	"with single validator, single superfluid delegation, add more tokens to the lock, and single undelegation",
		// 	[]stakingtypes.BondStatus{stakingtypes.Bonded},
		// 	[]superfluidDelegation{{0, 0, 1000000}},
		// 	[]uint64{1},
		// 	[]uint64{1},
		// 	[]bool{false},
		// 	[]sdk.Dec{sdk.ZeroDec()},
		// },
		{
			"with single validator and additional superfluid delegations and single undelegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			1,
			[]superfluidDelegation{{0, 0, 0, 1000000}, {0, 0, 0, 1000000}},
			[]uint64{},
			[]uint64{1},
			[]bool{false},
			[]sdk.Dec{sdk.ZeroDec()},
		},
		{
			"with multiple validators and multiple superfluid delegations and multiple undelegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			2,
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
			[]uint64{},
			[]uint64{1, 2},
			[]bool{false, false},
			[]sdk.Dec{sdk.ZeroDec()},
		},
		{
			"undelegating not available lock id",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			1,
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			[]uint64{},
			[]uint64{2},
			[]bool{true},
			[]sdk.Dec{},
		},
		{
			"try undelegating twice for same lock id",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			1,
			[]superfluidDelegation{{0, 0, 0, 1000000}},
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

			bondDenom := suite.App.StakingKeeper.GetParams(suite.Ctx).BondDenom

			// Generate delegator addresses
			delAddrs := CreateRandomAccounts(tc.delegatorNumber)

			// setup validators
			valAddrs := suite.SetupValidators(tc.validatorStats)

			denoms, _ := suite.SetupGammPoolsAndSuperfluidAssets([]sdk.Dec{sdk.NewDec(20), sdk.NewDec(20)})

			// setup superfluid delegations
			intermediaryAccs, _ := suite.SetupSuperfluidDelegations(delAddrs, valAddrs, tc.superDelegations, denoms)
			suite.checkIntermediaryAccountDelegations(intermediaryAccs)

			for _, lockId := range tc.addMoreTokensLockIds {
				lock, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, lockId)
				suite.Require().NoError(err)
				lockOwner, err := sdk.AccAddressFromBech32(lock.Owner)
				suite.Require().NoError(err)
				coin := sdk.NewInt64Coin("gamm/pool/1", 1000000)
				suite.App.BankKeeper.MintCoins(suite.Ctx, minttypes.ModuleName, sdk.NewCoins(coin))
				suite.App.BankKeeper.SendCoinsFromModuleToAccount(suite.Ctx, minttypes.ModuleName, lockOwner, sdk.NewCoins(coin))
				_, err = suite.App.LockupKeeper.AddTokensToLockByID(suite.Ctx, lockId, lockOwner, coin)
				suite.Require().NoError(err)
			}

			for index, lockId := range tc.superUnbondingLockIds {
				// get intermediary account
				accAddr := suite.App.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(suite.Ctx, lockId)
				intermediaryAcc := suite.App.SuperfluidKeeper.GetIntermediaryAccount(suite.Ctx, accAddr)
				valAddr := intermediaryAcc.ValAddr

				lock, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, lockId)
				if err != nil {
					lock = &lockuptypes.PeriodLock{}
				}

				// get pre-superfluid delgations osmo supply and supplyWithOffset
				presupply := suite.App.BankKeeper.GetSupply(suite.Ctx, bondDenom)
				presupplyWithOffset := suite.App.BankKeeper.GetSupplyWithOffset(suite.Ctx, bondDenom)

				// superfluid undelegate
				err = suite.App.SuperfluidKeeper.SuperfluidUndelegate(suite.Ctx, lock.Owner, lockId)
				if tc.expSuperUnbondingErr[index] {
					suite.Require().Error(err)
					continue
				}
				suite.Require().NoError(err)

				// ensure post-superfluid delegations osmo supplywithoffset is the same while supply is not
				postsupply := suite.App.BankKeeper.GetSupply(suite.Ctx, bondDenom)
				postsupplyWithOffset := suite.App.BankKeeper.GetSupplyWithOffset(suite.Ctx, bondDenom)
				suite.Require().False(postsupply.IsEqual(presupply), "presupply: %s   postsupply: %s", presupply, postsupply)
				suite.Require().True(postsupplyWithOffset.IsEqual(presupplyWithOffset))

				// check lockId and intermediary account connection deletion
				addr := suite.App.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(suite.Ctx, lockId)
				suite.Require().Equal(addr.String(), "")

				// check bonding synthetic lockup deletion
				_, err = suite.App.LockupKeeper.GetSyntheticLockup(suite.Ctx, lockId, keeper.StakingSyntheticDenom(lock.Coins[0].Denom, valAddr))
				suite.Require().Error(err)

				// check unbonding synthetic lockup creation
				unbondingDuration := suite.App.StakingKeeper.GetParams(suite.Ctx).UnbondingTime
				synthLock, err := suite.App.LockupKeeper.GetSyntheticLockup(suite.Ctx, lockId, keeper.UnstakingSyntheticDenom(lock.Coins[0].Denom, valAddr))
				suite.Require().NoError(err)
				suite.Require().Equal(synthLock.UnderlyingLockId, lockId)
				suite.Require().Equal(synthLock.SynthDenom, keeper.UnstakingSyntheticDenom(lock.Coins[0].Denom, valAddr))
				suite.Require().Equal(synthLock.EndTime, suite.Ctx.BlockTime().Add(unbondingDuration))
			}

			// check invariant is fine
			reason, broken := keeper.AllInvariants(*suite.App.SuperfluidKeeper)(suite.Ctx)
			suite.Require().False(broken, reason)

			// check remaining intermediary account delegation
			for index, expDelegation := range tc.expInterDelegation {
				acc := intermediaryAccs[index]
				valAddr, err := sdk.ValAddressFromBech32(acc.ValAddr)
				suite.Require().NoError(err)
				delegation, found := suite.App.StakingKeeper.GetDelegation(suite.Ctx, acc.GetAccAddress(), valAddr)
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

				lock, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, lockId)
				suite.Require().NoError(err)

				err = suite.App.SuperfluidKeeper.SuperfluidUndelegate(suite.Ctx, lock.Owner, lockId)
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

	denoms, _ := suite.SetupGammPoolsAndSuperfluidAssets([]sdk.Dec{sdk.NewDec(20), sdk.NewDec(20)})

	// setup superfluid delegations
	intermediaryAccs, locks := suite.SetupSuperfluidDelegations(delAddrs, valAddrs, []superfluidDelegation{{0, 0, 0, 1000000}}, denoms)
	suite.checkIntermediaryAccountDelegations(intermediaryAccs)

	for _, lock := range locks {
		startTime := time.Now()
		suite.Ctx = suite.Ctx.WithBlockTime(startTime)
		accAddr := suite.App.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(suite.Ctx, lock.ID)
		intermediaryAcc := suite.App.SuperfluidKeeper.GetIntermediaryAccount(suite.Ctx, accAddr)
		valAddr := intermediaryAcc.ValAddr

		// first we test that SuperfluidUnbondLock would cause error before undelegating
		err := suite.App.SuperfluidKeeper.SuperfluidUnbondLock(suite.Ctx, lock.ID, lock.GetOwner())
		suite.Require().Error(err)

		// undelegation needs to happen prior to SuperfluidUnbondLock
		err = suite.App.SuperfluidKeeper.SuperfluidUndelegate(suite.Ctx, lock.Owner, lock.ID)
		suite.Require().NoError(err)
		balances := suite.App.BankKeeper.GetAllBalances(suite.Ctx, lock.OwnerAddress())
		suite.Require().Equal(0, balances.Len())

		// check that unbonding synth has been created correctly after undelegation
		unbondingDuration := suite.App.StakingKeeper.GetParams(suite.Ctx).UnbondingTime
		synthLock, err := suite.App.LockupKeeper.GetSyntheticLockup(suite.Ctx, lock.ID, keeper.UnstakingSyntheticDenom(lock.Coins[0].Denom, valAddr))
		suite.Require().NoError(err)
		suite.Require().Equal(synthLock.UnderlyingLockId, lock.ID)
		suite.Require().Equal(synthLock.SynthDenom, keeper.UnstakingSyntheticDenom(lock.Coins[0].Denom, valAddr))
		suite.Require().Equal(synthLock.EndTime, suite.Ctx.BlockTime().Add(unbondingDuration))

		// test SuperfluidUnbondLock
		unbondLockStartTime := startTime.Add(time.Hour)
		suite.Ctx = suite.Ctx.WithBlockTime(unbondLockStartTime)
		err = suite.App.SuperfluidKeeper.SuperfluidUnbondLock(suite.Ctx, lock.ID, lock.GetOwner())
		suite.Require().NoError(err)

		// check that SuperfluidUnbondLock makes underlying lock start unlocking
		// we run WithdrawAllMaturedLocks to ensure that lock isn't getting finished immediately
		suite.App.LockupKeeper.WithdrawAllMaturedLocks(suite.Ctx)
		updatedLock, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, lock.ID)
		suite.Require().NoError(err)
		suite.Require().True(updatedLock.IsUnlocking())

		// check if finsihed unlocking synth lock did not increase balance
		balances = suite.App.BankKeeper.GetAllBalances(suite.Ctx, lock.OwnerAddress())
		suite.Require().Equal(0, balances.Len())

		// test that synth lock finish does not mean underlying lock is finished
		suite.Ctx = suite.Ctx.WithBlockTime((startTime.Add(unbondingDuration)))
		suite.App.LockupKeeper.DeleteAllMaturedSyntheticLocks(suite.Ctx)
		suite.App.LockupKeeper.WithdrawAllMaturedLocks(suite.Ctx)
		_, err = suite.App.LockupKeeper.GetSyntheticLockup(suite.Ctx, lock.ID, keeper.UnstakingSyntheticDenom(lock.Coins[0].Denom, valAddr))
		suite.Require().Error(err)
		updatedLock, err = suite.App.LockupKeeper.GetLockByID(suite.Ctx, lock.ID)
		suite.Require().NoError(err)
		suite.Require().True(updatedLock.IsUnlocking())

		// test after SuperfluidUnbondLock + lockup unbonding duration, lock is finished and does not exist
		suite.Ctx = suite.Ctx.WithBlockTime(unbondLockStartTime.Add(unbondingDuration))
		suite.App.LockupKeeper.WithdrawAllMaturedLocks(suite.Ctx)
		_, err = suite.App.LockupKeeper.GetLockByID(suite.Ctx, lock.ID)
		suite.Require().Error(err)

		// check if finished unlocking successfully increased balance
		balances = suite.App.BankKeeper.GetAllBalances(suite.Ctx, lock.OwnerAddress())
		suite.Require().Equal(1, balances.Len())
		suite.Require().Equal(denoms[0], balances[0].Denom)
		suite.Require().Equal(sdk.NewInt(1000000), balances[0].Amount)

		// check invariant is fine
		reason, broken := keeper.AllInvariants(*suite.App.SuperfluidKeeper)(suite.Ctx)
		suite.Require().False(broken, reason)
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
// 				lock, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, srd.lockId)
// 				if err != nil {
// 					lock = &lockuptypes.PeriodLock{}
// 				}

// 				// superfluid redelegate
// 				err = suite.App.SuperfluidKeeper.SuperfluidRedelegate(suite.Ctx, lock.Owner, srd.lockId, valAddrs[srd.newValIndex].String())
// 				if tc.expSuperRedelegationErr[index] {
// 					suite.Require().Error(err)
// 					continue
// 				}
// 				suite.Require().NoError(err)

// 				// check previous validator bonding synthetic lockup deletion
// 				_, err = suite.App.LockupKeeper.GetSyntheticLockup(suite.Ctx, srd.lockId, keeper.StakingSuffix(valAddrs[srd.oldValIndex].String()))
// 				suite.Require().Error(err)

// 				// check unbonding synthetic lockup creation
// 				params := suite.App.SuperfluidKeeper.GetParams(suite.Ctx)
// 				synthLock, err := suite.App.LockupKeeper.GetSyntheticLockup(suite.Ctx, srd.lockId, keeper.UnstakingSuffix(valAddrs[srd.oldValIndex].String()))
// 				suite.Require().NoError(err)
// 				suite.Require().Equal(synthLock.UnderlyingLockId, srd.lockId)
// 				suite.Require().Equal(synthLock.Suffix, keeper.UnstakingSuffix(valAddrs[srd.oldValIndex].String()))
// 				suite.Require().Equal(synthLock.EndTime, suite.Ctx.BlockTime().Add(params.UnbondingDuration))

// 				// check synthetic lockup creation
// 				synthLock2, err := suite.App.LockupKeeper.GetSyntheticLockup(suite.Ctx, srd.lockId, keeper.StakingSuffix(valAddrs[srd.newValIndex].String()))
// 				suite.Require().NoError(err)
// 				suite.Require().Equal(synthLock2.UnderlyingLockId, srd.lockId)
// 				suite.Require().Equal(synthLock2.Suffix, keeper.StakingSuffix(valAddrs[srd.newValIndex].String()))
// 				suite.Require().Equal(synthLock2.EndTime, time.Time{})

// 				// check intermediary account creation
// 				lock, err = suite.App.LockupKeeper.GetLockByID(suite.Ctx, srd.lockId)
// 				suite.Require().NoError(err)

// 				expAcc := types.NewSuperfluidIntermediaryAccount(lock.Coins[0].Denom, valAddrs[srd.newValIndex].String(), 1)
// 				gotAcc := suite.App.SuperfluidKeeper.GetIntermediaryAccount(suite.Ctx, expAcc.GetAccAddress())
// 				suite.Require().Equal(gotAcc.Denom, expAcc.Denom)
// 				suite.Require().Equal(gotAcc.ValAddr, expAcc.ValAddr)

// 				// check gauge creation
// 				gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gotAcc.GaugeId)
// 				suite.Require().NoError(err)
// 				suite.Require().Equal(gauge.Id, gotAcc.GaugeId)
// 				suite.Require().Equal(gauge.IsPerpetual, true)
// 				suite.Require().Equal(gauge.DistributeTo, lockuptypes.QueryCondition{
// 					LockQueryType: lockuptypes.ByDuration,
// 					Denom:         expAcc.Denom + keeper.StakingSuffix(valAddrs[srd.newValIndex].String()),
// 					Duration:      params.UnbondingDuration,
// 				})
// 				suite.Require().Equal(gauge.Coins, sdk.Coins(nil))
// 				suite.Require().Equal(gauge.StartTime, suite.Ctx.BlockTime())
// 				suite.Require().Equal(gauge.NumEpochsPaidOver, uint64(1))
// 				suite.Require().Equal(gauge.FilledEpochs, uint64(0))
// 				suite.Require().Equal(gauge.DistributedCoins, sdk.Coins(nil))

// 				// Check lockID connection with intermediary account
// 				intAcc := suite.App.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(suite.Ctx, srd.lockId)
// 				suite.Require().Equal(intAcc.String(), expAcc.GetAccAddress().String())

// 				// check delegation from intermediary account to validator
// 				_, found := suite.App.StakingKeeper.GetDelegation(suite.Ctx, expAcc.GetAccAddress(), valAddrs[srd.newValIndex])
// 				suite.Require().True(found)
// 			}

// 			// try redelegating twice
// 			for index, srd := range tc.superRedelegations {
// 				if tc.expSuperRedelegationErr[index] {
// 					continue
// 				}
// 				cacheCtx, _ := suite.Ctx.CacheContext()
// 				lock, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, srd.lockId)
// 				suite.Require().NoError(err)
// 				err = suite.App.SuperfluidKeeper.SuperfluidRedelegate(cacheCtx, lock.Owner, srd.lockId, valAddrs[srd.newValIndex].String())
// 				suite.Require().Error(err)
// 			}
// 		})
// 	}
// }

func (suite *KeeperTestSuite) TestRefreshIntermediaryDelegationAmounts() {
	testCases := []struct {
		name                string
		validatorStats      []stakingtypes.BondStatus
		delegatorNumber     int
		superDelegations    []superfluidDelegation
		roundOneMultipliers []osmoEquivalentMultiplier
		roundTwoMultipliers []osmoEquivalentMultiplier
		checkAccIndexes     []int64
	}{
		{
			"with single validator and single delegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			1,
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			[]osmoEquivalentMultiplier{{0, sdk.NewDec(10)}},
			[]osmoEquivalentMultiplier{},
			[]int64{0},
		},
		{
			"with single validator and additional delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			1,
			[]superfluidDelegation{{0, 0, 0, 1000000}, {0, 0, 0, 1000000}},
			[]osmoEquivalentMultiplier{{0, sdk.NewDec(10)}},
			[]osmoEquivalentMultiplier{},
			[]int64{0},
		},
		{
			"with multiple validator and multiple superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			2,
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
			[]osmoEquivalentMultiplier{{0, sdk.NewDec(10)}},
			[]osmoEquivalentMultiplier{},
			[]int64{0, 1},
		},
		{
			"with single validator and multiple denom superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			2,
			[]superfluidDelegation{{0, 0, 0, 1000000}, {0, 0, 1, 1000000}},
			[]osmoEquivalentMultiplier{{0, sdk.NewDec(10)}, {1, sdk.NewDec(10)}},
			[]osmoEquivalentMultiplier{},
			[]int64{0, 1},
		},
		{
			"with multiple validators and multiple denom superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			2,
			[]superfluidDelegation{{0, 0, 0, 1000000}, {0, 1, 1, 1000000}},
			[]osmoEquivalentMultiplier{{0, sdk.NewDec(10)}, {1, sdk.NewDec(10)}},
			[]osmoEquivalentMultiplier{},
			[]int64{0, 1},
		},
		{
			"zero price multiplier check",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			2,
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			[]osmoEquivalentMultiplier{{0, sdk.NewDec(0)}},
			[]osmoEquivalentMultiplier{},
			[]int64{0},
		},
		{
			"refresh case from zero to non-zero",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			2,
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			[]osmoEquivalentMultiplier{{0, sdk.NewDec(0)}},
			[]osmoEquivalentMultiplier{{0, sdk.NewDec(10)}},
			[]int64{0},
		},
		{
			"dust price multiplier check",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			2,
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			[]osmoEquivalentMultiplier{{0, sdk.NewDecWithPrec(1, 10)}}, // 10^-10
			[]osmoEquivalentMultiplier{},
			[]int64{0},
		},
		{
			"refresh case from dust to non-dust",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			2,
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			[]osmoEquivalentMultiplier{{0, sdk.NewDecWithPrec(1, 10)}}, // 10^-10
			[]osmoEquivalentMultiplier{{0, sdk.NewDec(10)}},
			[]int64{0},
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()

			bondDenom := suite.App.StakingKeeper.BondDenom(suite.Ctx)

			// Generate delegator addresses
			delAddrs := CreateRandomAccounts(tc.delegatorNumber)

			// setup validators
			valAddrs := suite.SetupValidators(tc.validatorStats)

			denoms, _ := suite.SetupGammPoolsAndSuperfluidAssets([]sdk.Dec{sdk.NewDec(20), sdk.NewDec(20)})

			// setup superfluid delegations
			intermediaryAccs, locks := suite.SetupSuperfluidDelegations(delAddrs, valAddrs, tc.superDelegations, denoms)
			suite.checkIntermediaryAccountDelegations(intermediaryAccs)
			intermediaryDels := []sdk.Dec{}

			for _, intAccIndex := range tc.checkAccIndexes {
				expAcc := intermediaryAccs[intAccIndex]
				valAddr, err := sdk.ValAddressFromBech32(expAcc.ValAddr)
				suite.Require().NoError(err)

				// check delegation from intermediary account to validator
				delegation, found := suite.App.StakingKeeper.GetDelegation(suite.Ctx, expAcc.GetAccAddress(), valAddr)
				suite.Require().True(found)
				intermediaryDels = append(intermediaryDels, delegation.Shares)
			}

			// multiplier change before refresh
			multiplierByDenom := make(map[string]sdk.Dec)
			for _, multiplier := range tc.roundOneMultipliers {
				denom := denoms[multiplier.lpIndex]
				multiplierByDenom[denom] = multiplier.price
				suite.App.SuperfluidKeeper.SetOsmoEquivalentMultiplier(suite.Ctx, 2, denom, multiplier.price)
			}

			// get pre-superfluid delgations osmo supply and supplyWithOffset
			presupply := suite.App.BankKeeper.GetSupply(suite.Ctx, bondDenom)
			presupplyWithOffset := suite.App.BankKeeper.GetSupplyWithOffset(suite.Ctx, bondDenom)

			// refresh intermediary account delegations
			suite.App.SuperfluidKeeper.RefreshIntermediaryDelegationAmounts(suite.Ctx)

			// ensure post-superfluid delegations osmo supplywithoffset is the same while supply is not
			postsupply := suite.App.BankKeeper.GetSupply(suite.Ctx, bondDenom)
			postsupplyWithOffset := suite.App.BankKeeper.GetSupplyWithOffset(suite.Ctx, bondDenom)
			suite.Require().False(postsupply.IsEqual(presupply), "presupply: %s   postsupply: %s", presupply, postsupply)
			suite.Require().True(postsupplyWithOffset.IsEqual(presupplyWithOffset))

			originMultiplier := sdk.NewDec(20)
			for index, intAccIndex := range tc.checkAccIndexes {
				expAcc := intermediaryAccs[intAccIndex]
				multiplier, ok := multiplierByDenom[expAcc.Denom]
				if !ok {
					multiplier = originMultiplier
				}

				// calculating the estimated delegation amount for multiplier change
				expDelegation := intermediaryDels[index].Mul(multiplier).Quo(originMultiplier)
				lpTokenAmount := sdk.NewInt(1000000)
				decAmt := multiplier.Mul(lpTokenAmount.ToDec())
				asset := suite.App.SuperfluidKeeper.GetSuperfluidAsset(suite.Ctx, expAcc.Denom)
				expAmount := suite.App.SuperfluidKeeper.GetRiskAdjustedOsmoValue(suite.Ctx, asset, decAmt.RoundInt())

				valAddr, err := sdk.ValAddressFromBech32(expAcc.ValAddr)
				suite.Require().NoError(err)

				// check delegation changes
				delegation, found := suite.App.StakingKeeper.GetDelegation(suite.Ctx, expAcc.GetAccAddress(), valAddr)
				if expAmount.IsPositive() {
					suite.Require().True(found)
					suite.Require().Equal(delegation.Shares, expDelegation)
				} else {
					suite.Require().False(found)
				}
			}

			// if second round multipliers exists, execute round two multipliers mock and finish tests
			if len(tc.roundTwoMultipliers) > 0 {
				multiplier2ByDenom := make(map[string]sdk.Dec)
				for _, multiplier := range tc.roundTwoMultipliers {
					denom := denoms[multiplier.lpIndex]
					multiplier2ByDenom[denom] = multiplier.price
					suite.App.SuperfluidKeeper.SetOsmoEquivalentMultiplier(suite.Ctx, 3, denom, multiplier.price)
				}
				// refresh intermediary account delegations
				suite.App.SuperfluidKeeper.RefreshIntermediaryDelegationAmounts(suite.Ctx)

				for index, intAccIndex := range tc.checkAccIndexes {
					expAcc := intermediaryAccs[intAccIndex]
					valAddr, err := sdk.ValAddressFromBech32(expAcc.ValAddr)
					suite.Require().NoError(err)

					expDelegation := intermediaryDels[index].Mul(multiplier2ByDenom[expAcc.Denom]).Quo(originMultiplier)

					// check delegation changes
					delegation, found := suite.App.StakingKeeper.GetDelegation(suite.Ctx, expAcc.GetAccAddress(), valAddr)

					suite.Require().True(found)
					suite.Require().Equal(delegation.Shares, expDelegation)
				}
				return
			}

			// unbond all lockups
			for _, lock := range locks {
				// superfluid undelegate
				// handling the case same lockup is used for further delegation
				cacheCtx, write := suite.Ctx.CacheContext()
				err := suite.App.SuperfluidKeeper.SuperfluidUndelegate(cacheCtx, lock.Owner, lock.ID)
				if err == nil {
					write()
				}
			}
			unbondingDuration := suite.App.StakingKeeper.GetParams(suite.Ctx).UnbondingTime

			// check intermediary account changes after unbonding operations
			for _, intAccIndex := range tc.checkAccIndexes {
				expAcc := intermediaryAccs[intAccIndex]
				suite.Ctx = suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(unbondingDuration + time.Second))
				suite.App.EndBlocker(suite.Ctx, abci.RequestEndBlock{Height: suite.Ctx.BlockHeight()})

				unbonded := suite.App.BankKeeper.GetBalance(suite.Ctx, expAcc.GetAccAddress(), sdk.DefaultBondDenom)
				suite.Require().True(unbonded.IsZero())
			}

			// refresh intermediary account delegations
			suite.App.SuperfluidKeeper.RefreshIntermediaryDelegationAmounts(suite.Ctx)

			// check if intermediary accounts does not have free balance after refresh operation
			for _, intAccIndex := range tc.checkAccIndexes {
				expAcc := intermediaryAccs[intAccIndex]
				// check unbonded amount is removed after refresh operation
				refreshed := suite.App.BankKeeper.GetBalance(suite.Ctx, expAcc.GetAccAddress(), sdk.DefaultBondDenom)
				suite.Require().True(refreshed.IsZero())
			}
		})
	}
}

func (suite *KeeperTestSuite) TestSuperfluidDelegationGovernanceVoting() {
	testCases := []struct {
		name              string
		validatorStats    []stakingtypes.BondStatus
		superDelegations  [][]superfluidDelegation
		normalDelegations []normalDelegation
	}{
		{
			"with single validator and single delegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[][]superfluidDelegation{{{0, 0, 0, 1000000}}},
			nil,
		},
		{
			"with single validator and additional delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[][]superfluidDelegation{{{0, 0, 0, 1000000}, {0, 0, 0, 1000000}}},
			nil,
		},
		{
			"with multiple validator and multiple superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			[][]superfluidDelegation{{{0, 0, 0, 1000000}}, {{1, 1, 0, 1000000}}},
			nil,
		},
		{
			"with single validator and multiple denom superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			[][]superfluidDelegation{{{0, 0, 0, 1000000}, {0, 0, 1, 1000000}}},
			nil,
		},
		{
			"with multiple validators and multiple denom superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			[][]superfluidDelegation{{{0, 0, 0, 1000000}, {0, 1, 1, 1000000}}},
			nil,
		},
		{
			"many delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			[][]superfluidDelegation{
				{{0, 0, 0, 1000000}, {0, 1, 1, 1000000}},
				{{1, 0, 0, 1000000}, {1, 0, 1, 1000000}},
				{{2, 1, 1, 1000000}, {2, 1, 0, 1000000}},
				{{3, 0, 0, 1000000}, {3, 1, 1, 1000000}},
			},
			nil,
		},
		{
			"with normal delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[][]superfluidDelegation{
				{{0, 0, 0, 1000000}, {0, 0, 1, 1000000}},
			},
			[]normalDelegation{
				{0, 0, 1000000},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()

			denoms, _ := suite.SetupGammPoolsAndSuperfluidAssets([]sdk.Dec{sdk.NewDec(20), sdk.NewDec(20)})

			// Generate delegator addresses
			delAddrs := CreateRandomAccounts(len(tc.superDelegations))

			// setup validators
			valAddrs := suite.SetupValidators(tc.validatorStats)

			// setup superfluid delegations
			for _, sfdel := range tc.superDelegations {
				intermediaryAccs, _ := suite.SetupSuperfluidDelegations(delAddrs, valAddrs, sfdel, denoms)
				suite.checkIntermediaryAccountDelegations(intermediaryAccs)
			}

			// setup normal delegations
			for _, del := range tc.normalDelegations {
				err := suite.SetupNormalDelegation(delAddrs, valAddrs, del)
				suite.NoError(err)
			}

			// all expected delegated amounts to a validator from a delegator
			delegatedAmount := func(delidx, validx int) sdk.Int {
				res := sdk.ZeroInt()
				for _, del := range tc.superDelegations[delidx] {
					if del.valIndex == int64(validx) {
						res = res.AddRaw(del.lpAmount)
					}
				}
				if len(tc.normalDelegations) != 0 {
					del := tc.normalDelegations[delidx]
					res = res.AddRaw(del.coinAmount / 10) // LP price is 10 osmo in this test
				}
				return res
			}
			for delidx := range tc.superDelegations {
				// store all actual delegations to a validator
				sharePerValidatorMap := make(map[string]sdk.Dec)
				for validx := range tc.validatorStats {
					sharePerValidatorMap[valAddrs[validx].String()] = sdk.ZeroDec()
				}
				addToSharePerValidatorMap := func(val sdk.ValAddress, share sdk.Dec) {
					if existing, ok := sharePerValidatorMap[val.String()]; ok {
						share.AddMut(existing)
					}
					sharePerValidatorMap[val.String()] = share
				}

				// iterate delegations and add eligible shares to the sharePerValidatorMap
				suite.App.SuperfluidKeeper.IterateDelegations(suite.Ctx, delAddrs[delidx], func(_ int64, del stakingtypes.DelegationI) bool {
					addToSharePerValidatorMap(del.GetValidatorAddr(), del.GetShares())
					return false
				})

				// check if the expected delegated amount equals to actual
				for validx := range tc.validatorStats {
					suite.Equal(delegatedAmount(delidx, validx).Int64()*10, sharePerValidatorMap[valAddrs[validx].String()].RoundInt().Int64())
				}
			}
		})
	}
}
