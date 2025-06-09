package keeper_test

import (
	"context"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting/assets"
	"github.com/osmosis-labs/osmosis/v27/x/stablestaking/types"
)

type GRPCQueryTestSuite struct {
	apptesting.KeeperTestHelper
	queryClient types.QueryClient
}

func TestGRPCQueryTestSuite(t *testing.T) {
	suite.Run(t, new(GRPCQueryTestSuite))
}

func (s *GRPCQueryTestSuite) SetupTest() {
	s.Setup()
	// Set Oracle Price
	sdrPriceInMelody := osmomath.NewDecWithPrec(17, 1)
	s.App.OracleKeeper.SetMelodyExchangeRate(s.Ctx, assets.MicroUSDDenom, sdrPriceInMelody)

	// Mint initial tokens
	totalUsddSupply := sdk.NewCoins(sdk.NewCoin(assets.MicroUSDDenom, InitTokens.MulRaw(int64(len(Addr)*10))))
	err := s.App.BankKeeper.MintCoins(s.Ctx, FaucetAccountName, totalUsddSupply)
	s.Require().NoError(err)

	// Fund test accounts
	staker := sdk.AccAddress("staker1")
	err = s.App.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, FaucetAccountName, staker, InitUSDDCoins)
	s.Require().NoError(err)

	s.queryClient = types.NewQueryClient(s.QueryHelper)
}

func (s *GRPCQueryTestSuite) TestParams() {
	resp, err := s.queryClient.Params(context.Background(), &types.QueryParamsRequest{})
	require.NoError(s.T(), err)
	require.NotNil(s.T(), resp)
	require.NotNil(s.T(), resp.Params)
}

func (s *GRPCQueryTestSuite) TestStakingPools() {
	resp, err := s.queryClient.StablePools(context.Background(), &types.QueryPoolsRequest{})
	require.NoError(s.T(), err)
	require.NotNil(s.T(), resp)
	require.NotNil(s.T(), resp.Pools)
}

func (s *GRPCQueryTestSuite) TestStakingPool() {
	// First create a staking pool
	staker := sdk.AccAddress("staker1")
	token := sdk.NewCoin(assets.MicroUSDDenom, math.NewInt(100))
	_, err := s.App.StableStakingKeeper.StakeTokens(s.Ctx, staker, token)
	require.NoError(s.T(), err)

	// Query the pool
	resp, err := s.queryClient.StablePool(context.Background(), &types.QueryPoolRequest{
		Denom: assets.MicroUSDDenom,
	})
	require.NoError(s.T(), err)
	require.NotNil(s.T(), resp)
	require.NotNil(s.T(), resp.Pool)
	require.Equal(s.T(), assets.MicroUSDDenom, resp.Pool.Denom)
	require.Equal(s.T(), math.LegacyNewDecFromInt(math.NewInt(100)), resp.Pool.TotalShares)
	require.Equal(s.T(), math.NewInt(100), resp.Pool.TotalStaked)
}

func (s *GRPCQueryTestSuite) TestUserStakes() {
	staker := sdk.AccAddress("staker1")
	token := sdk.NewCoin(assets.MicroUSDDenom, math.NewInt(100))
	_, err := s.App.StableStakingKeeper.StakeTokens(s.Ctx, staker, token)
	require.NoError(s.T(), err)

	resp, err := s.queryClient.UserStake(context.Background(), &types.QueryUserStakeRequest{
		Address: staker.String(),
	})
	require.NoError(s.T(), err)
	require.NotNil(s.T(), resp)
	require.Len(s.T(), resp.Stakes, 1)
	require.Equal(s.T(), math.LegacyNewDecFromInt(math.NewInt(100)), resp.Stakes.Shares)
}

func (s *GRPCQueryTestSuite) TestUserUnbonding() {
	staker := sdk.AccAddress("staker1")
	token := sdk.NewCoin(assets.MicroUSDDenom, math.NewInt(100))
	_, err := s.App.StableStakingKeeper.StakeTokens(s.Ctx, staker, token)
	require.NoError(s.T(), err)

	// Unbond some tokens
	unbondAmount := math.NewInt(50)
	_, err = s.App.StableStakingKeeper.UnStakeTokens(s.Ctx, staker, sdk.NewCoin(assets.MicroUSDDenom, unbondAmount))
	require.NoError(s.T(), err)

	resp, err := s.queryClient.UserUnbonding(context.Background(), &types.QueryUserUnbondingRequest{
		Address: staker.String(),
		Denom:   assets.MicroUSDDenom,
	})
	require.NoError(s.T(), err)
	require.NotNil(s.T(), resp)
	require.Equal(s.T(), unbondAmount, resp.Info.Amount)
	require.Equal(s.T(), s.Ctx.BlockTime().Day()+12, resp.Info.UnbondEpoch)
}

func (s *GRPCQueryTestSuite) TestUserTotalUnbonding() {
	staker := sdk.AccAddress("staker1")
	token := sdk.NewCoin(assets.MicroUSDDenom, math.NewInt(100))
	_, err := s.App.StableStakingKeeper.StakeTokens(s.Ctx, staker, token)
	require.NoError(s.T(), err)

	// Unbond some tokens
	unbondAmount := math.NewInt(50)
	_, err = s.App.StableStakingKeeper.UnStakeTokens(s.Ctx, staker, sdk.NewCoin(assets.MicroUSDDenom, unbondAmount))
	require.NoError(s.T(), err)

	resp, err := s.queryClient.UserTotalUnbonding(context.Background(), &types.QueryUserTotalUnbondingRequest{
		Address: staker.String(),
	})
	require.NoError(s.T(), err)
	require.NotNil(s.T(), resp)
	require.Len(s.T(), resp.Info, 1)
	require.Equal(s.T(), assets.MicroUSDDenom, resp.Info[0].Denom)
	require.Equal(s.T(), unbondAmount, resp.Info[0].Amount)
}
