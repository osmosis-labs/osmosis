package keeper_test

import (
	"testing"

	"github.com/cometbft/cometbft/crypto/secp256k1"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting/assets"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	tokenfactorytypes "github.com/osmosis-labs/osmosis/v27/x/tokenfactory/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	pubKey = secp256k1.GenPrivKey().PubKey()
	Addr   = sdk.AccAddress(pubKey.Address())

	InitTokens    = sdk.TokensFromConsensusPower(200, sdk.DefaultPowerReduction)
	InitBaseCoins = sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, InitTokens))
	InitUSDRCoins = sdk.NewCoins(sdk.NewCoin(assets.MicroSDRDenom, InitTokens))

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

	// Set the bond denom to be note to make volume tracking tests more readable.
	skParams, err := s.App.StakingKeeper.GetParams(s.Ctx)
	s.Require().NoError(err)
	skParams.BondDenom = "note"
	err = s.App.StakingKeeper.SetParams(s.Ctx, skParams)
	s.Require().NoError(err)
	marketParams := s.App.MarketKeeper.GetParams(s.Ctx)
	s.App.MarketKeeper.SetParams(s.Ctx, marketParams)

	totalSupply := sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, InitTokens.MulRaw(int64(len(Addr)*10))))
	err = s.App.BankKeeper.MintCoins(s.Ctx, FaucetAccountName, totalSupply)
	s.Require().NoError(err)

	acc := authtypes.NewBaseAccountWithAddress(Addr)
	acc.AccountNumber = 1000

	s.App.AccountKeeper.SetAccount(s.Ctx, acc)

	err = s.App.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, FaucetAccountName, Addr, InitBaseCoins)
	s.Require().NoError(err)
}
