package keeper_test

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v23/app/apptesting/assets"
	appparams "github.com/osmosis-labs/osmosis/v23/app/params"

	"github.com/osmosis-labs/osmosis/v23/x/market/keeper"
	"github.com/osmosis-labs/osmosis/v23/x/market/types"
)

func (s *KeeperTestSuite) setupServer() types.MsgServer {
	totalSupply := sdk.NewCoins(sdk.NewCoin(assets.MicroSDRDenom, InitTokens.MulRaw(int64(len(Addr)*10))))
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
	sdrPriceInMelody := sdk.NewDecWithPrec(17, 1) // 1 SDR -> 1.7 Melody
	s.App.OracleKeeper.SetMelodyExchangeRate(s.Ctx, assets.MicroSDRDenom, sdrPriceInMelody)

	swapAmountInSDR := sdrPriceInMelody.MulInt64(rand.Int63()%10000 + 2).TruncateInt()
	offerCoin := sdk.NewCoin(assets.MicroSDRDenom, swapAmountInSDR)

	// 1) empty both vaults, expected ErrNotEnoughBalanceOnMarketVaults error
	// Swapping SDR(stable) -> Melody
	swapMsg := types.NewMsgSwap(Addr, offerCoin, appparams.BaseCoinUnit)
	_, err := msgServer.Swap(sdk.WrapSDKContext(s.Ctx), swapMsg)

	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrNotEnoughBalanceOnMarketVaults)
	s.Require().ErrorContains(err, "Market vaults do not have enough coins to swap. Available amount: (main: 0), (reserve: 0)")

	// 2) Happy case when exchange vault has enough balance
	err = s.App.BankKeeper.SendCoinsFromModuleToModule(s.Ctx, FaucetAccountName, types.ModuleName, sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(30000))))
	s.Require().NoError(err)

	exchangeVaultBalanceBefore := s.App.MarketKeeper.GetExchangePoolBalance(s.Ctx)
	reserveVaultBalanceBefore := s.App.MarketKeeper.GetReservePoolBalance(s.Ctx)
	userBalanceBefore := s.App.BankKeeper.GetBalance(s.Ctx, Addr, appparams.BaseCoinUnit)
	sdrSupplyBefore := s.App.BankKeeper.GetSupply(s.Ctx, assets.MicroSDRDenom)

	resp, err := msgServer.Swap(sdk.WrapSDKContext(s.Ctx), swapMsg)
	s.Require().NoError(err)

	exchangeVaultBalanceAfter := s.App.MarketKeeper.GetExchangePoolBalance(s.Ctx)
	reserveVaultBalanceAfter := s.App.MarketKeeper.GetReservePoolBalance(s.Ctx)
	userBalanceAfter := s.App.BankKeeper.GetBalance(s.Ctx, Addr, appparams.BaseCoinUnit)
	sdrSupplyAfter := s.App.BankKeeper.GetSupply(s.Ctx, assets.MicroSDRDenom)

	s.Require().Equal(resp.SwapCoin.Amount, userBalanceAfter.Amount.Sub(userBalanceBefore.Amount))
	//s.Require().Equal(resp.SwapFee.Amount, reserveVaultBalanceAfter.Amount)
	s.Require().Equal(resp.SwapCoin.Amount.Add(resp.SwapFee.Amount), exchangeVaultBalanceBefore.Amount.Sub(exchangeVaultBalanceAfter.Amount))
	s.Require().Equal(sdrSupplyBefore.Amount.Sub(sdrSupplyAfter.Amount), swapAmountInSDR, "supply should decrease by swap amount since we burn stable coin")
	s.Require().Equal(reserveVaultBalanceBefore.Amount, reserveVaultBalanceAfter.Amount)
	s.Require().True(reserveVaultBalanceBefore.IsZero())
}

func (s *KeeperTestSuite) TestMsgServer_SwapToNativeBalancePool() {
	msgServer := s.setupServer()

	// Set Oracle Price
	sdrPriceInMelody := sdk.NewDecWithPrec(17, 1) // 1 SDR -> 1.7 Melody
	s.App.OracleKeeper.SetMelodyExchangeRate(s.Ctx, assets.MicroSDRDenom, sdrPriceInMelody)

	swapAmountInSDR := sdrPriceInMelody.MulInt64(rand.Int63()%10000 + 2).TruncateInt()
	offerCoin := sdk.NewCoin(assets.MicroSDRDenom, swapAmountInSDR)

	swapMsg := types.NewMsgSwap(Addr, offerCoin, appparams.BaseCoinUnit)

	err := s.App.BankKeeper.SendCoinsFromModuleToModule(s.Ctx, FaucetAccountName, types.ModuleName, sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(30000))))
	s.Require().NoError(err)

	exchangeVaultBalanceBefore := s.App.MarketKeeper.GetExchangePoolBalance(s.Ctx)
	userBalanceBefore := s.App.BankKeeper.GetBalance(s.Ctx, Addr, appparams.BaseCoinUnit)

	resp, err := msgServer.Swap(sdk.WrapSDKContext(s.Ctx), swapMsg)
	s.Require().NoError(err)

	offerCoin = resp.SwapCoin
	swapMsg = types.NewMsgSwap(Addr, offerCoin, assets.MicroSDRDenom)

	resp, err = msgServer.Swap(sdk.WrapSDKContext(s.Ctx), swapMsg)
	s.Require().NoError(err)

	exchangeVaultBalanceAfter := s.App.MarketKeeper.GetExchangePoolBalance(s.Ctx)
	userBalanceAfter := s.App.BankKeeper.GetBalance(s.Ctx, Addr, appparams.BaseCoinUnit)

	s.Require().Equal(userBalanceBefore.Amount, userBalanceAfter.Amount)
	s.Require().Equal(exchangeVaultBalanceBefore.Amount, exchangeVaultBalanceAfter.Amount)
}
