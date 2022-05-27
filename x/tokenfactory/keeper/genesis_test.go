package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	simapp "github.com/osmosis-labs/osmosis/v8/app"
	appparams "github.com/osmosis-labs/osmosis/v8/app/params"

	"github.com/osmosis-labs/osmosis/v8/x/tokenfactory/types"
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

	app.TokenFactoryKeeper.InitGenesis(ctx, genesisState)
	exportedGenesis := app.TokenFactoryKeeper.ExportGenesis(ctx)
	require.NotNil(t, exportedGenesis)
	require.Equal(t, genesisState, *exportedGenesis)
}
