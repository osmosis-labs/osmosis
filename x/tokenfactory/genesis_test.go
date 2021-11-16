package tokenfactory_test

import (
	"testing"

	simapp "github.com/osmosis-labs/osmosis/app"
	appparams "github.com/osmosis-labs/osmosis/app/params"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/osmosis-labs/osmosis/x/tokenfactory/types"
)

func TestGenesis(t *testing.T) {
	appparams.SetAddressPrefixes()

	genesisState := types.GenesisState{
		FactoryDenoms: []types.GenesisDenom{
			{
				Denom: "factory/osmo1t7egva48prqmzl59x5ngv4zx0dtrwewc9m7z44/bitcoin",
				AuthorityMetadata: types.DenomAuthorityMetadata{
					Admin: "osmo1t7egva48prqmzl59x5ngv4zx0dtrwewc9m7z44",
				},
			},
			{
				Denom: "factory/osmo1t7egva48prqmzl59x5ngv4zx0dtrwewc9m7z44/litecoin",
				AuthorityMetadata: types.DenomAuthorityMetadata{
					Admin: "osmo1t7egva48prqmzl59x5ngv4zx0dtrwewc9m7z44",
				},
			},
		},
	}
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	// tokenfactory.InitGenesis(ctx, *k, genesisState)
	// got := tokenfactory.ExportGenesis(ctx, *k)
	// require.NotNil(t, got)
	// require.Equal(t, genesisState, got)
}
