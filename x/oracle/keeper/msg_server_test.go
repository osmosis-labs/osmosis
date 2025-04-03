package keeper_test

import (
	"fmt"
	"github.com/osmosis-labs/osmosis/osmomath"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting/assets"

	"github.com/osmosis-labs/osmosis/v27/x/oracle/keeper"
	"github.com/osmosis-labs/osmosis/v27/x/oracle/types"

	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
)

var (
	stakingAmt         = sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
	randomExchangeRate = osmomath.NewDec(1700)
)

func (s *KeeperTestSuite) setupServer() types.MsgServer {
	params := s.App.OracleKeeper.GetParams(s.Ctx)
	params.VotePeriod = 1
	params.SlashWindow = 100
	params.RewardDistributionWindow = 100
	params.Whitelist = types.DenomList{
		{Name: assets.MicroSDRDenom, TobinTax: types.DefaultTobinTax},
		{Name: assets.StakeDenom, TobinTax: types.DefaultTobinTax},
	}
	s.App.OracleKeeper.SetParams(s.Ctx, params)
	s.App.OracleKeeper.SetTobinTax(s.Ctx, assets.MicroSDRDenom, types.DefaultTobinTax)
	s.App.OracleKeeper.SetTobinTax(s.Ctx, assets.StakeDenom, types.DefaultTobinTax)
	msgServer := keeper.NewMsgServerImpl(*s.App.OracleKeeper)

	stakingMsgSvr := stakingkeeper.NewMsgServerImpl(s.App.StakingKeeper)

	// Validator created
	_, err := stakingMsgSvr.CreateValidator(s.Ctx, s.NewTestMsgCreateValidator(ValAddrs[0], s.valPubKeys[0], stakingAmt))
	s.Require().NoError(err)
	_, err = stakingMsgSvr.CreateValidator(s.Ctx, s.NewTestMsgCreateValidator(ValAddrs[1], s.valPubKeys[1], stakingAmt))
	s.Require().NoError(err)
	_, err = stakingMsgSvr.CreateValidator(s.Ctx, s.NewTestMsgCreateValidator(ValAddrs[2], s.valPubKeys[2], stakingAmt))
	s.Require().NoError(err)
	staking.EndBlocker(s.Ctx, s.App.StakingKeeper)

	return msgServer
}

func (s *KeeperTestSuite) TestMsgServer_FeederDelegation() {
	msgServer := s.setupServer()

	salt := "1"
	hash := types.GetAggregateVoteHash(salt, randomExchangeRate.String()+assets.MicroSDRDenom, ValAddrs[0])

	// Case 1: empty message
	delegateFeedConsentMsg := types.MsgDelegateFeedConsent{}
	_, err := msgServer.DelegateFeedConsent(sdk.WrapSDKContext(s.Ctx), &delegateFeedConsentMsg)
	s.Require().Error(err)

	// Case 2: Normal Prevote - without delegation
	prevoteMsg := types.NewMsgAggregateExchangeRatePrevote(hash, Addrs[0], ValAddrs[0])
	_, err = msgServer.AggregateExchangeRatePrevote(sdk.WrapSDKContext(s.Ctx.WithBlockHeight(0)), prevoteMsg)
	s.Require().NoError(err)

	// Case 2.1: Normal Prevote - with delegation fails
	prevoteMsg = types.NewMsgAggregateExchangeRatePrevote(hash, Addrs[1], ValAddrs[0])
	_, err = msgServer.AggregateExchangeRatePrevote(sdk.WrapSDKContext(s.Ctx), prevoteMsg)
	s.Require().Error(err)

	// Case 2.2: Normal Vote - without delegation
	voteMsg := types.NewMsgAggregateExchangeRateVote(salt, randomExchangeRate.String()+assets.MicroSDRDenom, Addrs[0], ValAddrs[0])
	_, err = msgServer.AggregateExchangeRateVote(sdk.WrapSDKContext(s.Ctx.WithBlockHeight(1)), voteMsg)
	s.Require().NoError(err)

	// Case 2.3: Normal Vote - with delegation fails
	voteMsg = types.NewMsgAggregateExchangeRateVote(salt, randomExchangeRate.String()+assets.MicroSDRDenom, Addrs[1], ValAddrs[0])
	_, err = msgServer.AggregateExchangeRateVote(sdk.WrapSDKContext(s.Ctx.WithBlockHeight(1)), voteMsg)
	s.Require().Error(err)

	// Case 3: Normal MsgDelegateFeedConsent succeeds
	msg := types.NewMsgDelegateFeedConsent(ValAddrs[0], Addrs[1])
	_, err = msgServer.DelegateFeedConsent(sdk.WrapSDKContext(s.Ctx), msg)
	s.Require().NoError(err)

	// Case 4.1: Normal Prevote - without delegation fails
	prevoteMsg = types.NewMsgAggregateExchangeRatePrevote(hash, Addrs[2], ValAddrs[0])
	_, err = msgServer.AggregateExchangeRatePrevote(sdk.WrapSDKContext(s.Ctx), prevoteMsg)
	s.Require().Error(err)

	// Case 4.2: Normal Prevote - with delegation succeeds
	prevoteMsg = types.NewMsgAggregateExchangeRatePrevote(hash, Addrs[1], ValAddrs[0])
	_, err = msgServer.AggregateExchangeRatePrevote(sdk.WrapSDKContext(s.Ctx), prevoteMsg)
	s.Require().NoError(err)

	// Case 4.3: Normal Vote - without delegation fails
	voteMsg = types.NewMsgAggregateExchangeRateVote(salt, randomExchangeRate.String()+assets.MicroSDRDenom, Addrs[2], ValAddrs[0])
	_, err = msgServer.AggregateExchangeRateVote(sdk.WrapSDKContext(s.Ctx.WithBlockHeight(2)), voteMsg)
	s.Require().Error(err)

	// Case 4.4: Normal Vote - with delegation succeeds
	voteMsg = types.NewMsgAggregateExchangeRateVote(salt, randomExchangeRate.String()+assets.MicroSDRDenom, Addrs[1], ValAddrs[0])
	_, err = msgServer.AggregateExchangeRateVote(sdk.WrapSDKContext(s.Ctx.WithBlockHeight(2)), voteMsg)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestMsgServer_AggregatePrevoteVote() {
	msgServer := s.setupServer()

	salt := "1"
	exchangeRatesStr := fmt.Sprintf("1000.23%s,0.29%s", assets.MicroUSDDenom, assets.MicroSDRDenom)
	otherExchangeRateStr := fmt.Sprintf("1000.12%s,0.29%s", assets.MicroUSDDenom, assets.MicroUSDDenom)
	unintendedExchageRateStr := fmt.Sprintf("1000.23%s,0.29%s", assets.MicroUSDDenom, assets.MicroCNYDenom)
	invalidExchangeRateStr := fmt.Sprintf("1000.23%s,0.29", assets.MicroUSDDenom)

	hash := types.GetAggregateVoteHash(salt, exchangeRatesStr, ValAddrs[0])

	aggregateExchangeRatePrevoteMsg := types.NewMsgAggregateExchangeRatePrevote(hash, Addrs[0], ValAddrs[0])
	_, err := msgServer.AggregateExchangeRatePrevote(sdk.WrapSDKContext(s.Ctx), aggregateExchangeRatePrevoteMsg)
	s.Require().NoError(err)

	// Unauthorized feeder
	aggregateExchangeRatePrevoteMsg = types.NewMsgAggregateExchangeRatePrevote(hash, Addrs[1], ValAddrs[0])
	_, err = msgServer.AggregateExchangeRatePrevote(sdk.WrapSDKContext(s.Ctx), aggregateExchangeRatePrevoteMsg)
	s.Require().Error(err)

	// Invalid addr
	aggregateExchangeRatePrevoteMsg = types.NewMsgAggregateExchangeRatePrevote(hash, sdk.AccAddress{}, ValAddrs[0])
	_, err = msgServer.AggregateExchangeRatePrevote(sdk.WrapSDKContext(s.Ctx), aggregateExchangeRatePrevoteMsg)
	s.Require().Error(err)

	// Invalid validator addr
	aggregateExchangeRatePrevoteMsg = types.NewMsgAggregateExchangeRatePrevote(hash, Addrs[0], sdk.ValAddress{})
	_, err = msgServer.AggregateExchangeRatePrevote(sdk.WrapSDKContext(s.Ctx), aggregateExchangeRatePrevoteMsg)
	s.Require().Error(err)

	// Invalid reveal period
	aggregateExchangeRateVoteMsg := types.NewMsgAggregateExchangeRateVote(salt, exchangeRatesStr, Addrs[0], ValAddrs[0])
	_, err = msgServer.AggregateExchangeRateVote(sdk.WrapSDKContext(s.Ctx), aggregateExchangeRateVoteMsg)
	s.Require().Error(err)

	// Invalid reveal period
	s.Ctx = s.Ctx.WithBlockHeight(1)
	aggregateExchangeRateVoteMsg = types.NewMsgAggregateExchangeRateVote(salt, exchangeRatesStr, Addrs[0], ValAddrs[0])
	_, err = msgServer.AggregateExchangeRateVote(sdk.WrapSDKContext(s.Ctx), aggregateExchangeRateVoteMsg)
	s.Require().Error(err)

	// Other exchange rate with valid real period
	s.Ctx = s.Ctx.WithBlockHeight(1)
	aggregateExchangeRateVoteMsg = types.NewMsgAggregateExchangeRateVote(salt, otherExchangeRateStr, Addrs[0], ValAddrs[0])
	_, err = msgServer.AggregateExchangeRateVote(sdk.WrapSDKContext(s.Ctx), aggregateExchangeRateVoteMsg)
	s.Require().Error(err)

	// Invalid exchange rate with valid real period
	s.Ctx = s.Ctx.WithBlockHeight(1)
	aggregateExchangeRateVoteMsg = types.NewMsgAggregateExchangeRateVote(salt, invalidExchangeRateStr, Addrs[0], ValAddrs[0])
	_, err = msgServer.AggregateExchangeRateVote(sdk.WrapSDKContext(s.Ctx), aggregateExchangeRateVoteMsg)
	s.Require().Error(err)

	// Unauthorized feeder
	aggregateExchangeRateVoteMsg = types.NewMsgAggregateExchangeRateVote(salt, invalidExchangeRateStr, Addrs[1], ValAddrs[0])
	_, err = msgServer.AggregateExchangeRateVote(sdk.WrapSDKContext(s.Ctx), aggregateExchangeRateVoteMsg)
	s.Require().Error(err)

	// Unintended denom vote
	aggregateExchangeRateVoteMsg = types.NewMsgAggregateExchangeRateVote(salt, unintendedExchageRateStr, Addrs[0], ValAddrs[0])
	_, err = msgServer.AggregateExchangeRateVote(sdk.WrapSDKContext(s.Ctx), aggregateExchangeRateVoteMsg)
	s.Require().Error(err)

	// Valid exchange rate reveal submission
	s.Ctx = s.Ctx.WithBlockHeight(2)
	aggregateExchangeRateVoteMsg = types.NewMsgAggregateExchangeRateVote(salt, exchangeRatesStr, Addrs[0], ValAddrs[0])
	_, err = msgServer.AggregateExchangeRateVote(sdk.WrapSDKContext(s.Ctx.WithBlockHeight(2)), aggregateExchangeRateVoteMsg)
	s.Require().NoError(err)
}
