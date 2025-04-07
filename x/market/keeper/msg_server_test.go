package keeper_test

import (
	"fmt"
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting/assets"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"

	"github.com/osmosis-labs/osmosis/v27/x/market/keeper"
	"github.com/osmosis-labs/osmosis/v27/x/market/types"
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

// TestMsgServer_SwapToNativeCoins tests the case when the user wants to swap from a stable coin to a native coin.
func (s *KeeperTestSuite) TestMsgServer_SwapToNativeCoins() {
	msgServer := s.setupServer()

	// Set Oracle Price
	sdrPriceInMelody := osmomath.NewDecWithPrec(17, 1) // 1 SDR -> 1.7 Melody
	s.App.OracleKeeper.SetMelodyExchangeRate(s.Ctx, assets.MicroSDRDenom, sdrPriceInMelody)

	swapAmountInSDR := sdrPriceInMelody.MulInt64(rand.Int63()%10000 + 2).TruncateInt()
	offerCoin := sdk.NewCoin(assets.MicroSDRDenom, swapAmountInSDR)

	// 1) empty both vaults, expected ErrNotEnoughBalanceOnMarketVaults error
	// Swapping SDR(stable) -> Melody
	swapMsg := types.NewMsgSwap(Addr, offerCoin, appparams.BaseCoinUnit)
	_, err := msgServer.Swap(sdk.WrapSDKContext(s.Ctx), swapMsg)

	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrNotEnoughBalanceOnMarketVaults)
	s.Require().ErrorContains(err, "Market vaults do not have enough coins to swap. Available amount: (main: 0)")

	// 2) Happy case when exchange vault has enough balance
	err = s.App.BankKeeper.SendCoinsFromModuleToModule(s.Ctx, FaucetAccountName, types.ModuleName, sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(30000))))
	s.Require().NoError(err)

	exchangeVaultBalanceBefore := s.App.MarketKeeper.GetExchangePoolBalance(s.Ctx)
	reserveVaultBalanceBefore := s.App.TreasuryKeeper.GetReservePoolBalance(s.Ctx)
	userBalanceBefore := s.App.BankKeeper.GetBalance(s.Ctx, Addr, appparams.BaseCoinUnit)
	sdrSupplyBefore := s.App.BankKeeper.GetSupply(s.Ctx, assets.MicroSDRDenom)

	resp, err := msgServer.Swap(sdk.WrapSDKContext(s.Ctx), swapMsg)
	s.Require().NoError(err)

	exchangeVaultBalanceAfter := s.App.MarketKeeper.GetExchangePoolBalance(s.Ctx)
	reserveVaultBalanceAfter := s.App.TreasuryKeeper.GetReservePoolBalance(s.Ctx)
	userBalanceAfter := s.App.BankKeeper.GetBalance(s.Ctx, Addr, appparams.BaseCoinUnit)
	sdrSupplyAfter := s.App.BankKeeper.GetSupply(s.Ctx, assets.MicroSDRDenom)

	s.Require().Equal(resp.SwapCoin.Amount, userBalanceAfter.Amount.Sub(userBalanceBefore.Amount), "user balance should increase by swap amount")
	s.Require().Equal(resp.SwapCoin.Amount, exchangeVaultBalanceBefore.Amount.Sub(exchangeVaultBalanceAfter.Amount), "all asked amount should be deducted from exchange pool")
	s.Require().False(resp.SwapFee.Amount.IsZero())
	s.Require().Equal(sdrSupplyBefore.Amount.Sub(sdrSupplyAfter.Amount), swapAmountInSDR, "supply should decrease by swap amount since we burn stable coin")
	s.Require().Equal(reserveVaultBalanceBefore.Amount, reserveVaultBalanceAfter.Amount, "reserve pool balance should not change")
}

// TestMsgServer_SwapToNativeBalancePool tests the case when the user wants to swap from a stable coin to a native coin and vica verse.
// In this case, the balance should be the same.
func (s *KeeperTestSuite) TestMsgServer_SwapToNativeBalancePool() {
	msgServer := s.setupServer()

	// Set Oracle Price
	sdrPriceInMelody := osmomath.NewDecWithPrec(17, 1) // 1 SDR -> 1.7 Melody
	s.App.OracleKeeper.SetMelodyExchangeRate(s.Ctx, assets.MicroSDRDenom, sdrPriceInMelody)

	swapAmountInSDR := sdrPriceInMelody.MulInt64(rand.Int63()%10000 + 2).TruncateInt()
	offerCoin := sdk.NewCoin(assets.MicroSDRDenom, swapAmountInSDR)

	swapMsg := types.NewMsgSwap(Addr, offerCoin, appparams.BaseCoinUnit)

	err := s.App.BankKeeper.SendCoinsFromModuleToModule(s.Ctx, FaucetAccountName, types.ModuleName, sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(30000))))
	s.Require().NoError(err)

	exchangeVaultBalanceBefore := s.App.MarketKeeper.GetExchangePoolBalance(s.Ctx)
	reserveVaultBalanceBefore := s.App.TreasuryKeeper.GetReservePoolBalance(s.Ctx)
	userBalanceBefore := s.App.BankKeeper.GetBalance(s.Ctx, Addr, appparams.BaseCoinUnit)

	resp, err := msgServer.Swap(sdk.WrapSDKContext(s.Ctx), swapMsg)
	s.Require().NoError(err)

	offerCoin = resp.SwapCoin
	swapMsg = types.NewMsgSwap(Addr, offerCoin, assets.MicroSDRDenom)

	resp, err = msgServer.Swap(sdk.WrapSDKContext(s.Ctx), swapMsg)
	s.Require().NoError(err)

	exchangeVaultBalanceAfter := s.App.MarketKeeper.GetExchangePoolBalance(s.Ctx)
	reserveVaultBalanceAfter := s.App.TreasuryKeeper.GetReservePoolBalance(s.Ctx)
	userBalanceAfter := s.App.BankKeeper.GetBalance(s.Ctx, Addr, appparams.BaseCoinUnit)

	s.Require().Equal(userBalanceBefore.Amount, userBalanceAfter.Amount, "user balance should not change")
	s.Require().Equal(exchangeVaultBalanceBefore.Amount, exchangeVaultBalanceAfter.Amount, "exchange pool balance should not change")
	s.Require().Equal(reserveVaultBalanceBefore.Amount, reserveVaultBalanceAfter.Amount, "reserve pool balance should not change")
}

// TestMsgServer_SwapStableToStable tests the case when the user wants to swap from a stable coin to a stable coin.
func (s *KeeperTestSuite) TestMsgServer_SwapStableToStable() {
	msgServer := s.setupServer()

	// Set Oracle Price
	sdrPriceInMelody := osmomath.NewDecWithPrec(17, 1) // 1 SDR -> 1.7 Melody
	usdPriceInMelody := osmomath.NewDecWithPrec(13, 1) // 1 USD -> 1.3 Melody
	s.App.OracleKeeper.SetMelodyExchangeRate(s.Ctx, assets.MicroSDRDenom, sdrPriceInMelody)
	s.App.OracleKeeper.SetMelodyExchangeRate(s.Ctx, assets.MicroUSDDenom, usdPriceInMelody)

	swapAmountInSDR := sdrPriceInMelody.MulInt64(rand.Int63()%10000 + 2).TruncateInt()
	offerCoin := sdk.NewCoin(assets.MicroSDRDenom, swapAmountInSDR)

	userBalanceSDRBefore := s.App.BankKeeper.GetBalance(s.Ctx, Addr, assets.MicroSDRDenom)

	swapMsg := types.NewMsgSwap(Addr, offerCoin, assets.MicroUSDDenom)
	resp, err := msgServer.Swap(sdk.WrapSDKContext(s.Ctx), swapMsg)
	s.Require().NoError(err)

	userBalanceSDRAfter := s.App.BankKeeper.GetBalance(s.Ctx, Addr, assets.MicroSDRDenom)
	userBalanceUSDAfter := s.App.BankKeeper.GetBalance(s.Ctx, Addr, assets.MicroUSDDenom)
	s.Require().Equal(resp.SwapCoin, userBalanceUSDAfter, "user balance should increase by swap amount")
	s.Require().Equal(userBalanceSDRBefore.Amount.Sub(userBalanceSDRAfter.Amount), swapAmountInSDR, "user balance should decrease by swap amount")
}

// TestMsgServe_SwapMainPoolEmpty tests the case when the main pool is empty and swap fail.
func (s *KeeperTestSuite) TestMsgServe_SwapMainPoolEmpty() {
	msgServer := s.setupServer()

	// Set Oracle Price
	sdrPriceInMelody := osmomath.NewDecWithPrec(1, 0) // 1 SDR -> 1 Melody
	s.App.OracleKeeper.SetMelodyExchangeRate(s.Ctx, assets.MicroSDRDenom, sdrPriceInMelody)

	swapAmountInSDR := sdrPriceInMelody.MulInt64(appparams.MicroUnit).TruncateInt()
	offerCoin := sdk.NewCoin(assets.MicroSDRDenom, swapAmountInSDR)

	swapMsg := types.NewMsgSwap(Addr, offerCoin, appparams.BaseCoinUnit)

	_, err := msgServer.Swap(sdk.WrapSDKContext(s.Ctx), swapMsg)
	s.Require().Error(err)
}

// TestMsgServe_SwapNotEnoughInReservePool tests the case when there is not enough balance in the reserve pool and swap should fail.
func (s *KeeperTestSuite) TestMsgServe_SwapNotEnoughInReservePool() {
	msgServer := s.setupServer()

	// Set Oracle Price
	sdrPriceInMelody := osmomath.NewDecWithPrec(1, 0) // 1 SDR -> 1 Melody
	s.App.OracleKeeper.SetMelodyExchangeRate(s.Ctx, assets.MicroSDRDenom, sdrPriceInMelody)

	swapAmountInSDR := sdrPriceInMelody.MulInt64(appparams.MicroUnit).TruncateInt()
	offerCoin := sdk.NewCoin(assets.MicroSDRDenom, swapAmountInSDR)

	swapMsg := types.NewMsgSwap(Addr, offerCoin, appparams.BaseCoinUnit)

	err := s.App.BankKeeper.SendCoinsFromModuleToModule(s.Ctx, FaucetAccountName, types.ModuleName, sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(appparams.MicroUnit/2))))
	s.Require().NoError(err)

	resp, err := msgServer.Swap(sdk.WrapSDKContext(s.Ctx), swapMsg)
	s.Require().Error(err)
	s.Require().ErrorContains(err, fmt.Sprintf("Market vaults do not have enough coins to swap. Available amount: (main: %v)", appparams.MicroUnit/2))
	s.Require().Nil(resp)
}
