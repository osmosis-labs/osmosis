package cmd

// DONTCOVER

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/config"
	tmjson "github.com/tendermint/tendermint/libs/json"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/types"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	FlagHeight           = "height"
	FlagForZeroHeight    = "for-zero-height"
	FlagJailAllowedAddrs = "jail-allowed-addrs"
	FlagModulesToExport  = "modules-to-export"
)

func openDB(rootDir string) (dbm.DB, error) {
	dataDir := filepath.Join(rootDir, "data")
	return sdk.NewLevelDB("application", dataDir)
}

// FasterExportCmd dumps app state to JSON.
func FasterExportCmd(appExporter types.AppExporter, defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export state to JSON",
		RunE: func(cmd *cobra.Command, args []string) error {
			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			homeDir, _ := cmd.Flags().GetString(flags.FlagHome)
			config.SetRoot(homeDir)

			if _, err := os.Stat(config.GenesisFile()); os.IsNotExist(err) {
				return err
			}

			height, _ := cmd.Flags().GetInt64(FlagHeight)
			modulesToExport, _ := cmd.Flags().GetStringSlice(FlagModulesToExport)

			return exportLogic(serverCtx.Logger, cmd, serverCtx.Viper, config, appExporter, height, modulesToExport)
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	cmd.Flags().String("out", "", "File to output to, if blank prints to stderr")
	cmd.Flags().Int64(FlagHeight, -1, "Export state from a particular height (-1 means latest height)")
	cmd.Flags().StringSlice(FlagModulesToExport, []string{}, "Comma-separated list of modules to export. If empty, will export all modules")

	return cmd
}

type printer interface {
	Println(i ...interface{})
}

func exportLogic(logger log.Logger, cmd printer, appOpts servertypes.AppOptions, config *config.Config, appExporter types.AppExporter, height int64, modulesToExport []string) error {
	db, err := openDB(config.RootDir)
	if err != nil {
		return err
	}

	if appExporter == nil {
		if _, err := fmt.Fprintln(os.Stderr, "WARNING: App exporter not defined. Returning genesis file."); err != nil {
			return err
		}

		genesis, err := ioutil.ReadFile(config.GenesisFile())
		if err != nil {
			return err
		}

		fmt.Println(string(genesis))
		return nil
	}

	forZeroHeight := false
	jailAllowedAddrs := []string{}

	var dummy io.Writer
	exported, err := appExporter(logger, db, dummy, height, forZeroHeight, jailAllowedAddrs, appOpts, modulesToExport)
	if err != nil {
		return fmt.Errorf("error exporting state: %v", err)
	}

	doc, err := tmtypes.GenesisDocFromFile(config.GenesisFile())
	if err != nil {
		return err
	}

	doc.AppState = exported.AppState
	doc.Validators = exported.Validators
	doc.InitialHeight = exported.Height
	doc.ConsensusParams = &tmproto.ConsensusParams{
		Block: tmproto.BlockParams{
			MaxBytes:   exported.ConsensusParams.Block.MaxBytes,
			MaxGas:     exported.ConsensusParams.Block.MaxGas,
			TimeIotaMs: doc.ConsensusParams.Block.TimeIotaMs,
		},
		Evidence: tmproto.EvidenceParams{
			MaxAgeNumBlocks: exported.ConsensusParams.Evidence.MaxAgeNumBlocks,
			MaxAgeDuration:  exported.ConsensusParams.Evidence.MaxAgeDuration,
			MaxBytes:        exported.ConsensusParams.Evidence.MaxBytes,
		},
		Validator: tmproto.ValidatorParams{
			PubKeyTypes: exported.ConsensusParams.Validator.PubKeyTypes,
		},
	}

	// NOTE: Tendermint uses a custom JSON decoder for GenesisDoc
	// (except for stuff inside AppState). Inside AppState, we're free
	// to encode as protobuf or amino.
	encoded, err := tmjson.Marshal(doc)
	if err != nil {
		return err
	}

	fmt.Println(string(encoded))
	// cmd.Println(string(sdk.MustSortJSON(encoded)))
	return nil
}
