package app

import (
	"encoding/json"
	"fmt"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

// ExportAppStateAndValidators exports the state of the application for a genesis
// file.
func (app *OsmosisApp) ExportAppStateAndValidators(
	forZeroHeight bool, jailAllowedAddrs []string, modulesToExport []string,
) (servertypes.ExportedApp, error) {
	// as if they could withdraw from the start of the next block
	ctx := app.NewContext(true, tmproto.Header{Height: app.LastBlockHeight()})

	// We export at last height + 1, because that's the height at which
	// Tendermint will start InitChain.
	height := app.LastBlockHeight() + 1
	if forZeroHeight {
		return servertypes.ExportedApp{}, fmt.Errorf("forZeroHeight not supported")
	}

	genState := app.ExportState(ctx)
	appState, err := json.MarshalIndent(genState, "", "  ")
	if err != nil {
		return servertypes.ExportedApp{}, err
	}

	validators, err := staking.WriteValidators(ctx, *app.StakingKeeper)
	return servertypes.ExportedApp{
		AppState:        appState,
		Validators:      validators,
		Height:          height,
		ConsensusParams: app.BaseApp.GetConsensusParams(ctx),
	}, err
}

func (app *OsmosisApp) ExportState(ctx sdk.Context) map[string]json.RawMessage {
	genesisData := make(map[string]json.RawMessage)
	for _, moduleName := range app.mm.OrderExportGenesis {
		// NOTE: the wasm module is making the export is making the
		// export state machine run out of RAM, skip it to allow state
		// export
		if moduleName == "wasm" {
			continue
		}
		genesisData[moduleName] = app.mm.Modules[moduleName].ExportGenesis(ctx, app.AppCodec())
	}
	return genesisData
}
