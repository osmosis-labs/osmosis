package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/osmosis-labs/osmosis/v23/x/tokenfactory/types"
)

func (s *KeeperTestSuite) TestGenesis() {
	genesisState := types.GenesisState{
		FactoryDenoms: []types.GenesisDenom{
			{
				Denom: "factory/symphony1p822vyk8ylf3hpwh9qgv6p6dft7hedntyqyw7w/bitcoin",
				AuthorityMetadata: types.DenomAuthorityMetadata{
					Admin: "symphony1p822vyk8ylf3hpwh9qgv6p6dft7hedntyqyw7w",
				},
			},
			{
				Denom: "factory/symphony1p822vyk8ylf3hpwh9qgv6p6dft7hedntyqyw7w/diff-admin",
				AuthorityMetadata: types.DenomAuthorityMetadata{
					Admin: "symphony167psfumv4erluc3ksfw23n2pdrx6zx3rj2guvl",
				},
			},
			{
				Denom: "factory/symphony1p822vyk8ylf3hpwh9qgv6p6dft7hedntyqyw7w/litecoin",
				AuthorityMetadata: types.DenomAuthorityMetadata{
					Admin: "symphony1p822vyk8ylf3hpwh9qgv6p6dft7hedntyqyw7w",
				},
			},
		},
	}

	s.SetupTestForInitGenesis()
	app := s.App

	// Test both with bank denom metadata set, and not set.
	for i, denom := range genesisState.FactoryDenoms {
		// hacky, sets bank metadata to exist if i != 0, to cover both cases.
		if i != 0 {
			app.BankKeeper.SetDenomMetaData(s.Ctx, banktypes.Metadata{
				DenomUnits: []*banktypes.DenomUnit{{
					Denom:    denom.GetDenom(),
					Exponent: 0,
				}},
				Base:    denom.GetDenom(),
				Display: denom.GetDenom(),
				Name:    denom.GetDenom(),
				Symbol:  denom.GetDenom(),
			})
		}
	}

	// check before initGenesis that the module account is nil
	tokenfactoryModuleAccount := app.AccountKeeper.GetAccount(s.Ctx, app.AccountKeeper.GetModuleAddress(types.ModuleName))
	s.Require().Nil(tokenfactoryModuleAccount)

	app.TokenFactoryKeeper.SetParams(s.Ctx, types.Params{DenomCreationFee: sdk.Coins{sdk.NewInt64Coin("note", 100)}})
	app.TokenFactoryKeeper.InitGenesis(s.Ctx, genesisState)

	// check that the module account is now initialized
	tokenfactoryModuleAccount = app.AccountKeeper.GetAccount(s.Ctx, app.AccountKeeper.GetModuleAddress(types.ModuleName))
	s.Require().NotNil(tokenfactoryModuleAccount)

	exportedGenesis := app.TokenFactoryKeeper.ExportGenesis(s.Ctx)
	s.Require().NotNil(exportedGenesis)
	s.Require().Equal(genesisState, *exportedGenesis)

	// verify that the exported bank genesis is valid
	app.BankKeeper.SetParams(s.Ctx, banktypes.DefaultParams())
	exportedBankGenesis := app.BankKeeper.ExportGenesis(s.Ctx)
	s.Require().NoError(exportedBankGenesis.Validate())

	app.BankKeeper.InitGenesis(s.Ctx, exportedBankGenesis)
	for i, denom := range genesisState.FactoryDenoms {
		// hacky, check whether bank metadata is not replaced if i != 0, to cover both cases.
		if i != 0 {
			metadata, found := app.BankKeeper.GetDenomMetaData(s.Ctx, denom.GetDenom())
			s.Require().True(found)
			s.Require().EqualValues(metadata, banktypes.Metadata{
				DenomUnits: []*banktypes.DenomUnit{{
					Denom:    denom.GetDenom(),
					Exponent: 0,
				}},
				Base:    denom.GetDenom(),
				Display: denom.GetDenom(),
				Name:    denom.GetDenom(),
				Symbol:  denom.GetDenom(),
			})
		}
	}
}
