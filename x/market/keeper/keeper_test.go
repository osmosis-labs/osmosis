package keeper_test

import (
	"testing"

	"github.com/cometbft/cometbft/crypto/secp256k1"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/osmosis-labs/osmosis/v23/app/apptesting"
	appparams "github.com/osmosis-labs/osmosis/v23/app/params"
	tokenfactorytypes "github.com/osmosis-labs/osmosis/v23/x/tokenfactory/types"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	pubKey = secp256k1.GenPrivKey().PubKey()
	Addr   = sdk.AccAddress(pubKey.Address())

	InitTokens    = sdk.TokensFromConsensusPower(200, sdk.DefaultPowerReduction)
	InitBaseCoins = sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, InitTokens))
	InitUSDRCoins = sdk.NewCoins(sdk.NewCoin(appparams.MicroSDRDenom, InitTokens))

	FaucetAccountName = tokenfactorytypes.ModuleName
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()

	// Set the bond denom to be uosmo to make volume tracking tests more readable.
	skParams := s.App.StakingKeeper.GetParams(s.Ctx)
	skParams.BondDenom = "uosmo"
	s.App.StakingKeeper.SetParams(s.Ctx, skParams)
	s.App.TxFeesKeeper.SetBaseDenom(s.Ctx, "uosmo")
	marketParams := s.App.MarketKeeper.GetParams(s.Ctx)
	s.App.MarketKeeper.SetParams(s.Ctx, marketParams)

	totalSupply := sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, InitTokens.MulRaw(int64(len(Addr)*10))))
	err := s.App.BankKeeper.MintCoins(s.Ctx, FaucetAccountName, totalSupply)
	s.Require().NoError(err)

	s.App.AccountKeeper.SetAccount(s.Ctx, authtypes.NewBaseAccountWithAddress(Addr))

	err = s.App.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, FaucetAccountName, Addr, InitBaseCoins)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestOsmosisPoolDeltaUpdate() {
	terraPoolDelta := s.App.MarketKeeper.GetOsmosisPoolDelta(s.Ctx)
	s.Require().Equal(sdk.ZeroDec(), terraPoolDelta)

	diff := sdk.NewDec(10)
	s.App.MarketKeeper.SetOsmosisPoolDelta(s.Ctx, diff)

	terraPoolDelta = s.App.MarketKeeper.GetOsmosisPoolDelta(s.Ctx)
	s.Require().Equal(diff, terraPoolDelta)
}

// TestReplenishPools tests that
// each pools move towards base pool
func (s *KeeperTestSuite) TestReplenishPools() {
	s.App.OracleKeeper.SetOsmoExchangeRate(s.Ctx, appparams.StakeDenom, sdk.OneDec())

	basePool := s.App.MarketKeeper.BasePool(s.Ctx)
	terraPoolDelta := s.App.MarketKeeper.GetOsmosisPoolDelta(s.Ctx)
	s.Require().True(terraPoolDelta.IsZero())

	// Positive delta
	diff := basePool.QuoInt64((int64)(appparams.BlocksPerDay))
	s.App.MarketKeeper.SetOsmosisPoolDelta(s.Ctx, diff)

	s.App.MarketKeeper.ReplenishPools(s.Ctx)

	terraPoolDelta = s.App.MarketKeeper.GetOsmosisPoolDelta(s.Ctx)
	replenishAmt := diff.QuoInt64((int64)(s.App.MarketKeeper.PoolRecoveryPeriod(s.Ctx)))
	expectedDelta := diff.Sub(replenishAmt)
	s.Require().Equal(expectedDelta, terraPoolDelta)

	// Negative delta
	diff = diff.Neg()
	s.App.MarketKeeper.SetOsmosisPoolDelta(s.Ctx, diff)

	s.App.MarketKeeper.ReplenishPools(s.Ctx)

	osmosisPoolDelta := s.App.MarketKeeper.GetOsmosisPoolDelta(s.Ctx)
	replenishAmt = diff.QuoInt64((int64)(s.App.MarketKeeper.PoolRecoveryPeriod(s.Ctx)))
	expectedDelta = diff.Sub(replenishAmt)
	s.Require().Equal(expectedDelta, osmosisPoolDelta)
}
