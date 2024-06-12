package keeper_test

import (
	gocontext "context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	appparams "github.com/osmosis-labs/osmosis/v25/app/params"
	balancertypes "github.com/osmosis-labs/osmosis/v25/x/gamm/pool-models/balancer"
	minttypes "github.com/osmosis-labs/osmosis/v25/x/mint/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v25/x/poolmanager/types"
	"github.com/osmosis-labs/osmosis/v25/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/v25/x/superfluid/types"
)

type TestSuite struct {
	KeeperTestSuite

	proposal govv1.Proposal
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) SetupTest() {
	s.KeeperTestSuite.SetupTest()

	// set the bond denom to be osmo (because it's hardcoded in protorev)
	stakingParams, err := s.App.StakingKeeper.GetParams(s.Ctx)
	s.Require().NoError(err)
	stakingParams.BondDenom = appparams.BaseCoinUnit
	err = s.App.StakingKeeper.SetParams(s.Ctx, stakingParams)
	s.Require().NoError(err)

	// set incentives min value in osmo
	incentivesParams := s.App.IncentivesKeeper.GetParams(s.Ctx)
	incentivesParams.MinValueForDistribution.Denom = appparams.BaseCoinUnit
	s.App.IncentivesKeeper.SetParams(s.Ctx, incentivesParams)

	// make pool creation fees be paid in the bond denom. Also make them low.
	poolmanagerParams := s.App.PoolManagerKeeper.GetParams(s.Ctx)
	bondDenom, err := s.App.StakingKeeper.BondDenom(s.Ctx)
	s.Require().NoError(err)
	poolmanagerParams.PoolCreationFee = sdk.NewCoins(sdk.NewInt64Coin(bondDenom, 1))
	s.App.PoolManagerKeeper.SetParams(s.Ctx, poolmanagerParams)

	// create a proposal
	// Populate the gov keeper in advance with an active proposal
	testProposal := &govtypes.TextProposal{
		Title:       "IBC Gov Proposal",
		Description: "tokens for all!",
	}

	proposalMsg, err := govv1.NewLegacyContent(testProposal, "")
	s.Require().NoError(err)

	proposal, err := govv1.NewProposal(
		[]sdk.Msg{proposalMsg},
		govtypes.DefaultStartingProposalID,
		s.Ctx.BlockTime(),
		s.Ctx.BlockTime(),
		"test proposal",
		"title",
		"Description",
		sdk.AccAddress("proposer"),
		false,
	)
	s.Require().NoError(err)
	s.App.GovKeeper.SetProposal(s.Ctx, proposal)
	s.App.GovKeeper.ActivateVotingPeriod(s.Ctx, proposal)
	s.proposal = proposal
}

func createPoolMsgGen(sender sdk.AccAddress, assets sdk.Coins) *balancertypes.MsgCreateBalancerPool {
	if len(assets) != 2 {
		panic("baseCreatePoolMsg requires 2 assets")
	}
	poolAssets := []balancertypes.PoolAsset{
		{
			Weight: osmomath.NewInt(1),
			Token:  assets[0],
		},
		{
			Weight: osmomath.NewInt(1),
			Token:  assets[1],
		},
	}

	poolParams := &balancertypes.PoolParams{
		SwapFee: osmomath.NewDecWithPrec(1, 2),
		ExitFee: osmomath.ZeroDec(),
	}

	msg := &balancertypes.MsgCreateBalancerPool{
		Sender:             sender.String(),
		PoolAssets:         poolAssets,
		PoolParams:         poolParams,
		FuturePoolGovernor: "",
	}

	return msg
}

func (s *TestSuite) mintToAccount(amount osmomath.Int, denom string, acc sdk.AccAddress) {
	err := s.App.BankKeeper.MintCoins(s.Ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin(denom, amount)))
	s.Require().NoError(err)
	// send the coins to user1
	err = s.App.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, minttypes.ModuleName, acc, sdk.NewCoins(sdk.NewCoin(denom, amount)))
	s.Require().NoError(err)
}

func (s *TestSuite) TestGammSuperfluid() {
	////////
	// Setup
	////////

	s.SetupTest()

	// denoms
	btcDenom := "factory/osmo1pfyxruwvtwk00y8z06dh2lqjdj82ldvy74wzm3/allBTC" // Asset to superfluid stake
	bondDenom, err := s.App.StakingKeeper.BondDenom(s.Ctx)
	s.Require().NoError(err)

	// accounts
	// pool creator
	lpKey := ed25519.GenPrivKey().PubKey()
	lpAddr := sdk.AccAddress(lpKey.Address())

	osmoPoolAmount := osmomath.NewInt(1_000_000_000_000)
	btcPoolAmount := osmomath.NewInt(10_000_000_000)
	// default bond denom

	// mint necessary tokens
	s.mintToAccount(btcPoolAmount, btcDenom, lpAddr)
	s.mintToAccount(osmoPoolAmount.Mul(osmomath.NewInt(2)), bondDenom, lpAddr)

	nextPoolId := s.App.PoolManagerKeeper.GetNextPoolId(s.Ctx) // the pool id we'll create

	// create an bondDenom/btcDenom pool
	createPoolMsg := createPoolMsgGen(
		lpAddr,
		sdk.NewCoins(sdk.NewCoin(btcDenom, btcPoolAmount), sdk.NewCoin(bondDenom, osmoPoolAmount)),
	)

	_, err = s.RunMsg(createPoolMsg)
	s.Require().NoError(err)
	gammToken := fmt.Sprintf("gamm/pool/%d", nextPoolId)

	totalGammTokens := s.App.BankKeeper.GetBalance(s.Ctx, lpAddr, gammToken)

	// Add btcDenom as an allowed superfluid asset
	err = s.App.SuperfluidKeeper.AddNewSuperfluidAsset(s.Ctx, types.SuperfluidAsset{Denom: gammToken, AssetType: types.SuperfluidAssetTypeLPShare})
	s.Require().NoError(err)

	// Mint assets to the lockup module. This will ensure there are assets to distribute.
	err = s.App.BankKeeper.MintCoins(s.Ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin(bondDenom, osmomath.NewInt(1_000_000_000))))
	s.Require().NoError(err)
	err = s.App.BankKeeper.SendCoinsFromModuleToModule(s.Ctx, minttypes.ModuleName, authtypes.FeeCollectorName, sdk.NewCoins(sdk.NewCoin(bondDenom, osmomath.NewInt(1_000_000_000))))
	s.Require().NoError(err)

	// Keep track of the original balance of the bond denom to make sure rewards are distributed later on
	originalBondDenomBalance := s.App.BankKeeper.GetBalance(s.Ctx, lpAddr, bondDenom).Amount

	////////
	// TEST: Delegate gamm tokens
	////////

	// No delegations
	delegations := s.App.LockupKeeper.GetAllSyntheticLockupsByAddr(s.Ctx, lpAddr)
	s.Require().Equal(0, len(delegations))

	// superfluid stake gamm token
	validators, err := s.App.StakingKeeper.GetAllValidators(s.Ctx)
	s.Require().NoError(err)
	validator := validators[0]
	gammDelegationAmount := osmomath.NewInt(1000000000000000000)
	delegateMsg := &types.MsgLockAndSuperfluidDelegate{
		Sender:  lpAddr.String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(gammToken, gammDelegationAmount)),
		ValAddr: validator.GetOperator(),
	}
	_, err = s.RunMsg(delegateMsg)
	s.Require().NoError(err)

	// Check delegations
	delegations = s.App.LockupKeeper.GetAllSyntheticLockupsByAddr(s.Ctx, lpAddr)
	s.Require().Equal(1, len(delegations))
	synthLock := delegations[0]

	// Get underlying lock
	underlyingLock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, synthLock.UnderlyingLockId)
	s.Require().NoError(err)
	s.Require().Equal(lpAddr.String(), underlyingLock.Owner)
	s.Require().Equal(gammToken, underlyingLock.Coins[0].Denom)

	remainingGammTokens := s.App.BankKeeper.GetBalance(s.Ctx, lpAddr, gammToken)
	s.Require().Equal(totalGammTokens.Amount.Sub(gammDelegationAmount), remainingGammTokens.Amount)

	queryDelegations := types.SuperfluidDelegationsByDelegatorRequest{DelegatorAddress: lpAddr.String()}
	querier := keeper.NewQuerier(*s.App.SuperfluidKeeper)
	res, err := querier.SuperfluidDelegationsByDelegator(s.Ctx, &queryDelegations)
	s.Require().NoError(err)
	s.Require().Len(res.SuperfluidDelegationRecords, 1)
	s.Require().Equal(lpAddr.String(), res.SuperfluidDelegationRecords[0].DelegatorAddress)
	s.Require().Equal(validator.GetOperator(), res.SuperfluidDelegationRecords[0].ValidatorAddress)
	s.Require().Equal(gammToken, res.SuperfluidDelegationRecords[0].DelegationAmount.Denom)
	s.Require().Equal(gammDelegationAmount, res.SuperfluidDelegationRecords[0].DelegationAmount.Amount)
	s.Require().Equal(appparams.BaseCoinUnit, res.SuperfluidDelegationRecords[0].EquivalentStakedAmount.Denom)
	riskFactor := s.App.SuperfluidKeeper.CalculateRiskFactor(s.Ctx, gammToken)
	multiplier := s.App.SuperfluidKeeper.GetOsmoEquivalentMultiplier(s.Ctx, gammToken)
	equivalentAmount := riskFactor.Mul(osmomath.NewDec(gammDelegationAmount.Int64())).Mul(multiplier)
	fmt.Println("riskFactor", riskFactor)
	fmt.Println("equivalentAmount", equivalentAmount)
	fmt.Println("res.SuperfluidDelegationRecords[0].EquivalentStakedAmount.Amount", res.SuperfluidDelegationRecords[0].EquivalentStakedAmount.Amount)
	fmt.Println("res.SuperfluidDelegationRecords[0].EquivalentStakedAmount.Amount.Int64()", res.SuperfluidDelegationRecords[0].EquivalentStakedAmount.Amount.Int64())
	s.Require().Equal(equivalentAmount, osmomath.NewDec(res.SuperfluidDelegationRecords[0].EquivalentStakedAmount.Amount.Int64()))
	s.Require().Equal(appparams.BaseCoinUnit, res.TotalEquivalentStakedAmount.Denom)
	s.Require().Equal(equivalentAmount, osmomath.NewDec(res.TotalEquivalentStakedAmount.Amount.Int64()))
	s.Require().Equal(appparams.BaseCoinUnit, res.TotalEquivalentNonOsmoStakedAmount.Denom)
	s.Require().Equal(osmomath.NewDec(0), osmomath.NewDec(res.TotalEquivalentNonOsmoStakedAmount.Amount.Int64()))

	////////
	// TEST: Reward distribution
	////////

	// move time beyond the needed time for rewards to be distributed
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(49 * time.Hour))

	// Check that the user has not received any rewards yet
	bondDenomBalance := s.App.BankKeeper.GetBalance(s.Ctx, lpAddr, bondDenom)
	s.Require().Equal(originalBondDenomBalance, bondDenomBalance.Amount)

	// There are no rewards assigned to the validator yet
	validatorRewards := new(distrtypes.QueryValidatorOutstandingRewardsResponse)
	err = s.QueryHelper.Invoke(gocontext.Background(),
		"/cosmos.distribution.v1beta1.Query/ValidatorOutstandingRewards",
		&distrtypes.QueryValidatorOutstandingRewardsRequest{
			ValidatorAddress: validator.GetOperator(),
		},
		validatorRewards)
	s.Require().Equal(0, len(validatorRewards.Rewards.Rewards))
	s.Require().NoError(err)

	// Move to block 50 because rewards are only distributed every 50 blocks. Rewards will be available after unstaking
	s.AdvanceToBlockNAndRunEpoch(50)

	// After a block that is not a multiple of 50, the rewards will be assigned to the validator
	validatorRewards = new(distrtypes.QueryValidatorOutstandingRewardsResponse)
	err = s.QueryHelper.Invoke(gocontext.Background(),
		"/cosmos.distribution.v1beta1.Query/ValidatorOutstandingRewards",
		&distrtypes.QueryValidatorOutstandingRewardsRequest{
			ValidatorAddress: validator.GetOperator(),
		},
		validatorRewards)
	s.Require().NoError(err)
	s.Require().Equal(2, len(validatorRewards.Rewards.Rewards))

	////////
	// TEST:  Check delegation rewards were distributed
	////////
	bondDenomBalance = s.App.BankKeeper.GetBalance(s.Ctx, lpAddr, bondDenom)
	s.Require().True(bondDenomBalance.Amount.GT(originalBondDenomBalance))

	////////
	// TEST: Voting. User can vote
	////////

	// Reset the voting period
	s.App.GovKeeper.ActivateVotingPeriod(s.Ctx, s.proposal)

	// check user can vote
	voteMsg := &govtypes.MsgVote{
		ProposalId: 1,
		Voter:      lpAddr.String(),
		Option:     govtypes.OptionYes,
	}
	_, err = s.RunMsg(voteMsg)
	s.Require().NoError(err)

	// Move time beyond voting end time
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(96 * time.Hour))
	s.EndBlock()
	s.BeginNewBlock(true)

	proposal, err := s.App.GovKeeper.Proposals.Get(s.Ctx, 1)
	s.Require().NoError(err)
	s.Require().Equal(govv1.StatusFailed, proposal.Status)
	s.Require().Equal("5000000000", proposal.FinalTallyResult.YesCount)

	////////
	// TEST: Unstake
	////////

	// Check that the user can unstake and the delegation is removed
	undelegateMsg := &types.MsgSuperfluidUndelegateAndUnbondLock{
		Sender: lpAddr.String(),
		LockId: underlyingLock.ID,
		Coin:   sdk.NewCoin(gammToken, gammDelegationAmount),
	}
	_, err = s.RunMsg(undelegateMsg)
	s.Require().NoError(err)

	// Check delegations
	queryDelegations = types.SuperfluidDelegationsByDelegatorRequest{DelegatorAddress: lpAddr.String()}
	res, err = querier.SuperfluidDelegationsByDelegator(s.Ctx, &queryDelegations)
	s.Require().NoError(err)
	s.Require().Len(res.SuperfluidDelegationRecords, 0)

	// Check undelegations
	queryUndelegations := types.SuperfluidUndelegationsByDelegatorRequest{DelegatorAddress: lpAddr.String()}
	undelegationResponse, err := querier.SuperfluidUndelegationsByDelegator(s.Ctx, &queryUndelegations)
	s.Require().NoError(err)
	s.Require().Len(undelegationResponse.SuperfluidDelegationRecords, 1)

	// check pool token balance before undelegation time passes. Should be the same as before
	balance := s.App.BankKeeper.GetBalance(s.Ctx, lpAddr, gammToken)
	s.Require().Equal(remainingGammTokens.Amount, balance.Amount)

	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(undelegationResponse.SyntheticLocks[0].Duration + time.Second))
	// move forward to block 60 because we only check matured locks every 30 blocks
	s.AdvanceToBlockNAndRunEpoch(60)

	// No more undelegations
	undelegationResponse, err = querier.SuperfluidUndelegationsByDelegator(s.Ctx, &queryUndelegations)
	s.Require().NoError(err)
	s.Require().Len(undelegationResponse.SuperfluidDelegationRecords, 0)

	// check pool token balance after undelegation time passes. Should be back to original
	balance = s.App.BankKeeper.GetBalance(s.Ctx, lpAddr, gammToken)
	s.Require().Equal(totalGammTokens.Amount, balance.Amount)
}

func (s *TestSuite) TestNativeSuperfluid() {
	//
	// TEST: Setup
	//

	s.SetupTest()

	// Set the mint denom to be osmo
	params := s.App.MintKeeper.GetParams(s.Ctx)
	params.MintDenom = appparams.BaseCoinUnit
	s.App.MintKeeper.SetParams(s.Ctx, params)

	// denoms
	btcDenom := "factory/osmo1pfyxruwvtwk00y8z06dh2lqjdj82ldvy74wzm3/allBTC" // Asset to superfluid stake
	bondDenom, err := s.App.StakingKeeper.BondDenom(s.Ctx)
	s.Require().NoError(err)

	// accounts
	// pool creator
	lpKey := ed25519.GenPrivKey().PubKey()
	poolAddr := sdk.AccAddress(lpKey.Address())
	userKey := ed25519.GenPrivKey().PubKey()
	userAddr := sdk.AccAddress(userKey.Address())

	osmoPoolAmount := osmomath.NewInt(1_000_000_000_000)
	btcPoolAmount := osmomath.NewInt(10_000_000_000)
	// default bond denom

	// mint necessary tokens
	s.mintToAccount(btcPoolAmount, btcDenom, poolAddr)
	s.mintToAccount(osmoPoolAmount.Mul(osmomath.NewInt(2)), bondDenom, poolAddr)
	s.mintToAccount(osmomath.NewInt(100_000_000), bondDenom, userAddr)
	totalBTCAmount := osmomath.NewInt(1_000_000)
	s.mintToAccount(totalBTCAmount, btcDenom, userAddr)

	nextPoolId := s.App.PoolManagerKeeper.GetNextPoolId(s.Ctx) // the pool id we'll create

	// create an bondDenom/btcDenom pool. This is only used so that the native asset can have a price.
	createPoolMsg := createPoolMsgGen(
		poolAddr,
		sdk.NewCoins(sdk.NewCoin(btcDenom, btcPoolAmount), sdk.NewCoin(bondDenom, osmoPoolAmount)),
	)

	_, err = s.RunMsg(createPoolMsg)
	s.Require().NoError(err)

	// move time forward and advance a few blocks to get twaps
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(20 * time.Minute))
	s.AdvanceToBlockNAndRunEpoch(5)

	// Creating a native type without a pool should fail
	err = s.App.SuperfluidKeeper.AddNewSuperfluidAsset(s.Ctx, types.SuperfluidAsset{Denom: btcDenom, AssetType: types.SuperfluidAssetTypeNative})
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "failed to get twap price")

	// Creating a native type with a non-existing pool should fail
	err = s.App.SuperfluidKeeper.AddNewSuperfluidAsset(s.Ctx, types.SuperfluidAsset{Denom: btcDenom, AssetType: types.SuperfluidAssetTypeNative, PriceRoute: []*poolmanagertypes.SwapAmountInRoute{{PoolId: nextPoolId + 10, TokenOutDenom: bondDenom}}})
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "failed to get twap price")

	// Add btcDenom as an allowed superfluid asset
	err = s.App.SuperfluidKeeper.AddNewSuperfluidAsset(s.Ctx, types.SuperfluidAsset{Denom: btcDenom, AssetType: types.SuperfluidAssetTypeNative, PriceRoute: []*poolmanagertypes.SwapAmountInRoute{{PoolId: nextPoolId, TokenOutDenom: bondDenom}}})
	s.Require().NoError(err)

	// Mint assets to the lockup module. This will ensure there are assets to distribute.
	err = s.App.BankKeeper.MintCoins(s.Ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin(bondDenom, osmomath.NewInt(1_000_000_000))))
	s.Require().NoError(err)
	err = s.App.BankKeeper.SendCoinsFromModuleToModule(s.Ctx, minttypes.ModuleName, authtypes.FeeCollectorName, sdk.NewCoins(sdk.NewCoin(bondDenom, osmomath.NewInt(1_000_000_000))))
	s.Require().NoError(err)

	// Keep track of the original balance of the bond denom to make sure rewards are distributed later on
	originalBondDenomBalance := s.App.BankKeeper.GetBalance(s.Ctx, userAddr, bondDenom).Amount

	//
	// TEST: Delegation
	//

	// No delegations
	delegations := s.App.LockupKeeper.GetAllSyntheticLockupsByAddr(s.Ctx, userAddr)
	s.Require().Equal(0, len(delegations))

	balance := s.App.BankKeeper.GetBalance(s.Ctx, userAddr, btcDenom)
	s.Require().Equal(totalBTCAmount, balance.Amount)

	// superfluid stake btcDenom
	btcStakeAmount := osmomath.NewInt(500_000)
	validators, err := s.App.StakingKeeper.GetAllValidators(s.Ctx)
	s.Require().NoError(err)
	validator := validators[0]
	delegateMsg := &types.MsgLockAndSuperfluidDelegate{
		Sender:  userAddr.String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(btcDenom, btcStakeAmount)),
		ValAddr: validator.GetOperator(),
	}
	result, err := s.RunMsg(delegateMsg)
	s.Require().NoError(err)
	// Extract the lock id to use later when undelegating
	attrs := s.ExtractAttributes(s.FindEvent(result.Events, "superfluid_delegate"))
	lockId, err := strconv.ParseUint(attrs["lock_id"], 10, 64)
	s.Require().NoError(err)

	// Check delegations
	delegations = s.App.LockupKeeper.GetAllSyntheticLockupsByAddr(s.Ctx, userAddr)
	s.Require().Equal(1, len(delegations))
	synthLock := delegations[0]
	s.Require().Equal(synthLock.UnderlyingLockId, lockId)

	// check balance
	balance = s.App.BankKeeper.GetBalance(s.Ctx, userAddr, btcDenom)
	s.Require().Equal(btcStakeAmount, balance.Amount)

	underlyingLock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, synthLock.UnderlyingLockId)
	s.Require().NoError(err)
	s.Require().Equal(userAddr.String(), underlyingLock.Owner)
	s.Require().Equal(btcDenom, underlyingLock.Coins[0].Denom)

	queryDelegations := types.SuperfluidDelegationsByDelegatorRequest{DelegatorAddress: userAddr.String()}
	querier := keeper.NewQuerier(*s.App.SuperfluidKeeper)
	res, err := querier.SuperfluidDelegationsByDelegator(s.Ctx, &queryDelegations)
	s.Require().NoError(err)
	s.Require().Len(res.SuperfluidDelegationRecords, 1)
	s.Require().Equal(userAddr.String(), res.SuperfluidDelegationRecords[0].DelegatorAddress)
	s.Require().Equal(validator.GetOperator(), res.SuperfluidDelegationRecords[0].ValidatorAddress)
	s.Require().Equal(btcDenom, res.SuperfluidDelegationRecords[0].DelegationAmount.Denom)
	s.Require().Equal(btcStakeAmount, res.SuperfluidDelegationRecords[0].DelegationAmount.Amount)
	s.Require().Equal(appparams.BaseCoinUnit, res.SuperfluidDelegationRecords[0].EquivalentStakedAmount.Denom)
	riskFactor := s.App.SuperfluidKeeper.CalculateRiskFactor(s.Ctx, btcDenom)
	twapStartTime := s.Ctx.BlockTime().Add(-5 * time.Minute)
	price, err := s.App.TwapKeeper.UnsafeGetMultiPoolArithmeticTwapToNow(s.Ctx, []*poolmanagertypes.SwapAmountInRoute{{PoolId: nextPoolId, TokenOutDenom: bondDenom}}, btcDenom, bondDenom, twapStartTime)
	equivalentAmount := riskFactor.Mul(osmomath.NewDec(btcStakeAmount.Int64())).Mul(price)
	s.Require().NoError(err)
	s.Require().Equal(equivalentAmount, osmomath.NewDec(res.SuperfluidDelegationRecords[0].EquivalentStakedAmount.Amount.Int64()))
	s.Require().Equal(appparams.BaseCoinUnit, res.TotalEquivalentStakedAmount.Denom)
	s.Require().Equal(equivalentAmount, osmomath.NewDec(res.TotalEquivalentStakedAmount.Amount.Int64()))
	s.Require().Equal(appparams.BaseCoinUnit, res.TotalEquivalentNonOsmoStakedAmount.Denom)
	s.Require().Equal(equivalentAmount, osmomath.NewDec(res.TotalEquivalentNonOsmoStakedAmount.Amount.Int64()))

	//
	// TEST: Reward distribution
	//

	// move time beyond the needed time for rewards to be distributed
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(44 * time.Hour))

	// Check that the user has not received any rewards yet
	bondDenomBalance := s.App.BankKeeper.GetBalance(s.Ctx, userAddr, bondDenom)
	s.Require().Equal(originalBondDenomBalance, bondDenomBalance.Amount)

	// There are no rewards assigned to the validator yet
	validatorRewards := new(distrtypes.QueryValidatorOutstandingRewardsResponse)
	err = s.QueryHelper.Invoke(gocontext.Background(),
		"/cosmos.distribution.v1beta1.Query/ValidatorOutstandingRewards",
		&distrtypes.QueryValidatorOutstandingRewardsRequest{
			ValidatorAddress: validator.GetOperator(),
		},
		validatorRewards)
	s.Require().Equal(0, len(validatorRewards.Rewards.Rewards))
	s.Require().NoError(err)

	// Move to block 50 because rewards are only distributed every 50 blocks. Rewards will be available after unstaking
	s.AdvanceToBlockNAndRunEpoch(50)
	fmt.Println("time", s.Ctx.BlockTime())

	// After a block that is not a multiple of 50, the rewards will be assigned to the validator
	validatorRewards = new(distrtypes.QueryValidatorOutstandingRewardsResponse)
	err = s.QueryHelper.Invoke(gocontext.Background(),
		"/cosmos.distribution.v1beta1.Query/ValidatorOutstandingRewards",
		&distrtypes.QueryValidatorOutstandingRewardsRequest{
			ValidatorAddress: validator.GetOperator(),
		},
		validatorRewards)
	s.Require().NoError(err)
	// Validators get uosmo
	s.Require().Equal(1, len(validatorRewards.Rewards.Rewards))

	////////
	// TEST:  Check delegation rewards were distributed
	////////
	bondDenomBalance = s.App.BankKeeper.GetBalance(s.Ctx, userAddr, bondDenom)
	s.Require().True(bondDenomBalance.Amount.GT(originalBondDenomBalance))

	//
	// TEST: Voting. Users should not be allowed to vote when superfluid staking native assets
	//

	// Reset the voting period
	s.App.GovKeeper.ActivateVotingPeriod(s.Ctx, s.proposal)

	// Send vote message
	voteMsg := &govtypes.MsgVote{
		ProposalId: 1,
		Voter:      userAddr.String(),
		Option:     1,
	}
	_, err = s.RunMsg(voteMsg)
	s.Require().NoError(err)

	// Move time beyond voting end time
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(96 * time.Hour))
	s.EndBlock()
	s.BeginNewBlock(true)

	proposal, err := s.App.GovKeeper.Proposals.Get(s.Ctx, 1)
	s.Require().NoError(err)
	s.Require().Equal(govv1.StatusRejected, proposal.Status)
	s.Require().Equal("0", proposal.FinalTallyResult.YesCount)

	////////
	// TEST: Unstake
	////////

	// Check that the user can unstake and the delegation is removed
	undelegateMsg := &types.MsgSuperfluidUndelegateAndUnbondLock{
		Sender: userAddr.String(),
		LockId: underlyingLock.ID,
		Coin:   sdk.NewCoin(btcDenom, btcStakeAmount),
	}
	_, err = s.RunMsg(undelegateMsg)
	s.Require().NoError(err)

	// Check delegations
	res, err = querier.SuperfluidDelegationsByDelegator(s.Ctx, &queryDelegations)
	s.Require().NoError(err)
	s.Require().Len(res.SuperfluidDelegationRecords, 0)

	// Check undelegations
	queryUndelegations := types.SuperfluidUndelegationsByDelegatorRequest{DelegatorAddress: userAddr.String()}
	undelegationResponse, err := querier.SuperfluidUndelegationsByDelegator(s.Ctx, &queryUndelegations)
	s.Require().NoError(err)
	s.Require().Len(undelegationResponse.SuperfluidDelegationRecords, 1)

	// check balance before undelegation time passes
	balance = s.App.BankKeeper.GetBalance(s.Ctx, userAddr, btcDenom)
	s.Require().Equal(btcStakeAmount, balance.Amount)

	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(undelegationResponse.SyntheticLocks[0].Duration - time.Second))
	// move forward to block 60 because we only check matured locks every 30 blocks
	s.AdvanceToBlockNAndRunEpoch(60)

	// No more undelegations
	undelegationResponse, err = querier.SuperfluidUndelegationsByDelegator(s.Ctx, &queryUndelegations)
	s.Require().NoError(err)
	s.Require().Len(undelegationResponse.SuperfluidDelegationRecords, 0)

	// check the btc balance after undelegation time passes. Funds should be restored
	balance = s.App.BankKeeper.GetBalance(s.Ctx, userAddr, btcDenom)
	s.Require().Equal(totalBTCAmount, balance.Amount)
}

func (s *TestSuite) AdvanceToBlockNAndRunEpoch(n int64) {
	for i := s.Ctx.BlockHeight(); i < n; i++ {
		s.EndBlock()
		s.BeginNewBlock(i%n == 0)
	}
	s.EndBlock()
	fmt.Printf("moved to block %d and ran epoch\n", s.Ctx.BlockHeight())
	s.BeginNewBlock(false)
}
