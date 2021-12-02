package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	epochstypes "github.com/osmosis-labs/osmosis/x/epochs/types"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
	"github.com/osmosis-labs/osmosis/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/x/superfluid/types"
)

// _, app, ctx := createTestInput()
// delTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 10)
// delCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), delTokens))

// addrDels := simapp.AddTestAddrsIncremental(app, ctx, 2, sdk.NewInt(0))
// addrVals := simapp.ConvertAddrsToValAddrs(addrDels)

// //create a validator with a self-delegation
// validator := teststaking.NewValidator(t, addrVals[0], PKs[0])
// app.StakingKeeper.SetValidatorByConsAddr(ctx, validator)

// validator, issuedShares := validator.AddTokensFromDel(delTokens)
// require.Equal(t, delTokens, issuedShares.RoundInt())

// // add bonded tokens to pool for delegations
// notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)
// require.NoError(t, simapp.FundModuleAccount(app.BankKeeper, ctx, notBondedPool.GetName(), delCoins))
// app.AccountKeeper.SetModuleAccount(ctx, notBondedPool)

// validator = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validator, true)
// require.True(t, validator.IsBonded())

// selfDelegation := types.NewDelegation(addrVals[0].Bytes(), addrVals[0], issuedShares)
// app.StakingKeeper.SetDelegation(ctx, selfDelegation)

// // add bonded tokens to pool for delegations
// bondedPool := app.StakingKeeper.GetBondedPool(ctx)
// require.NoError(t, simapp.FundModuleAccount(app.BankKeeper, ctx, bondedPool.GetName(), delCoins))
// app.AccountKeeper.SetModuleAccount(ctx, bondedPool)

// // create a second delegation to this validator
// app.StakingKeeper.DeleteValidatorByPowerIndex(ctx, validator)

// validator, issuedShares = validator.AddTokensFromDel(delTokens)
// require.Equal(t, delTokens, issuedShares.RoundInt())

// require.NoError(t, simapp.FundModuleAccount(app.BankKeeper, ctx, bondedPool.GetName(), delCoins))
// app.AccountKeeper.SetModuleAccount(ctx, bondedPool)

// validator = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validator, true)
// delegation := types.NewDelegation(addrDels[1], addrVals[0], issuedShares)
// app.StakingKeeper.SetDelegation(ctx, delegation)

// require.NoError(t, simapp.FundModuleAccount(app.BankKeeper, ctx, bondedPool.GetName(), delCoins))
// app.AccountKeeper.SetModuleAccount(ctx, bondedPool)

// header := ctx.BlockHeader()
// blockHeight := int64(10)
// header.Height = blockHeight
// blockTime := time.Unix(333, 0)
// header.Time = blockTime
// ctx = ctx.WithBlockHeader(header)

// // unbond the all self-delegation to put validator in unbonding state
// val0AccAddr := sdk.AccAddress(addrVals[0])
// _, err := app.StakingKeeper.Undelegate(ctx, val0AccAddr, addrVals[0], delTokens.ToDec())
// require.NoError(t, err)

// // end block
// applyValidatorSetUpdates(t, ctx, app.StakingKeeper, 1)

// validator, found := app.StakingKeeper.GetValidator(ctx, addrVals[0])
// require.True(t, found)
// require.Equal(t, blockHeight, validator.UnbondingHeight)
// params := app.StakingKeeper.GetParams(ctx)
// require.True(t, blockTime.Add(params.UnbondingTime).Equal(validator.UnbondingTime))

// blockHeight2 := int64(20)
// blockTime2 := time.Unix(444, 0).UTC()
// ctx = ctx.WithBlockHeight(blockHeight2)
// ctx = ctx.WithBlockTime(blockTime2)

// // unbond some of the other delegation's shares
// _, err = app.StakingKeeper.Undelegate(ctx, addrDels[1], addrVals[0], sdk.NewDec(6))
// require.NoError(t, err)

// // retrieve the unbonding delegation
// ubd, found := app.StakingKeeper.GetUnbondingDelegation(ctx, addrDels[1], addrVals[0])
// require.True(t, found)
// require.Len(t, ubd.Entries, 1)
// require.True(t, ubd.Entries[0].Balance.Equal(sdk.NewInt(6)))
// assert.Equal(t, blockHeight2, ubd.Entries[0].CreationHeight)
// assert.True(t, blockTime2.Add(params.UnbondingTime).Equal(ubd.Entries[0].CompletionTime))

func (suite *KeeperTestSuite) LockTokens(addr sdk.AccAddress, coins sdk.Coins, duration time.Duration) lockuptypes.PeriodLock {
	err := suite.app.BankKeeper.SetBalances(suite.ctx, addr, coins)
	suite.Require().NoError(err)
	lock, err := suite.app.LockupKeeper.LockTokens(suite.ctx, addr, coins, duration)
	suite.Require().NoError(err)
	return lock
}

func (suite *KeeperTestSuite) TestSuperfluidDelegate() {
	suite.SetupTest()
	suite.app.IncentivesKeeper.SetLockableDurations(suite.ctx, []time.Duration{
		time.Hour * 24 * 14,
	})

	// create a validator
	valPub := secp256k1.GenPrivKey().PubKey()
	valAddr := sdk.ValAddress(valPub.Address())

	validator, err := stakingtypes.NewValidator(valAddr, valPub, stakingtypes.NewDescription("moniker", "", "", "", ""))
	suite.Require().NoError(err)

	amount := sdk.NewInt(1000000)
	issuedShares := amount.ToDec()
	validator.Tokens = validator.Tokens.Add(amount)
	validator.DelegatorShares = validator.DelegatorShares.Add(issuedShares)

	suite.app.StakingKeeper.SetValidator(suite.ctx, validator)
	suite.app.StakingKeeper.SetValidatorByConsAddr(suite.ctx, validator)
	suite.app.StakingKeeper.SetValidatorByPowerIndex(suite.ctx, validator)
	suite.app.StakingKeeper.AfterValidatorCreated(suite.ctx, validator.GetOperator())

	// register a LP token as a superfluid asset
	suite.app.SuperfluidKeeper.SetSuperfluidAsset(suite.ctx, types.SuperfluidAsset{
		Denom:     "lptoken",
		AssetType: types.SuperfluidAssetTypeLPShare,
	})

	// set OSMO TWAP price for LP token
	suite.app.SuperfluidKeeper.SetEpochOsmoEquivalentTWAP(suite.ctx, 1, "lptoken", sdk.NewDec(2))
	params := suite.app.SuperfluidKeeper.GetParams(suite.ctx)
	suite.app.EpochsKeeper.SetEpochInfo(suite.ctx, epochstypes.EpochInfo{
		Identifier:   params.RefreshEpochIdentifier,
		CurrentEpoch: 2,
	})

	// create lockup of LP token
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}
	lock := suite.LockTokens(addr1, coins, time.Hour*24*14)

	// call SuperfluidDelegate and check response
	err = suite.app.SuperfluidKeeper.SuperfluidDelegate(suite.ctx, lock.ID, valAddr.String())
	suite.Require().NoError(err)

	// check synthetic lockup creation
	synthLock, err := suite.app.LockupKeeper.GetSyntheticLockup(suite.ctx, lock.ID, keeper.StakingSuffix(valAddr.String()))
	suite.Require().NoError(err)
	suite.Require().Equal(synthLock.LockId, lock.ID)
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
	suite.Require().Equal(gotAcc.GaugeId, uint64(1))

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

	// TODO: Check lockID connection with intermediary account
	// TODO: Check delegation from intermediary account to validator
	// TODO: add table driven test for all edge cases
}

func (suite *KeeperTestSuite) TestSuperfluidUndelegate() {
	// TODO: do SuperfluidDelgate to test undelegation - utility function
	// TODO: add test for SuperfluidUndelegate
	// TODO: Check superfluid delegate result error
	// TODO: check synthetic lockup deletion for delegation
	// TODO: check unbonding synthetic lockup creation
}

func (suite *KeeperTestSuite) TestSuperfluidRedelegate() {
	// TODO: do SuperfluidDelgate to test undelegation - utility function
	// TODO: add test for SuperfluidRedelegate
	// TODO: check the changes for undelegate function call changes
	// TODO: check the changes for new delegation function call changes
}

func (suite *KeeperTestSuite) TestRefreshIntermediaryDelegationAmounts() {
	// TODO: add test for refreshIntermediaryDelegationAmounts
}
