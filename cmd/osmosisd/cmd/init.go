package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cosmos/go-bip39"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	tmcfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/cli"
	tmrand "github.com/tendermint/tendermint/libs/rand"

	// import genesis state
	_ "github.com/osmosis-labs/osmosis/v8/networks/osmosis-1"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
)

const (
	// FlagOverwrite defines a flag to overwrite an existing genesis JSON file.
	FlagOverwrite = "overwrite"

	// FlagSeed defines a flag to initialize the private validator key from a specific seed.
	FlagRecover = "recover"
)

type printInfo struct {
	Moniker    string          `json:"moniker" yaml:"moniker"`
	ChainID    string          `json:"chain_id" yaml:"chain_id"`
	NodeID     string          `json:"node_id" yaml:"node_id"`
	GenTxsDir  string          `json:"gentxs_dir" yaml:"gentxs_dir"`
	AppMessage json.RawMessage `json:"app_message" yaml:"app_message"`
}

func newPrintInfo(moniker, chainID, nodeID, genTxsDir string) printInfo {
	return printInfo{
		Moniker:   moniker,
		ChainID:   chainID,
		NodeID:    nodeID,
		GenTxsDir: genTxsDir,
	}
}

func displayInfo(info printInfo) error {
	out, err := json.MarshalIndent(info, "", " ")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(os.Stderr, "%s\n", string(sdk.MustSortJSON(out)))

	return err
}

// InitCmd returns a command that initializes all files needed for Tendermint
// and the respective application.
func InitCmd(mbm module.BasicManager, defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [moniker]",
		Short: "Initialize private validator, p2p, genesis, and application configuration files",
		Long:  `Initialize validators's and node's configuration files.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			// This is a slice of SEED nodes, not peers.  They must be configured in seed mode.
			// An easy way to run a lightweight seed node is to use tenderseed: github.com/binaryholdings/tenderseed
			// Another easy way to run a seed node is tinyseed: https://github.com/notional-labs/tinyseed
			// Tinyseed is made for Akash!
			seeds := []string{
				"21d7539792ee2e0d650b199bf742c56ae0cf499e@162.55.132.230:2000",               // Notional
				"295b417f995073d09ff4c6c141bd138a7f7b5922@65.21.141.212:2000",                // Notional
				"ec4d3571bf709ab78df61716e47b5ac03d077a1a@65.108.43.26:2000",                 // Notional
				"4cb8e1e089bdf44741b32638591944dc15b7cce3@65.108.73.18:2000",                 // Notional
				"f515a8599b40f0e84dfad935ba414674ab11a668@osmosis.blockpane.com:26656",       // [ block pane ]
				"6bcdbcfd5d2c6ba58460f10dbcfde58278212833@osmosis.artifact-staking.io:26656", // Artifact Staking
			}

			// Override default settings in config.toml
			config.P2P.Seeds = strings.Join(seeds[:], ",")
			config.P2P.MaxNumInboundPeers = 320
			config.P2P.MaxNumOutboundPeers = 40
			config.Mempool.Size = 10000
			config.StateSync.TrustPeriod = 112 * time.Hour
			config.FastSync.Version = "v0"

			config.SetRoot(clientCtx.HomeDir)

			chainID, _ := cmd.Flags().GetString(flags.FlagChainID)
			if chainID == "" {
				chainID = fmt.Sprintf("test-chain-%v", tmrand.Str(6))
			}

			// Get bip39 mnemonic
			var mnemonic string
			recover, _ := cmd.Flags().GetBool(FlagRecover)
			if recover {
				inBuf := bufio.NewReader(cmd.InOrStdin())
				mnemonic, err := input.GetString("Enter your bip39 mnemonic", inBuf)
				if err != nil {
					return err
				}

				if !bip39.IsMnemonicValid(mnemonic) {
					return errors.New("invalid mnemonic")
				}
			}

			nodeID, _, err := genutil.InitializeNodeValidatorFilesFromMnemonic(config, mnemonic)
			if err != nil {
				return err
			}

			config.Moniker = args[0]

			config.Genesis = "genesis.json"

			// this can be moved to another file for when we want to play with empty chains and the like.
			/*
				overwrite, _ := cmd.Flags().GetBool(FlagOverwrite)

				if !overwrite && tmos.FileExists(genFile) {
					return fmt.Errorf("genesis.json file already exists: %v", genFile)
				}
				appState, err := json.MarshalIndent(mbm.DefaultGenesis(cdc), "", " ")
				if err != nil {
					return errors.Wrap(err, "Failed to marshall default genesis state")
				}

				genDoc := &types.GenesisDoc{}
				if _, err := os.Stat(genFile); err != nil {
					if !os.IsNotExist(err) {
						return err
					}
				} else {
					genDoc, err = types.GenesisDocFromFile(genFile)
					if err != nil {
						return errors.Wrap(err, "Failed to read genesis doc from file")
					}
				}

				genDoc.ChainID = chainID
				genDoc.Validators = nil
				genDoc.AppState = appState
				if err = genutil.ExportGenesisFile(genDoc, genFile); err != nil {
					return errors.Wrap(err, "Failed to export genesis file")
				}
			*/

			toPrint := newPrintInfo(config.Moniker, chainID, nodeID, "")

			tmcfg.WriteConfigFile(filepath.Join(config.RootDir, "config", "config.toml"), config)

			return displayInfo(toPrint)
		},
	}

	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")
	cmd.Flags().BoolP(FlagOverwrite, "o", false, "overwrite the genesis.json file")
	cmd.Flags().Bool(FlagRecover, false, "provide seed phrase to recover existing key instead of creating")
	cmd.Flags().String(flags.FlagChainID, "", "genesis file chain-id, if left blank will be randomly created")

	return cmd
}
