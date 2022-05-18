package tokenfactory_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	simapp "github.com/osmosis-labs/osmosis/v9/app"
	appparams "github.com/osmosis-labs/osmosis/v9/app/params"

	"github.com/osmosis-labs/osmosis/v9/x/tokenfactory"
	"github.com/osmosis-labs/osmosis/v9/x/tokenfactory/types"
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

	tokenfactory.InitGenesis(ctx, *app.TokenFactoryKeeper, genesisState)
	exportedGenesis := tokenfactory.ExportGenesis(ctx, *app.TokenFactoryKeeper)
	require.NotNil(t, exportedGenesis)
	require.Equal(t, genesisState, *exportedGenesis)
}
