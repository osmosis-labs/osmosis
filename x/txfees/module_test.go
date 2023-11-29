package txfees_test

import (
	"encoding/json"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/stretchr/testify/require"

	osmosisapp "github.com/osmosis-labs/osmosis/v21/app"

	simapp "github.com/osmosis-labs/osmosis/v21/app"
)

func TestSetBaseDenomOnInitBlock(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	genesisState := osmosisapp.GenesisStateWithValSet(app)
	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	if err != nil {
		panic(err)
	}

	app.InitChain(
		abcitypes.RequestInitChain{
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: sims.DefaultConsensusParams,
			AppStateBytes:   stateBytes,
			ChainId:         "osmosis-1",
		},
	)

	baseDenom, err := app.TxFeesKeeper.GetBaseDenom(ctx)
	require.Nil(t, err)
	require.NotEmpty(t, baseDenom)
}
