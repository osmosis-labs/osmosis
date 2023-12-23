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

	tmcfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/cli"
	tmos "github.com/cometbft/cometbft/libs/os"
	tmrand "github.com/cometbft/cometbft/libs/rand"
	"github.com/cometbft/cometbft/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"

	"github.com/osmosis-labs/osmosis/v21/app"
)

const (
	// FlagOverwrite defines a flag to overwrite an existing genesis JSON file.
	FlagOverwrite = "overwrite"

	// FlagSeed defines a flag to initialize the private validator key from a specific seed.
	FlagRecover = "recover"

	// FlagSetEnv defines a flag to create environment file & save current home directory into it.
	FlagSetEnv = "set-env"
)

type printInfo struct {
	Moniker    string          `json:"moniker" yaml:"moniker"`
	ChainID    string          `json:"chain_id" yaml:"chain_id"`
	NodeID     string          `json:"node_id" yaml:"node_id"`
	GenTxsDir  string          `json:"gentxs_dir" yaml:"gentxs_dir"`
	AppMessage json.RawMessage `json:"app_message" yaml:"app_message"`
}

// newPrintInfo initializes a printInfo struct.
func newPrintInfo(moniker, chainID, nodeID, genTxsDir string, appMessage json.RawMessage) printInfo {
	return printInfo{
		Moniker:    moniker,
		ChainID:    chainID,
		NodeID:     nodeID,
		GenTxsDir:  genTxsDir,
		AppMessage: appMessage,
	}
}

// displayInfo displays printInfo in JSON format.
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
			cdc := clientCtx.Codec

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			// An easy way to run a lightweight seed node is to use tenderseed: github.com/binaryholdings/tenderseed

			seeds := []string{
				"21d7539792ee2e0d650b199bf742c56ae0cf499e@162.55.132.230:2000",                             // Notional
				"44ff091135ef2c69421eacfa136860472ac26e60@65.21.141.212:2000",                              // Notional
				"ec4d3571bf709ab78df61716e47b5ac03d077a1a@65.108.43.26:2000",                               // Notional
				"4cb8e1e089bdf44741b32638591944dc15b7cce3@65.108.73.18:2000",                               // Notional
				"f515a8599b40f0e84dfad935ba414674ab11a668@osmosis.blockpane.com:26656",                     // [ block pane ]
				"6bcdbcfd5d2c6ba58460f10dbcfde58278212833@osmosis.artifact-staking.io:26656",               // Artifact Staking
				"24841abfc8fbd401d8c86747eec375649a2e8a7e@osmosis.pbcups.org:26656",                        // Pbcups
				"77bb5fb9b6964d6e861e91c1d55cf82b67d838b5@bd-osmosis-seed-mainnet-us-01.bdnodes.net:26656", // Blockdaemon US
				"3243426ab56b67f794fa60a79cc7f11bc7aa752d@bd-osmosis-seed-mainnet-eu-02.bdnodes.net:26656", // Blockdaemon EU
				"ebc272824924ea1a27ea3183dd0b9ba713494f83@osmosis-mainnet-seed.autostake.com:26716",        // AutoStake.com
				"7c66126b64cd66bafd9ccfc721f068df451d31a3@osmosis-seed.sunshinevalidation.io:9393",         // Sunshine Validation
			}
			config.P2P.Seeds = strings.Join(seeds, ",")
			config.P2P.MaxNumInboundPeers = 80
			config.P2P.MaxNumOutboundPeers = 60
			config.Mempool.Size = 10000
			config.StateSync.TrustPeriod = 112 * time.Hour

			// The original default is 5s and is set in Cosmos SDK.
			// We lower it to 4s for faster block times.
			config.Consensus.TimeoutCommit = 4 * time.Second

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
				return errors.Wrap(err, "Failed to export genesis file")
			}

			toPrint := newPrintInfo(config.Moniker, chainID, nodeID, "", appState)

			tmcfg.WriteConfigFile(filepath.Join(config.RootDir, "config", "config.toml"), config)

			err = writeConfigToFile(filepath.Join(config.RootDir, "config", "client.toml"), nil)
			if err != nil {
				return errors.Wrap(err, "Failed to write client.toml file")
			}

			createEnv, _ := cmd.Flags().GetBool(FlagSetEnv)
			if createEnv {
				err = CreateEnvFile(cmd)
				if err != nil {
					return errors.Wrapf(err, "Failed to create environment file")
				}
			}
			return displayInfo(toPrint)
		},
	}

	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")
	cmd.Flags().BoolP(FlagOverwrite, "o", false, "overwrite the genesis.json file")
	cmd.Flags().Bool(FlagRecover, false, "provide seed phrase to recover existing key instead of creating")
	cmd.Flags().String(flags.FlagChainID, "", "genesis file chain-id, if left blank will be randomly created")
	cmd.Flags().Bool(FlagSetEnv, false, "set and save current directory as home directory")

	return cmd
}

func CreateEnvFile(cmd *cobra.Command) error {
	// Check if .env file was created in /.osmosisd
	envPath := filepath.Join(app.DefaultNodeHome, ".env")
	if _, err := os.Stat(envPath); err != nil {
		// If not exist, we create a new .env file with node dir passed
		if os.IsNotExist(err) {
			// Create ./osmosisd if not exist
			if _, err = os.Stat(app.DefaultNodeHome); err != nil {
				if os.IsNotExist(err) {
					err = os.MkdirAll(app.DefaultNodeHome, 0777)
					if err != nil {
						return err
					}
				}
			}

			// Create environment file
			envFile, err := os.Create(envPath)
			if err != nil {
				return err
			}

			// In case the user wants to init in a specific dir, save it to .env
			nodeHome, err := cmd.Flags().GetString(cli.HomeFlag)
			if err != nil {
				fmt.Println("using mainnet environment")
				nodeHome = EnvMainnet
			}
			_, err = envFile.WriteString(fmt.Sprintf("OSMOSISD_ENVIRONMENT=%s", nodeHome))
			if err != nil {
				return err
			}
		}
	}
	return nil
}
