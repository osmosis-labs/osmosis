package keeper_test

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v32/x/mint/types"
	txfeestypes "github.com/osmosis-labs/osmosis/v32/x/txfees/types"
)

// mintDenom is the bond/mint denom used by the app test harness.
func (s *KeeperTestSuite) mintDenom() string {
	return s.App.MintKeeper.GetParams(s.Ctx).MintDenom
}

// sendToBurnAddress mints amt of the mint denom and sends it to the null
// (burn) address.
func (s *KeeperTestSuite) sendToBurnAddress(amt osmomath.Int) {
	coins := sdk.NewCoins(sdk.NewCoin(s.mintDenom(), amt))
	s.MintCoins(coins)
	err := s.App.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, types.ModuleName, txfeestypes.DefaultNullAddress, coins)
	s.Require().NoError(err)
}

// TestGetMintedSupply asserts minted supply is the bank supply-with-offset plus
// the unvested developer-vesting balance, and that OSMO minted behind a
// negative supply offset (as superfluid staking does for synthetic stake) is
// not counted as minted.
func (s *KeeperTestSuite) TestGetMintedSupply() {
	s.SetupTest()
	denom := s.mintDenom()

	devVestingAddr := s.App.AccountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName)
	s.Require().NotNil(devVestingAddr)
	devVestingBalance := s.App.BankKeeper.GetBalance(s.Ctx, devVestingAddr, denom).Amount
	withOffset := s.App.BankKeeper.GetSupplyWithOffset(s.Ctx, denom).Amount

	minted := s.App.MintKeeper.GetMintedSupply(s.Ctx)
	s.Require().Equal(withOffset.Add(devVestingBalance), minted)

	// Mint synthetic OSMO and offset it, as superfluid delegation does. Raw
	// bank supply rises, but minted supply must not.
	syntheticAmt := osmomath.NewInt(3_000_000)
	rawBefore := s.App.BankKeeper.GetSupply(s.Ctx, denom).Amount
	s.MintCoins(sdk.NewCoins(sdk.NewCoin(denom, syntheticAmt)))
	s.App.BankKeeper.AddSupplyOffset(s.Ctx, denom, syntheticAmt.Neg())

	s.Require().Equal(rawBefore.Add(syntheticAmt), s.App.BankKeeper.GetSupply(s.Ctx, denom).Amount)
	s.Require().Equal(minted, s.App.MintKeeper.GetMintedSupply(s.Ctx))
}

// TestGetBurnedSupply asserts burned supply equals the balance of the null
// (burn) address, and tracks coins sent there.
func (s *KeeperTestSuite) TestGetBurnedSupply() {
	s.SetupTest()

	// Initially the burn address holds nothing of the mint denom.
	s.Require().True(s.App.MintKeeper.GetBurnedSupply(s.Ctx).IsZero())

	burnAmt := osmomath.NewInt(5_550_000)
	s.sendToBurnAddress(burnAmt)

	s.Require().Equal(burnAmt, s.App.MintKeeper.GetBurnedSupply(s.Ctx))
}

// TestGetTotalSupply asserts total supply == minted - burned, and that sending
// existing coins to the burn address lowers total supply without changing
// minted supply.
func (s *KeeperTestSuite) TestGetTotalSupply() {
	s.SetupTest()

	burnAmt := osmomath.NewInt(4_000_000)
	coins := sdk.NewCoins(sdk.NewCoin(s.mintDenom(), burnAmt))
	s.FundAcc(testAddressOne, coins)

	mintedBefore := s.App.MintKeeper.GetMintedSupply(s.Ctx)
	totalBefore := s.App.MintKeeper.GetTotalSupply(s.Ctx)
	s.Require().Equal(mintedBefore.Sub(s.App.MintKeeper.GetBurnedSupply(s.Ctx)), totalBefore)

	err := s.App.BankKeeper.SendCoins(s.Ctx, testAddressOne, txfeestypes.DefaultNullAddress, coins)
	s.Require().NoError(err)

	s.Require().Equal(mintedBefore, s.App.MintKeeper.GetMintedSupply(s.Ctx))
	s.Require().Equal(totalBefore.Sub(burnAmt), s.App.MintKeeper.GetTotalSupply(s.Ctx))
	s.Require().Equal(burnAmt, s.App.MintKeeper.GetBurnedSupply(s.Ctx))
}

// TestSupplyQueryHandlers exercises the supply gRPC handlers end to end through
// the query client.
func (s *KeeperTestSuite) TestSupplyQueryHandlers() {
	s.SetupTest()
	s.sendToBurnAddress(osmomath.NewInt(1_000_000))

	mintedRes, err := s.queryClient.MintedSupply(context.Background(), &types.QueryMintedSupplyRequest{})
	s.Require().NoError(err)
	s.Require().Equal(s.App.MintKeeper.GetMintedSupply(s.Ctx), mintedRes.MintedSupply)

	burnedRes, err := s.queryClient.BurnedSupply(context.Background(), &types.QueryBurnedRequest{})
	s.Require().NoError(err)
	s.Require().Equal(s.App.MintKeeper.GetBurnedSupply(s.Ctx), burnedRes.Burned)

	totalRes, err := s.queryClient.TotalSupply(context.Background(), &types.QueryTotalSupplyRequest{})
	s.Require().NoError(err)
	s.Require().Equal(mintedRes.MintedSupply.Sub(burnedRes.Burned), totalRes.TotalSupply)
}
