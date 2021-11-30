package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
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

func (suite *KeeperTestSuite) TestSuperfluidDelegate() {

	// create a validator
	valPub := secp256k1.GenPrivKey().PubKey()
	valAddr := sdk.ValAddress(valPub.Address())

	validator, err := stakingtypes.NewValidator(valAddr, valPub, stakingtypes.NewDescription("moniker", "", "", "", ""))
	suite.Require().NoError(err)
	suite.app.StakingKeeper.SetValidatorByConsAddr(suite.ctx, validator)

	// register a LP token as a superfluid asset
	suite.app.SuperfluidKeeper.SetSuperfluidAsset(suite.ctx, types.SuperfluidAsset{
		Denom:     "lptoken",
		AssetType: types.SuperfluidAssetTypeLPShare,
	})

	// TODO: set OSMO TWAP price for LP token
	// TODO: create lockup of LP token
	// TODO: call SuperfluidDelegate
	// TODO: Check superfluid delegate result
	// TODO: Check synthetic lockup creation
	// TODO: Check intermediary account creation
	// TODO: Check gauge creation
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
