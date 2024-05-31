package keeper_test

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/osmomath"
	appparams "github.com/osmosis-labs/osmosis/v23/app/params"

	"github.com/osmosis-labs/osmosis/v23/x/market/keeper"
	"github.com/osmosis-labs/osmosis/v23/x/market/types"
)

func (s *KeeperTestSuite) setupServer() types.MsgServer {
	totalSupply := sdk.NewCoins(sdk.NewCoin(appparams.MicroSDRDenom, InitTokens.MulRaw(int64(len(Addr)*10))))
	err := s.App.BankKeeper.MintCoins(s.Ctx, FaucetAccountName, totalSupply)
	s.Require().NoError(err)

	err = s.App.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, FaucetAccountName, Addr, InitUSDRCoins)
	s.Require().NoError(err)

	msgServer := keeper.NewMsgServerImpl(*s.App.MarketKeeper)
	return msgServer
}

func (s *KeeperTestSuite) TestMsgServer_SwapToNativeCoins() {
	msgServer := s.setupServer()

	// Set Oracle Price
	melodyPriceInSDR := sdk.NewDecWithPrec(17, 1)
	s.App.OracleKeeper.SetMelodyExchangeRate(s.Ctx, appparams.MicroSDRDenom, melodyPriceInSDR)

	swapAmountInSDR := melodyPriceInSDR.MulInt64(rand.Int63()%10000 + 2).TruncateInt()
	offerCoin := sdk.NewCoin(appparams.MicroSDRDenom, swapAmountInSDR)

	// 1) empty both vaults, expected ErrNotEnoughBalanceOnMarketVaults error
	swapMsg := types.NewMsgSwap(Addr, offerCoin, appparams.BaseCoinUnit)
	_, err := msgServer.Swap(sdk.WrapSDKContext(s.Ctx), swapMsg)

	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrNotEnoughBalanceOnMarketVaults)
	s.Require().ErrorContains(err, "Market vaults do not have enough coins to swap. Available amount: 0")

	// 2) Happy case when exchange vault has enough balance
	err = s.App.BankKeeper.SendCoinsFromModuleToModule(s.Ctx, FaucetAccountName, types.ModuleName, sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(10000))))
	s.Require().NoError(err)

	exchangeAcc := s.App.MarketKeeper.GetMarketAccount(s.Ctx)
	reserveAcc := s.App.MarketKeeper.GetReserveMarketAccount(s.Ctx)

	exchangeVaultBalanceBefore := s.App.BankKeeper.GetBalance(s.Ctx, exchangeAcc.GetAddress(), appparams.BaseCoinUnit)
	userBalanceBefore := s.App.BankKeeper.GetBalance(s.Ctx, Addr, appparams.BaseCoinUnit)

	resp, err := msgServer.Swap(sdk.WrapSDKContext(s.Ctx), swapMsg)
	s.Require().NoError(err)

	exchangeVaultBalanceAfter := s.App.BankKeeper.GetBalance(s.Ctx, exchangeAcc.GetAddress(), appparams.BaseCoinUnit)
	reserveVaultBalanceAfter := s.App.BankKeeper.GetBalance(s.Ctx, reserveAcc.GetAddress(), appparams.BaseCoinUnit)
	userBalanceAfter := s.App.BankKeeper.GetBalance(s.Ctx, Addr, appparams.BaseCoinUnit)

	s.Require().Equal(resp.SwapCoin.Amount, userBalanceAfter.Amount.Sub(userBalanceBefore.Amount))
	s.Require().Equal(resp.SwapFee.Amount, reserveVaultBalanceAfter.Amount)
	s.Require().Equal(resp.SwapCoin.Amount.Add(resp.SwapFee.Amount), exchangeVaultBalanceBefore.Amount.Sub(exchangeVaultBalanceAfter.Amount))
}
