package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cosmos/go-bip39"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	tmcfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/cli"
	tmos "github.com/tendermint/tendermint/libs/os"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	"github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/server"
	appcfg "github.com/cosmos/cosmos-sdk/server/config"
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

func newPrintInfo(moniker, chainID, nodeID, genTxsDir string, appMessage json.RawMessage) printInfo {
	return printInfo{
		Moniker:    moniker,
		ChainID:    chainID,
		NodeID:     nodeID,
		GenTxsDir:  genTxsDir,
		AppMessage: appMessage,
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
			cdc := clientCtx.JSONMarshaler

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			config.P2P.Seeds = "085f62d67bbf9c501e8ac84d4533440a1eef6c45@95.217.196.54:26656"
			config.P2P.PersistentPeers = "2e3e3b7703a598024a2fb287587095bc4d14fe52@95.217.196.54:2000,f5be19f84deb843c18e9b612b7987138ba13ac02@5.9.106.185:2000,f9c49739f0641a0a673e7a1e8edc38054fefc840@144.76.183.180:2000,40aafcd9b6959d58dd1c567d9daf2a82a23311cf@162.55.132.230:2000,3fea02d121cb24503d5fbc53216a527257a9ab55@143.198.145.208:26656,e7916387e05acd53d1b8c0f842c13def365c7bb6@176.9.64.212:26656,785bc83577e3980545bac051de8f57a9fd82695f@194.233.164.146:26656,421e7ffd6c239cc51620e2ec2330d392a7baeeb3@207.246.109.0:26656,778fdedf6effe996f039f22901a3360bc838b52e@161.97.187.189:36656,2115945f074ddb038de5d835e287fa03e32f0628@95.217.43.85:26656,a79b9904461a98cdb98a5c15d7d75b02a55b6ab9@162.216.4.214:36656,1639d885f47c96c5cc4be6e4443ba53cdc74fc71@23.111.129.54:36656"
			config.P2P.MaxNumOutboundPeers = 40
			config.P2P.MaxNumInboundPeers = 150
			config.Mempool.Size = 10000
			config.StateSync.TrustPeriod = 112 * time.Hour
			config.FastSync.Version = "v0"

			config.SetRoot(clientCtx.HomeDir)

			appConfig := appcfg.DefaultConfig()
			appConfig.API.Enable = true
			appConfig.StateSync.SnapshotInterval = 1500
			appConfig.StateSync.SnapshotKeepRecent = 2

			chainID, _ := cmd.Flags().GetString(flags.FlagChainID)
			if chainID == "" {
				chainID = fmt.Sprintf("test-chain-%v", tmrand.Str(6))
			}

			fmt.Println("help")

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

			genFile := config.GenesisFile()
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
				return errors.Wrap(err, "Failed to export gensis file")
			}

			toPrint := newPrintInfo(config.Moniker, chainID, nodeID, "", appState)

			tmcfg.WriteConfigFile(filepath.Join(config.RootDir, "config", "config.toml"), config)
			appcfg.WriteConfigFile(filepath.Join(config.RootDir, "config", "app.toml"), appConfig)

			return displayInfo(toPrint)
		},
	}

	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")
	cmd.Flags().BoolP(FlagOverwrite, "o", false, "overwrite the genesis.json file")
	cmd.Flags().Bool(FlagRecover, false, "provide seed phrase to recover existing key instead of creating")
	cmd.Flags().String(flags.FlagChainID, "", "genesis file chain-id, if left blank will be randomly created")

	return cmd
}
