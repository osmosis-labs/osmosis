package keeper_test

import (
	"github.com/osmosis-labs/osmosis/osmomath"
	"sort"

	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting/assets"
	"github.com/osmosis-labs/osmosis/v27/x/oracle/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

func (s *KeeperTestSuite) TestOrganizeAggregate() {
	power := int64(100)
	amt := sdk.TokensFromConsensusPower(power, sdk.DefaultPowerReduction)
	stakingMsgSvr := stakingkeeper.NewMsgServerImpl(s.App.StakingKeeper)
	ctx := s.Ctx

	// Validator created
	_, err := stakingMsgSvr.CreateValidator(ctx, s.NewTestMsgCreateValidator(ValAddrs[0], s.valPubKeys[0], amt))
	s.Require().NoError(err)
	_, err = stakingMsgSvr.CreateValidator(ctx, s.NewTestMsgCreateValidator(ValAddrs[1], s.valPubKeys[1], amt))
	s.Require().NoError(err)
	_, err = stakingMsgSvr.CreateValidator(ctx, s.NewTestMsgCreateValidator(ValAddrs[2], s.valPubKeys[2], amt))
	s.Require().NoError(err)
	staking.EndBlocker(ctx, s.App.StakingKeeper)

	sdrBallot := types.ExchangeRateBallot{
		types.NewVoteForTally(osmomath.NewDec(17), assets.MicroSDRDenom, ValAddrs[0], power),
		types.NewVoteForTally(osmomath.NewDec(10), assets.MicroSDRDenom, ValAddrs[1], power),
		types.NewVoteForTally(osmomath.NewDec(6), assets.MicroSDRDenom, ValAddrs[2], power),
	}
	krwBallot := types.ExchangeRateBallot{
		types.NewVoteForTally(osmomath.NewDec(1000), assets.MicroKRWDenom, ValAddrs[0], power),
		types.NewVoteForTally(osmomath.NewDec(1300), assets.MicroKRWDenom, ValAddrs[1], power),
		types.NewVoteForTally(osmomath.NewDec(2000), assets.MicroKRWDenom, ValAddrs[2], power),
	}

	for i := range sdrBallot {
		s.App.OracleKeeper.SetAggregateExchangeRateVote(s.Ctx, ValAddrs[i],
			types.NewAggregateExchangeRateVote(types.ExchangeRateTuples{
				{Denom: sdrBallot[i].Denom, ExchangeRate: sdrBallot[i].ExchangeRate},
				{Denom: krwBallot[i].Denom, ExchangeRate: krwBallot[i].ExchangeRate},
			}, ValAddrs[i]))
	}

	// organize votes by denom
	ballotMap := s.App.OracleKeeper.OrganizeBallotByDenom(s.Ctx, map[string]types.Claim{
		ValAddrs[0].String(): {
			Power:     power,
			WinCount:  0,
			Recipient: ValAddrs[0],
		},
		ValAddrs[1].String(): {
			Power:     power,
			WinCount:  0,
			Recipient: ValAddrs[1],
		},
		ValAddrs[2].String(): {
			Power:     power,
			WinCount:  0,
			Recipient: ValAddrs[2],
		},
	})

	// sort each ballot for comparison
	sort.Sort(sdrBallot)
	sort.Sort(krwBallot)
	sort.Sort(ballotMap[assets.MicroSDRDenom])
	sort.Sort(ballotMap[assets.MicroKRWDenom])

	s.Require().Equal(sdrBallot, ballotMap[assets.MicroSDRDenom])
	s.Require().Equal(krwBallot, ballotMap[assets.MicroKRWDenom])
}

func (s *KeeperTestSuite) TestClearBallots() {
	power := int64(100)
	amt := sdk.TokensFromConsensusPower(power, sdk.DefaultPowerReduction)
	stakingMsgSvr := stakingkeeper.NewMsgServerImpl(s.App.StakingKeeper)
	ctx := s.Ctx

	// Validator created
	_, err := stakingMsgSvr.CreateValidator(ctx, s.NewTestMsgCreateValidator(ValAddrs[0], s.valPubKeys[0], amt))
	s.Require().NoError(err)
	_, err = stakingMsgSvr.CreateValidator(ctx, s.NewTestMsgCreateValidator(ValAddrs[1], s.valPubKeys[1], amt))
	s.Require().NoError(err)
	_, err = stakingMsgSvr.CreateValidator(ctx, s.NewTestMsgCreateValidator(ValAddrs[2], s.valPubKeys[2], amt))
	s.Require().NoError(err)
	staking.EndBlocker(ctx, s.App.StakingKeeper)

	sdrBallot := types.ExchangeRateBallot{
		types.NewVoteForTally(osmomath.NewDec(17), assets.MicroSDRDenom, ValAddrs[0], power),
		types.NewVoteForTally(osmomath.NewDec(10), assets.MicroSDRDenom, ValAddrs[1], power),
		types.NewVoteForTally(osmomath.NewDec(6), assets.MicroSDRDenom, ValAddrs[2], power),
	}
	krwBallot := types.ExchangeRateBallot{
		types.NewVoteForTally(osmomath.NewDec(1000), assets.MicroKRWDenom, ValAddrs[0], power),
		types.NewVoteForTally(osmomath.NewDec(1300), assets.MicroKRWDenom, ValAddrs[1], power),
		types.NewVoteForTally(osmomath.NewDec(2000), assets.MicroKRWDenom, ValAddrs[2], power),
	}

	for i := range sdrBallot {
		s.App.OracleKeeper.SetAggregateExchangeRatePrevote(s.Ctx, ValAddrs[i], types.AggregateExchangeRatePrevote{
			Hash:        "",
			Voter:       ValAddrs[i].String(),
			SubmitBlock: uint64(s.Ctx.BlockHeight()),
		})

		s.App.OracleKeeper.SetAggregateExchangeRateVote(s.Ctx, ValAddrs[i],
			types.NewAggregateExchangeRateVote(types.ExchangeRateTuples{
				{Denom: sdrBallot[i].Denom, ExchangeRate: sdrBallot[i].ExchangeRate},
				{Denom: krwBallot[i].Denom, ExchangeRate: krwBallot[i].ExchangeRate},
			}, ValAddrs[i]))
	}

	s.App.OracleKeeper.ClearBallots(s.Ctx, 5)

	prevoteCounter := 0
	voteCounter := 0
	s.App.OracleKeeper.IterateAggregateExchangeRatePrevotes(s.Ctx, func(_ sdk.ValAddress, _ types.AggregateExchangeRatePrevote) bool {
		prevoteCounter++
		return false
	})
	s.App.OracleKeeper.IterateAggregateExchangeRateVotes(s.Ctx, func(_ sdk.ValAddress, _ types.AggregateExchangeRateVote) bool {
		voteCounter++
		return false
	})

	s.Require().Equal(prevoteCounter, 3)
	s.Require().Equal(voteCounter, 0)

	s.App.OracleKeeper.ClearBallots(s.Ctx.WithBlockHeight(s.Ctx.BlockHeight()+6), 5)

	prevoteCounter = 0
	s.App.OracleKeeper.IterateAggregateExchangeRatePrevotes(s.Ctx, func(_ sdk.ValAddress, _ types.AggregateExchangeRatePrevote) bool {
		prevoteCounter++
		return false
	})
	s.Require().Equal(prevoteCounter, 0)
}

func (s *KeeperTestSuite) TestApplyWhitelist() {
	// no update
	s.App.OracleKeeper.ApplyWhitelist(s.Ctx, types.DenomList{
		types.Denom{
			Name:     "uusd",
			TobinTax: osmomath.OneDec(),
		},
		types.Denom{
			Name:     "ukrw",
			TobinTax: osmomath.OneDec(),
		},
	}, map[string]osmomath.Dec{
		"uusd": osmomath.ZeroDec(),
		"ukrw": osmomath.ZeroDec(),
	})

	price, err := s.App.OracleKeeper.GetTobinTax(s.Ctx, "uusd")
	s.Require().NoError(err)
	s.Require().Equal(price, osmomath.OneDec())

	price, err = s.App.OracleKeeper.GetTobinTax(s.Ctx, "ukrw")
	s.Require().NoError(err)
	s.Require().Equal(price, osmomath.OneDec())

	metadata, ok := s.App.BankKeeper.GetDenomMetaData(s.Ctx, "uusd")
	s.Require().True(ok)
	s.Require().Equal(metadata.Base, "uusd")
	s.Require().Equal(metadata.Display, "usd")
	s.Require().Equal(len(metadata.DenomUnits), 3)
	s.Require().Equal(metadata.Description, "The native stable token of the Symphony.")

	metadata, ok = s.App.BankKeeper.GetDenomMetaData(s.Ctx, "ukrw")
	s.Require().True(ok)
	s.Require().Equal(metadata.Base, "ukrw")
	s.Require().Equal(metadata.Display, "krw")
	s.Require().Equal(len(metadata.DenomUnits), 3)
	s.Require().Equal(metadata.Description, "The native stable token of the Symphony.")
}
