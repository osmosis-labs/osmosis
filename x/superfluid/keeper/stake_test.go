package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	appparams "github.com/osmosis-labs/osmosis/app/params"
	epochstypes "github.com/osmosis-labs/osmosis/x/epochs/types"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
	minttypes "github.com/osmosis-labs/osmosis/x/mint/types"
	"github.com/osmosis-labs/osmosis/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/x/superfluid/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

type superfluidDelegation struct {
	valIndex int64
	lpDenom  string
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

func (suite *KeeperTestSuite) LockTokens(addr sdk.AccAddress, coins sdk.Coins, duration time.Duration) lockuptypes.PeriodLock {
	err := suite.app.BankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, coins)
	suite.Require().NoError(err)
	err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, minttypes.ModuleName, addr, coins)
	suite.Require().NoError(err)
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

	amount := sdk.TokensFromConsensusPower(1, sdk.DefaultPowerReduction)
	issuedShares := amount.ToDec()
	validator.Status = bondStatus
	validator.Tokens = validator.Tokens.Add(amount)
	validator.DelegatorShares = validator.DelegatorShares.Add(issuedShares)

	suite.app.StakingKeeper.SetValidator(suite.ctx, validator)
	suite.app.StakingKeeper.SetValidatorByConsAddr(suite.ctx, validator)
	suite.app.StakingKeeper.SetValidatorByPowerIndex(suite.ctx, validator)
	suite.app.StakingKeeper.AfterValidatorCreated(suite.ctx, validator.GetOperator())

	bondDenom := suite.app.StakingKeeper.BondDenom(suite.ctx)
	coins := sdk.Coins{sdk.NewCoin(bondDenom, amount)}
	err = suite.app.BankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, coins)
	suite.Require().NoError(err)
	if bondStatus == stakingtypes.Bonded {
		err = suite.app.BankKeeper.SendCoinsFromModuleToModule(suite.ctx, minttypes.ModuleName, stakingtypes.BondedPoolName, coins)
		suite.Require().NoError(err)
	} else {
		err = suite.app.BankKeeper.SendCoinsFromModuleToModule(suite.ctx, minttypes.ModuleName, stakingtypes.NotBondedPoolName, coins)
		suite.Require().NoError(err)
	}
	return valAddr
}

func (suite *KeeperTestSuite) SetupValidators(bondStatuses []stakingtypes.BondStatus) []sdk.ValAddress {
	valAddrs := []sdk.ValAddress{}
	for _, status := range bondStatuses {
		valAddr := suite.SetupValidator(status)
		valAddrs = append(valAddrs, valAddr)
	}
	return valAddrs
}

func (suite *KeeperTestSuite) SetupSuperfluidDelegations(valAddrs []sdk.ValAddress, superDelegations []superfluidDelegation) ([]types.SuperfluidIntermediaryAccount, []lockuptypes.PeriodLock) {
	flagIntermediaryAcc := make(map[string]bool)
	intermediaryAccs := []types.SuperfluidIntermediaryAccount{}
	locks := []lockuptypes.PeriodLock{}

	// setup superfluid delegations
	for _, del := range superDelegations {
		valAddr := valAddrs[del.valIndex]
		lock := suite.SetupSuperfluidDelegate(valAddr, del.lpDenom)
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
		suite.Require().True(delegation.Shares.GTE(sdk.NewDec(19000000)))

		// check delegated tokens
		validator, found := suite.app.StakingKeeper.GetValidator(suite.ctx, valAddr)
		suite.Require().True(found)
		delegatedTokens := validator.TokensFromShares(delegation.Shares).TruncateInt()
		suite.Require().True(delegatedTokens.GTE(sdk.NewInt(19000000)))
	}
}

func (suite *KeeperTestSuite) SetupSuperfluidDelegate(valAddr sdk.ValAddress, denom string) lockuptypes.PeriodLock {

	suite.app.IncentivesKeeper.SetLockableDurations(suite.ctx, []time.Duration{
		time.Hour * 24 * 14,
		time.Hour,
		time.Hour * 3,
		time.Hour * 7,
		suite.app.SuperfluidKeeper.GetParams(suite.ctx).UnbondingDuration,
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
	lock := suite.LockTokens(addr1, coins, params.UnbondingDuration)

	// call SuperfluidDelegate and check response
	err := suite.app.SuperfluidKeeper.SuperfluidDelegate(suite.ctx, lock.Owner, lock.ID, valAddr.String())
	suite.Require().NoError(err)

	return lock
}

func (suite *KeeperTestSuite) TestSuperfluidDelegate() {
	testCases := []struct {
		name               string
		validatorStats     []stakingtypes.BondStatus
		superDelegations   []superfluidDelegation
		expInterDelegation []sdk.Dec
	}{
		{
			"with single validator and single superfluid delegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}},
			[]sdk.Dec{sdk.NewDec(19000000)}, // 95% x 2 x 1000000
		},
		{
			"with single validator and multiple superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}, {0, "gamm/pool/1"}},
			[]sdk.Dec{sdk.NewDec(38000000)}, // 95% x 2 x 1000000 x 2
		},
		{
			"with multiples validator and multiple superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}, {1, "gamm/pool/1"}},
			[]sdk.Dec{sdk.NewDec(19000000), sdk.NewDec(19000000)}, // 95% x 2 x 1000000
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()

			poolId := suite.createGammPool([]string{appparams.BaseCoinUnit, "foo"})
			suite.Require().Equal(poolId, uint64(1))

			// setup validators
			valAddrs := suite.SetupValidators(tc.validatorStats)
			intermediaryAccs, locks := suite.SetupSuperfluidDelegations(valAddrs, tc.superDelegations)
			params := suite.app.SuperfluidKeeper.GetParams(suite.ctx)

			// setup superfluid delegations
			for index, del := range tc.superDelegations {
				lock := locks[index]
				valAddr := valAddrs[del.valIndex]

				// check synthetic lockup creation
				synthLock, err := suite.app.LockupKeeper.GetSyntheticLockup(suite.ctx, lock.ID, keeper.StakingSuffix(valAddr.String()))
				suite.Require().NoError(err)
				suite.Require().Equal(synthLock.UnderlyingLockId, lock.ID)
				suite.Require().Equal(synthLock.Suffix, keeper.StakingSuffix(valAddr.String()))
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
				suite.Require().Equal(delegation.Shares, tc.expInterDelegation[index])
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
		superDelegations      []superfluidDelegation
		addMoreTokensLockIds  []uint64
		superUnbondingLockIds []uint64
		expSuperUnbondingErr  []bool
		expInterDelegation    []sdk.Dec
	}{
		{
			"with single validator and single superfluid delegation and single undelegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}},
			[]uint64{},
			[]uint64{1},
			[]bool{false},
			[]sdk.Dec{sdk.ZeroDec()},
		},
		{
			"with single validator, single superfluid delegation, add more tokens to the lock, and single undelegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}},
			[]uint64{1},
			[]uint64{1},
			[]bool{false},
			[]sdk.Dec{sdk.ZeroDec()},
		},
		{
			"with single validator and multiple superfluid delegations and single undelegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}, {0, "gamm/pool/1"}},
			[]uint64{},
			[]uint64{1},
			[]bool{false},
			[]sdk.Dec{sdk.NewDec(19000000)},
		},
		{
			"with single validator and multiple superfluid delegations and multiple undelegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}, {0, "gamm/pool/1"}},
			[]uint64{},
			[]uint64{1, 2},
			[]bool{false, false},
			[]sdk.Dec{sdk.ZeroDec()},
		},
		{
			"with multiple validators and multiple superfluid delegations and multiple undelegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}, {1, "gamm/pool/1"}},
			[]uint64{},
			[]uint64{1, 2},
			[]bool{false, false},
			[]sdk.Dec{sdk.ZeroDec()},
		},
		{
			"undelegating not available lock id",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}},
			[]uint64{},
			[]uint64{2},
			[]bool{true},
			[]sdk.Dec{sdk.NewDec(19000000)},
		},
		{
			"try undelegating twice for same lock id",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}},
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

			poolId := suite.createGammPool([]string{appparams.BaseCoinUnit, "foo"})
			suite.Require().Equal(poolId, uint64(1))

			// setup validators
			valAddrs := suite.SetupValidators(tc.validatorStats)

			// setup superfluid delegations
			intermediaryAccs, _ := suite.SetupSuperfluidDelegations(valAddrs, tc.superDelegations)
			suite.checkIntermediaryAccountDelegations(intermediaryAccs)

			for _, lockId := range tc.addMoreTokensLockIds {
				lock, err := suite.app.LockupKeeper.GetLockByID(suite.ctx, lockId)
				suite.Require().NoError(err)
				lockOwner, err := sdk.AccAddressFromBech32(lock.Owner)
				suite.Require().NoError(err)
				coins := sdk.Coins{sdk.NewInt64Coin("gamm/pool/1", 1000000)}
				suite.app.BankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, coins)
				suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, minttypes.ModuleName, lockOwner, coins)
				_, err = suite.app.LockupKeeper.AddTokensToLockByID(suite.ctx, lockOwner, lockId, coins)
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

				// superfluid undelegate
				_, err = suite.app.SuperfluidKeeper.SuperfluidUndelegate(suite.ctx, lock.Owner, lockId)
				if tc.expSuperUnbondingErr[index] {
					suite.Require().Error(err)
					continue
				}
				suite.Require().NoError(err)

				// check lockId and intermediary account connection deletion
				addr := suite.app.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(suite.ctx, lockId)
				suite.Require().Equal(addr.String(), "")

				// check bonding synthetic lockup deletion
				_, err = suite.app.LockupKeeper.GetSyntheticLockup(suite.ctx, lockId, keeper.StakingSuffix(valAddr))
				suite.Require().Error(err)

				// check unbonding synthetic lockup creation
				params := suite.app.SuperfluidKeeper.GetParams(suite.ctx)
				synthLock, err := suite.app.LockupKeeper.GetSyntheticLockup(suite.ctx, lockId, keeper.UnstakingSuffix(valAddr))
				suite.Require().NoError(err)
				suite.Require().Equal(synthLock.UnderlyingLockId, lockId)
				suite.Require().Equal(synthLock.Suffix, keeper.UnstakingSuffix(valAddr))
				suite.Require().Equal(synthLock.EndTime, suite.ctx.BlockTime().Add(params.UnbondingDuration))
			}

			// check remaining intermediary account delegation
			for index, expDelegation := range tc.expInterDelegation {
				acc := intermediaryAccs[index]
				valAddr, err := sdk.ValAddressFromBech32(acc.ValAddr)
				suite.Require().NoError(err)
				delegation, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, acc.GetAccAddress(), valAddr)
				if expDelegation.IsZero() {
					suite.Require().False(found)
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

				_, err = suite.app.SuperfluidKeeper.SuperfluidUndelegate(suite.ctx, lock.Owner, lockId)
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestSuperfluidRedelegate() {
	testCases := []struct {
		name                    string
		validatorStats          []stakingtypes.BondStatus
		superDelegations        []superfluidDelegation
		superRedelegations      []superfluidRedelegation
		expSuperRedelegationErr []bool
	}{
		{
			"with single validator and single superfluid delegation with single redelegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}},
			[]superfluidRedelegation{{1, 0, 1}}, // lock1 => val0 -> val1
			[]bool{false},
		},
		{
			"with multiple superfluid delegations with single redelegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}, {0, "gamm/pool/1"}},
			[]superfluidRedelegation{{1, 0, 1}}, // lock1 => val0 -> val1
			[]bool{false},
		},
		{
			"with multiple superfluid delegations with multiple redelegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}, {0, "gamm/pool/1"}},
			[]superfluidRedelegation{{1, 0, 1}, {2, 0, 1}}, // lock1 => val0 -> val1, lock2 => val0 -> val1
			[]bool{false, false},
		},
		{
			"try redelegating back from new validator to original validator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}, {0, "gamm/pool/1"}},
			[]superfluidRedelegation{{1, 0, 1}, {1, 1, 0}}, // lock1 => val0 -> val1, lock1 => val1 -> val0
			[]bool{false, true},
		},
		{
			"not available lock id redelegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}},
			[]superfluidRedelegation{{2, 0, 1}}, // lock1 => val0 -> val1
			[]bool{true},
		},
		{
			"redelegation for same validator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}},
			[]superfluidRedelegation{{1, 0, 0}}, // lock1 => val0 -> val0
			[]bool{true},
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()

			poolId := suite.createGammPool([]string{appparams.BaseCoinUnit, "foo"})
			suite.Require().Equal(poolId, uint64(1))

			// setup validators
			valAddrs := suite.SetupValidators(tc.validatorStats)

			// setup superfluid delegations
			intermediaryAccs, _ := suite.SetupSuperfluidDelegations(valAddrs, tc.superDelegations)
			suite.checkIntermediaryAccountDelegations(intermediaryAccs)

			// execute redelegation and check changes on store
			for index, srd := range tc.superRedelegations {
				lock, err := suite.app.LockupKeeper.GetLockByID(suite.ctx, srd.lockId)
				if err != nil {
					lock = &lockuptypes.PeriodLock{}
				}

				// superfluid redelegate
				err = suite.app.SuperfluidKeeper.SuperfluidRedelegate(suite.ctx, lock.Owner, srd.lockId, valAddrs[srd.newValIndex].String())
				if tc.expSuperRedelegationErr[index] {
					suite.Require().Error(err)
					continue
				}
				suite.Require().NoError(err)

				// check previous validator bonding synthetic lockup deletion
				_, err = suite.app.LockupKeeper.GetSyntheticLockup(suite.ctx, srd.lockId, keeper.StakingSuffix(valAddrs[srd.oldValIndex].String()))
				suite.Require().Error(err)

				// check unbonding synthetic lockup creation
				params := suite.app.SuperfluidKeeper.GetParams(suite.ctx)
				synthLock, err := suite.app.LockupKeeper.GetSyntheticLockup(suite.ctx, srd.lockId, keeper.UnstakingSuffix(valAddrs[srd.oldValIndex].String()))
				suite.Require().NoError(err)
				suite.Require().Equal(synthLock.UnderlyingLockId, srd.lockId)
				suite.Require().Equal(synthLock.Suffix, keeper.UnstakingSuffix(valAddrs[srd.oldValIndex].String()))
				suite.Require().Equal(synthLock.EndTime, suite.ctx.BlockTime().Add(params.UnbondingDuration))

				// check synthetic lockup creation
				synthLock2, err := suite.app.LockupKeeper.GetSyntheticLockup(suite.ctx, srd.lockId, keeper.StakingSuffix(valAddrs[srd.newValIndex].String()))
				suite.Require().NoError(err)
				suite.Require().Equal(synthLock2.UnderlyingLockId, srd.lockId)
				suite.Require().Equal(synthLock2.Suffix, keeper.StakingSuffix(valAddrs[srd.newValIndex].String()))
				suite.Require().Equal(synthLock2.EndTime, time.Time{})

				// check intermediary account creation
				lock, err = suite.app.LockupKeeper.GetLockByID(suite.ctx, srd.lockId)
				suite.Require().NoError(err)

				expAcc := types.NewSuperfluidIntermediaryAccount(lock.Coins[0].Denom, valAddrs[srd.newValIndex].String(), 1)
				gotAcc := suite.app.SuperfluidKeeper.GetIntermediaryAccount(suite.ctx, expAcc.GetAccAddress())
				suite.Require().Equal(gotAcc.Denom, expAcc.Denom)
				suite.Require().Equal(gotAcc.ValAddr, expAcc.ValAddr)

				// check gauge creation
				gauge, err := suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gotAcc.GaugeId)
				suite.Require().NoError(err)
				suite.Require().Equal(gauge.Id, gotAcc.GaugeId)
				suite.Require().Equal(gauge.IsPerpetual, true)
				suite.Require().Equal(gauge.DistributeTo, lockuptypes.QueryCondition{
					LockQueryType: lockuptypes.ByDuration,
					Denom:         expAcc.Denom + keeper.StakingSuffix(valAddrs[srd.newValIndex].String()),
					Duration:      params.UnbondingDuration,
				})
				suite.Require().Equal(gauge.Coins, sdk.Coins(nil))
				suite.Require().Equal(gauge.StartTime, suite.ctx.BlockTime())
				suite.Require().Equal(gauge.NumEpochsPaidOver, uint64(1))
				suite.Require().Equal(gauge.FilledEpochs, uint64(0))
				suite.Require().Equal(gauge.DistributedCoins, sdk.Coins(nil))

				// Check lockID connection with intermediary account
				intAcc := suite.app.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(suite.ctx, srd.lockId)
				suite.Require().Equal(intAcc.String(), expAcc.GetAccAddress().String())

				// check delegation from intermediary account to validator
				_, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, expAcc.GetAccAddress(), valAddrs[srd.newValIndex])
				suite.Require().True(found)
			}

			// try redelegating twice
			for index, srd := range tc.superRedelegations {
				if tc.expSuperRedelegationErr[index] {
					continue
				}
				cacheCtx, _ := suite.ctx.CacheContext()
				lock, err := suite.app.LockupKeeper.GetLockByID(suite.ctx, srd.lockId)
				suite.Require().NoError(err)
				err = suite.app.SuperfluidKeeper.SuperfluidRedelegate(cacheCtx, lock.Owner, srd.lockId, valAddrs[srd.newValIndex].String())
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestRefreshIntermediaryDelegationAmounts() {
	testCases := []struct {
		name             string
		validatorStats   []stakingtypes.BondStatus
		superDelegations []superfluidDelegation
		roundOneTwaps    []assetTwap
		roundTwoTwaps    []assetTwap
		checkAccIndexes  []int64
	}{
		{
			"with single validator and single delegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}},
			[]assetTwap{{"gamm/pool/1", sdk.NewDec(10)}},
			[]assetTwap{},
			[]int64{0},
		},
		{
			"with single validator and multiple delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}, {0, "gamm/pool/1"}},
			[]assetTwap{{"gamm/pool/1", sdk.NewDec(10)}},
			[]assetTwap{},
			[]int64{0},
		},
		{
			"with multiple validator and multiple superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}, {1, "gamm/pool/1"}},
			[]assetTwap{{"gamm/pool/1", sdk.NewDec(10)}},
			[]assetTwap{},
			[]int64{0, 1},
		},
		{
			"with single validator and multiple denom superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}, {0, "gamm/pool/2"}},
			[]assetTwap{{"gamm/pool/1", sdk.NewDec(10)}, {"gamm/pool/2", sdk.NewDec(10)}},
			[]assetTwap{},
			[]int64{0, 1},
		},
		{
			"with multiple validators and multiple denom superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}, {1, "gamm/pool/2"}},
			[]assetTwap{{"gamm/pool/1", sdk.NewDec(10)}, {"gamm/pool/2", sdk.NewDec(10)}},
			[]assetTwap{},
			[]int64{0, 1},
		},
		{
			"zero price twap check",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}},
			[]assetTwap{{"gamm/pool/1", sdk.NewDec(0)}},
			[]assetTwap{},
			[]int64{0},
		},
		{
			"refresh case from zero to non-zero",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}},
			[]assetTwap{{"gamm/pool/1", sdk.NewDec(0)}},
			[]assetTwap{{"gamm/pool/1", sdk.NewDec(10)}},
			[]int64{0},
		},
		{
			"dust price twap check",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}},
			[]assetTwap{{"gamm/pool/1", sdk.NewDecWithPrec(1, 10)}}, // 10^-10
			[]assetTwap{},
			[]int64{0},
		},
		{
			"refresh case from dust to non-dust",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}},
			[]assetTwap{{"gamm/pool/1", sdk.NewDecWithPrec(1, 10)}}, // 10^-10
			[]assetTwap{{"gamm/pool/1", sdk.NewDec(10)}},
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
			valAddrs := suite.SetupValidators(tc.validatorStats)
			// setup superfluid delegations
			intermediaryAccs, locks := suite.SetupSuperfluidDelegations(valAddrs, tc.superDelegations)
			suite.checkIntermediaryAccountDelegations(intermediaryAccs)
			intermediaryDels := []sdk.Dec{}

			for _, intAccIndex := range tc.checkAccIndexes {
				expAcc := intermediaryAccs[intAccIndex]
				valAddr, err := sdk.ValAddressFromBech32(expAcc.ValAddr)
				suite.Require().NoError(err)

				// check delegation from intermediary account to validator
				delegation, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, expAcc.GetAccAddress(), valAddr)
				suite.Require().True(found)
				intermediaryDels = append(intermediaryDels, delegation.Shares)
			}

			// twap price change before refresh
			twapByDenom := make(map[string]sdk.Dec)
			for _, twap := range tc.roundOneTwaps {
				twapByDenom[twap.denom] = twap.price
				suite.app.SuperfluidKeeper.SetEpochOsmoEquivalentTWAP(suite.ctx, 2, twap.denom, twap.price)
			}

			suite.app.EpochsKeeper.SetEpochInfo(suite.ctx, epochstypes.EpochInfo{
				Identifier:   params.RefreshEpochIdentifier,
				CurrentEpoch: 2,
			})

			// refresh intermediary account delegations
			suite.NotPanics(func() {
				suite.app.SuperfluidKeeper.RefreshIntermediaryDelegationAmounts(suite.ctx)
			})

			originTwap := sdk.NewDec(20)
			targetDelegations := []sdk.Dec{}
			targetAmounts := []sdk.Int{}
			for index, intAccIndex := range tc.checkAccIndexes {
				expAcc := intermediaryAccs[intAccIndex]
				twap, ok := twapByDenom[expAcc.Denom]
				if !ok {
					twap = originTwap
				}

				targetDelegation := intermediaryDels[index].Mul(twap).Quo(originTwap)
				lpTokenAmount := sdk.NewInt(1000000)
				decAmt := twap.Mul(lpTokenAmount.ToDec())
				asset := suite.app.SuperfluidKeeper.GetSuperfluidAsset(suite.ctx, expAcc.Denom)
				targetAmount := suite.app.SuperfluidKeeper.GetRiskAdjustedOsmoValue(suite.ctx, asset, decAmt.RoundInt())

				targetDelegations = append(targetDelegations, targetDelegation)
				targetAmounts = append(targetAmounts, targetAmount)
			}

			for index, intAccIndex := range tc.checkAccIndexes {
				expAcc := intermediaryAccs[intAccIndex]
				valAddr, err := sdk.ValAddressFromBech32(expAcc.ValAddr)
				suite.Require().NoError(err)

				targetAmount := targetAmounts[index]
				targetDelegation := targetDelegations[index]

				// check delegation changes
				delegation, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, expAcc.GetAccAddress(), valAddr)
				if targetAmount.IsPositive() {
					suite.Require().True(found)
					suite.Require().Equal(delegation.Shares, targetDelegation)
				} else {
					suite.Require().False(found)
				}
			}

			// start new epoch
			suite.app.EpochsKeeper.SetEpochInfo(suite.ctx, epochstypes.EpochInfo{
				Identifier:   params.RefreshEpochIdentifier,
				CurrentEpoch: 3,
			})

			// if roundTwo twaps exists, execute round two twaps and finish tests
			if len(tc.roundTwoTwaps) > 0 {
				twap2ByDenom := make(map[string]sdk.Dec)
				for _, twap := range tc.roundTwoTwaps {
					twap2ByDenom[twap.denom] = twap.price
					suite.app.SuperfluidKeeper.SetEpochOsmoEquivalentTWAP(suite.ctx, 3, twap.denom, twap.price)
				}
				// refresh intermediary account delegations
				suite.NotPanics(func() {
					suite.app.SuperfluidKeeper.RefreshIntermediaryDelegationAmounts(suite.ctx)
				})

				for index, intAccIndex := range tc.checkAccIndexes {
					expAcc := intermediaryAccs[intAccIndex]
					valAddr, err := sdk.ValAddressFromBech32(expAcc.ValAddr)
					suite.Require().NoError(err)

					targetDelegation := intermediaryDels[index].Mul(twap2ByDenom[expAcc.Denom]).Quo(originTwap)

					// check delegation changes
					delegation, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, expAcc.GetAccAddress(), valAddr)

					suite.Require().True(found)
					suite.Require().Equal(delegation.Shares, targetDelegation)
				}
				return
			}

			// unbond all lockups
			for _, lock := range locks {
				// superfluid undelegate
				_, err := suite.app.SuperfluidKeeper.SuperfluidUndelegate(suite.ctx, lock.Owner, lock.ID)
				suite.Require().NoError(err)
			}

			// check intermediary account changes after unbonding operations
			for index, intAccIndex := range tc.checkAccIndexes {
				expAcc := intermediaryAccs[intAccIndex]
				suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(params.UnbondingDuration + time.Second))
				suite.app.EndBlocker(suite.ctx, abci.RequestEndBlock{Height: suite.ctx.BlockHeight()})

				targetAmount := targetAmounts[index]

				unbonded := suite.app.BankKeeper.GetBalance(suite.ctx, expAcc.GetAccAddress(), sdk.DefaultBondDenom)
				if targetAmount.IsPositive() {
					suite.Require().True(unbonded.IsPositive())
				} else {
					suite.Require().True(unbonded.IsZero())
				}
			}

			// refresh intermediary account delegations
			suite.NotPanics(func() {
				suite.app.SuperfluidKeeper.RefreshIntermediaryDelegationAmounts(suite.ctx)
			})

			// check changes after refresh operation
			for _, intAccIndex := range tc.checkAccIndexes {
				expAcc := intermediaryAccs[intAccIndex]
				// check unbonded amount is removed after refresh operation
				refreshed := suite.app.BankKeeper.GetBalance(suite.ctx, expAcc.GetAccAddress(), sdk.DefaultBondDenom)
				suite.Require().True(refreshed.IsZero())
			}
		})
	}
}
