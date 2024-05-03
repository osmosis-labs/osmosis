package keeper_test

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/osmosis-labs/osmosis/osmomath"
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
	s.SetupTest()
	//addr1 := sdk.AccAddress(pk1.Address())

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
	fmt.Println("remainingGammTokens", remainingGammTokens)
	fmt.Println("totalGammTokens", totalGammTokens)
	s.Require().Equal(totalGammTokens.Amount.Sub(gammDelegationAmount), remainingGammTokens.Amount)

	// Run epoch
	s.EndBlock()
	s.BeginNewBlock(true)

	// TODO: How do I check distribution happened properly?

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

	//
	// TEST: Unstake
	//

	// Check that the user can unstake and the delegation is removed
	undelegateMsg := &types.MsgSuperfluidUndelegate{
		Sender: lpAddr.String(),
		LockId: underlyingLock.ID,
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

	// check balance before undelegation time passes
	balance := s.App.BankKeeper.GetBalance(s.Ctx, lpAddr, gammToken)
	s.Require().Equal(remainingGammTokens.Amount, balance.Amount)
	fmt.Println("balance", balance)

	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(undelegationResponse.SyntheticLocks[0].Duration + time.Second))
	// move forward to block 29 because we only check matured locks every 30 blocks
	for i := s.Ctx.BlockHeight(); i < 30; i++ {
		s.EndBlock()
		s.BeginNewBlock(i%30 == 0)
	}
	s.EndBlock()
	s.BeginNewBlock(false)

	// No more undelegations
	undelegationResponse, err = querier.SuperfluidUndelegationsByDelegator(s.Ctx, &queryUndelegations)
	s.Require().NoError(err)
	s.Require().Len(undelegationResponse.SuperfluidDelegationRecords, 0)

	// check balance after undelegation time passes
	balance = s.App.BankKeeper.GetBalance(s.Ctx, lpAddr, gammToken)
	fmt.Println("balance", balance)
	s.Require().Equal(totalGammTokens.Amount, balance.Amount)

}

func (s *TestSuite) TestNativeSuperfluid() {
	s.SetupTest()
	//addr1 := sdk.AccAddress(pk1.Address())

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

	// No delegations
	delegations := s.App.LockupKeeper.GetAllSyntheticLockupsByAddr(s.Ctx, userAddr)
	s.Require().Equal(0, len(delegations))

	balance := s.App.BankKeeper.GetBalance(s.Ctx, userAddr, btcDenom)
	s.Require().Equal(sdk.NewInt(1_000_000), balance.Amount)

	// superfluid stake btcDenom
	validator := s.App.StakingKeeper.GetAllValidators(s.Ctx)[0]
	delegateMsg := &types.MsgLockAndSuperfluidDelegate{
		Sender:  userAddr.String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(btcDenom, sdk.NewInt(500_000))),
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

	// check balance
	balance = s.App.BankKeeper.GetBalance(s.Ctx, userAddr, btcDenom)
	s.Require().Equal(sdk.NewInt(500_000), balance.Amount)

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
	s.Require().Equal(sdk.NewInt(500_000), res.SuperfluidDelegationRecords[0].DelegationAmount.Amount)
	s.Require().Equal("uosmo", res.SuperfluidDelegationRecords[0].EquivalentStakedAmount.Denom)
	s.Require().Equal(sdk.NewInt(0), res.SuperfluidDelegationRecords[0].EquivalentStakedAmount.Amount)
	s.Require().Equal("uosmo", res.TotalEquivalentStakedAmount.Denom)
	s.Require().Equal(sdk.NewInt(0), res.TotalEquivalentStakedAmount.Amount)

	// Get underlying lock

	// Run epoch
	s.EndBlock()
	s.BeginNewBlock(true)

	// TODO: How do I check distribution happened properly?
	// I think what's happening here is that they only happen on multiple of 50 blocks

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
	undelegateMsg := &types.MsgSuperfluidUndelegate{
		Sender: userAddr.String(),
		LockId: lockId,
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
	s.Require().Equal(sdk.NewInt(500_000), balance.Amount)

	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(undelegationResponse.SyntheticLocks[0].Duration + time.Second))
	s.EndBlock()
	s.BeginNewBlock(true)

	// check balance after undelegation time passes
	balance = s.App.BankKeeper.GetBalance(s.Ctx, userAddr, btcDenom)
	s.Require().Equal(sdk.NewInt(1_000_000), balance.Amount)
}
