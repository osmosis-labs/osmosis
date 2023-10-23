package app

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
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

	genStateDir, err := app.mm.ExportGenesisForModules(ctx, app.appCodec, modulesToExport)
	if err != nil {
		return servertypes.ExportedApp{}, err
	}

	// Stream the data from the files in genStateDir when marshalling the AppState
	fmt.Println("streamAndMarshalAppState")
	appState, err := streamAndMarshalAppState(genStateDir)
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

func streamAndMarshalAppState(genStateDir string) ([]byte, error) {
	genesisData := make(map[string]json.RawMessage)

	err := filepath.Walk(genStateDir, func(path string, info os.FileInfo, err error) error {
		fmt.Println("filepath.Walks", path)
		fmt.Println("filepath.Walks", info)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			data, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			// Skip if data is empty
			fmt.Println("check if skip empty file", path)
			if len(data) == 0 {
				fmt.Println("skipping empty file", path)
				return nil
			}

			moduleName := filepath.Base(path)
			genesisData[moduleName] = json.RawMessage(data)
		}

		return nil
	})

	fmt.Println("json.MarshalIndent")
	appState, err := json.MarshalIndent(genesisData, "", "  ")
	if err != nil {
		return nil, err
	}
	fmt.Println("json.MarshalIndent fin")

	return appState, nil
}

// func (app *OsmosisApp) ExportState(ctx sdk.Context) map[string]json.RawMessage {
// 	return app.mm.ExportGenesis(ctx, app.AppCodec())
// }
