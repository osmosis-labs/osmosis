package keeper_test

import (
	gocontext "context"
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/osmosis-labs/osmosis/osmomath"
	appparams "github.com/osmosis-labs/osmosis/v25/app/params"
	balancertypes "github.com/osmosis-labs/osmosis/v25/x/gamm/pool-models/balancer"
	minttypes "github.com/osmosis-labs/osmosis/v25/x/mint/types"
	"github.com/osmosis-labs/osmosis/v25/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/v25/x/superfluid/types"
	"github.com/stretchr/testify/suite"
	"strconv"
	"testing"
	"time"
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
	stakingParams := s.App.StakingKeeper.GetParams(s.Ctx)
	stakingParams.BondDenom = appparams.BaseCoinUnit
	err := s.App.StakingKeeper.SetParams(s.Ctx, stakingParams)
	s.Require().NoError(err)

	// set incentives min value in osmo
	incentivesParams := s.App.IncentivesKeeper.GetParams(s.Ctx)
	incentivesParams.MinValueForDistribution.Denom = appparams.BaseCoinUnit
	s.App.IncentivesKeeper.SetParams(s.Ctx, incentivesParams)

	// make pool creation fees be paid in the bond denom. Also make them low.
	poolmanagerParams := s.App.PoolManagerKeeper.GetParams(s.Ctx)
	poolmanagerParams.PoolCreationFee = sdk.NewCoins(sdk.NewInt64Coin(s.App.StakingKeeper.BondDenom(s.Ctx), 1))
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
	btcDenom := "btc" // Asset to superfluid stake
	bondDenom := s.App.StakingKeeper.BondDenom(s.Ctx)

	// accounts
	// pool creator
	lpKey := ed25519.GenPrivKey().PubKey()
	lpAddr := sdk.AccAddress(lpKey.Address())

	osmoPoolAmount := sdk.NewInt(1_000_000_000_000)
	btcPoolAmount := sdk.NewInt(10_000_000_000)
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

	_, err := s.RunMsg(createPoolMsg)
	s.Require().NoError(err)
	gammToken := fmt.Sprintf("gamm/pool/%d", nextPoolId)

	totalGammTokens := s.App.BankKeeper.GetBalance(s.Ctx, lpAddr, gammToken)

	// Add btcDenom as an allowed superfluid asset
	err = s.App.SuperfluidKeeper.AddNewSuperfluidAsset(s.Ctx, types.SuperfluidAsset{Denom: gammToken, AssetType: types.SuperfluidAssetTypeLPShare})
	s.Require().NoError(err)

	// Mint assets to the lockup module. This will ensure there are assets to distribute.
	err = s.App.BankKeeper.MintCoins(s.Ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin(bondDenom, sdk.NewInt(1_000_000_000))))
	s.Require().NoError(err)
	err = s.App.BankKeeper.SendCoinsFromModuleToModule(s.Ctx, minttypes.ModuleName, authtypes.FeeCollectorName, sdk.NewCoins(sdk.NewCoin(bondDenom, sdk.NewInt(1_000_000_000))))
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
	validator := s.App.StakingKeeper.GetAllValidators(s.Ctx)[0]
	gammDelegationAmount := sdk.NewInt(1000000000000000000)
	delegateMsg := &types.MsgLockAndSuperfluidDelegate{
		Sender:  lpAddr.String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(gammToken, gammDelegationAmount)),
		ValAddr: validator.GetOperator().String(),
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

	////////
	// TEST: Reward distribution
	////////

	// Check that the user has not received any rewards yet
	bondDenomBalance := s.App.BankKeeper.GetBalance(s.Ctx, lpAddr, bondDenom)
	s.Require().Equal(originalBondDenomBalance, bondDenomBalance.Amount)

	// There are no rewards assigned to the validator yet
	validatorRewards := new(distrtypes.QueryValidatorOutstandingRewardsResponse)
	err = s.QueryHelper.Invoke(gocontext.Background(),
		"/cosmos.distribution.v1beta1.Query/ValidatorOutstandingRewards",
		&distrtypes.QueryValidatorOutstandingRewardsRequest{
			ValidatorAddress: validator.GetOperator().String(),
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
			ValidatorAddress: validator.GetOperator().String(),
		},
		validatorRewards)
	s.Require().NoError(err)
	s.Require().Equal(1, len(validatorRewards.Rewards.Rewards))
	// After unstaking we will check that those rewards were properly distributed to the delegator

	////////
	// TEST: Voting. User can vote
	////////

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

	proposal, found := s.App.GovKeeper.GetProposal(s.Ctx, 1)
	s.Require().True(found)
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
	querier := keeper.NewQuerier(*s.App.SuperfluidKeeper)
	queryDelegations := types.SuperfluidDelegationsByDelegatorRequest{DelegatorAddress: lpAddr.String()}
	res, err := querier.SuperfluidDelegationsByDelegator(s.Ctx, &queryDelegations)
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

	////////
	// TEST:  Check delegation rewards were distributed
	////////
	bondDenomBalance = s.App.BankKeeper.GetBalance(s.Ctx, lpAddr, bondDenom)
	s.Require().True(bondDenomBalance.Amount.GT(originalBondDenomBalance))
}

func (s *TestSuite) TestNativeSuperfluid() {
	//
	// TEST: Setup
	//

	s.SetupTest()

	// denoms
	btcDenom := "btc" // Asset to superfluid stake
	bondDenom := s.App.StakingKeeper.BondDenom(s.Ctx)

	// accounts
	// pool creator
	lpKey := ed25519.GenPrivKey().PubKey()
	lpAddr := sdk.AccAddress(lpKey.Address())
	userKey := ed25519.GenPrivKey().PubKey()
	userAddr := sdk.AccAddress(userKey.Address())

	osmoPoolAmount := sdk.NewInt(1_000_000_000_000)
	btcPoolAmount := sdk.NewInt(10_000_000_000)
	// default bond denom

	// mint necessary tokens
	s.mintToAccount(btcPoolAmount, btcDenom, lpAddr)
	s.mintToAccount(osmoPoolAmount.Mul(osmomath.NewInt(2)), bondDenom, lpAddr)
	s.mintToAccount(sdk.NewInt(100_000_000), bondDenom, userAddr)
	s.mintToAccount(sdk.NewInt(1_000_000), btcDenom, userAddr)

	nextPoolId := s.App.PoolManagerKeeper.GetNextPoolId(s.Ctx) // the pool id we'll create

	// create an bondDenom/btcDenom pool. This is only used so that the native asset can have a price.
	createPoolMsg := createPoolMsgGen(
		lpAddr,
		sdk.NewCoins(sdk.NewCoin(btcDenom, btcPoolAmount), sdk.NewCoin(bondDenom, osmoPoolAmount)),
	)

	_, err := s.RunMsg(createPoolMsg)
	s.Require().NoError(err)

	// Creating a native type without a pool should fail
	err = s.App.SuperfluidKeeper.AddNewSuperfluidAsset(s.Ctx, types.SuperfluidAsset{Denom: btcDenom, AssetType: types.SuperfluidAssetTypeNative})
	s.Require().Error(err)

	// Add btcDenom as an allowed superfluid asset
	err = s.App.SuperfluidKeeper.AddNewSuperfluidAsset(s.Ctx, types.SuperfluidAsset{Denom: btcDenom, AssetType: types.SuperfluidAssetTypeNative, PricePoolId: nextPoolId})
	s.Require().NoError(err)

	id, err := s.App.PoolIncentivesKeeper.GetInternalGaugeIDForPool(s.Ctx, nextPoolId)
	s.Require().NoError(err)
	s.mintToAccount(sdk.NewInt(1_000_000_000), bondDenom, lpAddr)
	err = s.App.IncentivesKeeper.AddToGaugeRewards(s.Ctx, lpAddr, sdk.NewCoins(sdk.NewCoin(bondDenom, sdk.NewInt(1_000_000_000))), id)
	s.Require().NoError(err)
	gauges := s.App.PoolIncentivesKeeper.GetAllGauges(s.Ctx)
	fmt.Println("gauges", gauges)

	//
	// TEST: Delegation
	//

	// No delegations
	delegations := s.App.LockupKeeper.GetAllSyntheticLockupsByAddr(s.Ctx, userAddr)
	s.Require().Equal(0, len(delegations))

	balance := s.App.BankKeeper.GetBalance(s.Ctx, userAddr, btcDenom)
	s.Require().Equal(sdk.NewInt(1_000_000), balance.Amount)

	// superfluid stake btcDenom
	btcStakeAmount := sdk.NewInt(500_000)
	validator := s.App.StakingKeeper.GetAllValidators(s.Ctx)[0]
	delegateMsg := &types.MsgLockAndSuperfluidDelegate{
		Sender:  userAddr.String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(btcDenom, btcStakeAmount)),
		ValAddr: validator.GetOperator().String(),
	}
	result, err := s.RunMsg(delegateMsg)
	s.Require().NoError(err)
	// Extract the lock id to use later when undelegating
	attrs := s.ExtractAttributes(s.FindEvent(result.GetEvents(), "superfluid_delegate"))
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
	s.Require().Equal(validator.GetOperator().String(), res.SuperfluidDelegationRecords[0].ValidatorAddress)
	s.Require().Equal(btcDenom, res.SuperfluidDelegationRecords[0].DelegationAmount.Denom)
	s.Require().Equal(btcStakeAmount, res.SuperfluidDelegationRecords[0].DelegationAmount.Amount)
	s.Require().Equal("uosmo", res.SuperfluidDelegationRecords[0].EquivalentStakedAmount.Denom)
	s.Require().Equal(sdk.NewInt(0), res.SuperfluidDelegationRecords[0].EquivalentStakedAmount.Amount)
	s.Require().Equal("uosmo", res.TotalEquivalentStakedAmount.Denom)
	s.Require().Equal(sdk.NewInt(0), res.TotalEquivalentStakedAmount.Amount)

	//
	// TEST: Reward distribution
	//

	// Run epoch
	// move forward to block 30 because we only check matured locks every 30 blocks
	s.AdvanceToBlockNAndRunEpoch(50)

	// TODO: How do I check distribution happened properly?

	//
	// TEST: Voting. Users should not be allowed to vote when superfluid staking native assets
	//

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

	proposal, found := s.App.GovKeeper.GetProposal(s.Ctx, 1)
	s.Require().True(found)
	s.Require().Equal(govv1.StatusRejected, proposal.Status)
	s.Require().Equal("0", proposal.FinalTallyResult.YesCount)

	//
	// TEST: Unstake
	//

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

	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(undelegationResponse.SyntheticLocks[0].Duration + time.Second))
	// move forward to block 60 because we only check matured locks every 30 blocks
	s.AdvanceToBlockNAndRunEpoch(60)

	// check balance after undelegation time passes
	balance = s.App.BankKeeper.GetBalance(s.Ctx, userAddr, btcDenom)
	s.Require().Equal(sdk.NewInt(1_000_000), balance.Amount)

	balances := s.App.BankKeeper.GetAllBalances(s.Ctx, userAddr)
	fmt.Println(balances)
}

func (s *TestSuite) TestCLSuperfluid() {
	//
	// Setup
	//

	s.SetupTest()

	// denoms
	btcDenom := "eth" // Asset to superfluid stake. Using eth because it's an authorized denom
	bondDenom := s.App.StakingKeeper.BondDenom(s.Ctx)

	// accounts
	// pool creator
	lpKey := ed25519.GenPrivKey().PubKey()
	lpAddr := sdk.AccAddress(lpKey.Address())

	osmoPoolAmount := sdk.NewInt(1_000_000_000_000)
	btcPoolAmount := sdk.NewInt(10_000_000_000)
	// default bond denom

	// mint necessary tokens
	s.mintToAccount(btcPoolAmount, btcDenom, lpAddr)
	s.mintToAccount(osmoPoolAmount.Mul(osmomath.NewInt(2)), bondDenom, lpAddr)

	nextPoolId := s.App.PoolManagerKeeper.GetNextPoolId(s.Ctx) // the pool id we'll create

	// create an bondDenom/btcDenom pool and add full range liquidity
	clPool := s.PrepareCustomConcentratedPool(lpAddr, bondDenom, btcDenom, uint64(1), osmomath.ZeroDec())
	s.CreateFullRangePosition(clPool, sdk.NewCoins(sdk.NewCoin(bondDenom, osmoPoolAmount), sdk.NewCoin(btcDenom, btcPoolAmount)))

	//
	// TEST: Delegate gamm tokens
	//

	// No delegations
	delegations := s.App.LockupKeeper.GetAllSyntheticLockupsByAddr(s.Ctx, lpAddr)
	s.Require().Equal(0, len(delegations))

	// Add the pool as an allowed superfluid asset
	err := s.App.SuperfluidKeeper.AddNewSuperfluidAsset(s.Ctx, types.SuperfluidAsset{Denom: "cl/pool/1", AssetType: types.SuperfluidAssetTypeConcentratedShare})
	s.Require().NoError(err)

	// superfluid stake cl position
	osmoDelegateAmount := sdk.NewInt(1_000_000)
	btcDelegateAmount := sdk.NewInt(100_000)
	validator := s.App.StakingKeeper.GetAllValidators(s.Ctx)[0]
	clDelegateMsg := &types.MsgCreateFullRangePositionAndSuperfluidDelegate{
		Sender:  lpAddr.String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(bondDenom, osmoDelegateAmount), sdk.NewCoin(btcDenom, btcDelegateAmount)),
		ValAddr: validator.GetOperator().String(),
		PoolId:  nextPoolId,
	}
	_, err = s.RunMsg(clDelegateMsg)
	s.Require().NoError(err)

	// Check delegations
	delegations = s.App.LockupKeeper.GetAllSyntheticLockupsByAddr(s.Ctx, lpAddr)
	s.Require().Equal(1, len(delegations))
	synthLock := delegations[0]

	// Get underlying lock
	underlyingLock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, synthLock.UnderlyingLockId)
	s.Require().NoError(err)
	s.Require().Equal(lpAddr.String(), underlyingLock.Owner)

	//
	// TEST: Reward distribution
	//

	// ensure there are some fees to distribute
	rewards := sdk.NewCoin(bondDenom, sdk.NewInt(5_000_000))
	err = s.App.BankKeeper.MintCoins(s.Ctx, minttypes.ModuleName, sdk.NewCoins(rewards))
	s.Require().NoError(err)
	err = s.App.MintKeeper.DistributeMintedCoin(s.Ctx, rewards)
	s.Require().NoError(err)

	// TODO: Still not sure how to check rewards.
	balances := s.App.BankKeeper.GetAllBalances(s.Ctx, lpAddr)
	fmt.Println("balances", balances)
	s.AdvanceToBlockNAndRunEpoch(30)
	s.AdvanceToBlockNAndRunEpoch(50)
	balances = s.App.BankKeeper.GetAllBalances(s.Ctx, lpAddr)
	fmt.Println("balances", balances)

	//
	// TEST: Voting. User can vote
	//

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

	coins, err := s.App.DistrKeeper.WithdrawDelegationRewards(s.Ctx, lpAddr, validator.GetOperator())
	fmt.Println("coins", coins)

	proposal, found := s.App.GovKeeper.GetProposal(s.Ctx, 1)
	s.Require().True(found)
	s.Require().Equal(govv1.StatusRejected, proposal.Status)
	s.Require().Equal("500000", proposal.FinalTallyResult.YesCount)

	//
	// TEST: Unstake
	//

	balances = s.App.BankKeeper.GetAllBalances(s.Ctx, lpAddr)
	fmt.Println("balances", balances)

	// Check that the user can unstake and the delegation is removed
	undelegateMsg := &types.MsgSuperfluidUndelegateAndUnbondLock{
		Sender: lpAddr.String(),
		LockId: underlyingLock.ID,
		Coin:   sdk.NewCoin("cl/pool/1", underlyingLock.Coins[0].Amount),
	}
	_, err = s.RunMsg(undelegateMsg)
	s.Require().NoError(err)

	// Check delegations
	querier := keeper.NewQuerier(*s.App.SuperfluidKeeper)
	queryDelegations := types.SuperfluidDelegationsByDelegatorRequest{DelegatorAddress: lpAddr.String()}
	res, err := querier.SuperfluidDelegationsByDelegator(s.Ctx, &queryDelegations)
	s.Require().NoError(err)
	s.Require().Len(res.SuperfluidDelegationRecords, 0)

	// Check undelegations
	queryUndelegations := types.SuperfluidUndelegationsByDelegatorRequest{DelegatorAddress: lpAddr.String()}
	undelegationResponse, err := querier.SuperfluidUndelegationsByDelegator(s.Ctx, &queryUndelegations)
	s.Require().NoError(err)
	s.Require().Len(undelegationResponse.SuperfluidDelegationRecords, 1)

	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(undelegationResponse.SyntheticLocks[0].Duration + time.Second))
	// move forward to block 60 because we only check matured locks every 30 blocks
	s.AdvanceToBlockNAndRunEpoch(60)

	// No more undelegations
	undelegationResponse, err = querier.SuperfluidUndelegationsByDelegator(s.Ctx, &queryUndelegations)
	s.Require().NoError(err)
	s.Require().Len(undelegationResponse.SuperfluidDelegationRecords, 0)

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
