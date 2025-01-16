package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	cmttypes "github.com/cometbft/cometbft/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	"github.com/osmosis-labs/osmosis/v29/app"
)

const (
	// FlagOverwrite defines a flag to overwrite an existing genesis JSON file.
	FlagOverwrite = "overwrite"

	// FlagSeed defines a flag to initialize the private validator key from a specific seed.
	FlagRecover = "recover"

	// FlagSetEnv defines a flag to create environment file & save current home directory into it.
	FlagSetEnv = "set-env"

	// FlagRejectConfigDefaults defines a flag to reject some select defaults that override what is in the config file.
	FlagRejectConfigDefaults = "reject-config-defaults"
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

			// Add Osmosis specific defaults to config.toml

			// P2P
			seeds := []string{
				"21d7539792ee2e0d650b199bf742c56ae0cf499e@162.55.132.230:2000",                             // Notional
				"44ff091135ef2c69421eacfa136860472ac26e60@65.21.141.212:2000",                              // Notional
				"ec4d3571bf709ab78df61716e47b5ac03d077a1a@65.108.43.26:2000",                               // Notional
				"4cb8e1e089bdf44741b32638591944dc15b7cce3@65.108.73.18:2000",                               // Notional
				"f515a8599b40f0e84dfad935ba414674ab11a668@osmosis.blockpane.com:26656",                     // [ block pane ]
				"6bcdbcfd5d2c6ba58460f10dbcfde58278212833@osmosis.artifact-staking.io:26656",               // Artifact Staking
				"77bb5fb9b6964d6e861e91c1d55cf82b67d838b5@bd-osmosis-seed-mainnet-us-01.bdnodes.net:26656", // Blockdaemon US
				"3243426ab56b67f794fa60a79cc7f11bc7aa752d@bd-osmosis-seed-mainnet-eu-02.bdnodes.net:26656", // Blockdaemon EU
				"ebc272824924ea1a27ea3183dd0b9ba713494f83@osmosis-mainnet-seed.autostake.com:26716",        // AutoStake.com
				"7c66126b64cd66bafd9ccfc721f068df451d31a3@osmosis-seed.sunshinevalidation.io:9393",         // Sunshine Validation
			}
			config.P2P.Seeds = strings.Join(seeds, ",")
			config.P2P.MaxNumInboundPeers = 80
			config.P2P.MaxNumOutboundPeers = 60
			// config.P2P.FlushThrottleTimeout = 10 * time.Millisecond

			// Mempool
			config.Mempool.Size = 10000

			// State Sync
			config.StateSync.TrustPeriod = 112 * time.Hour

			// Consensus
			config.Consensus.TimeoutCommit = 1500 * time.Millisecond // 1.5s
			// config.Consensus.PeerGossipSleepDuration = 10 * time.Millisecond

			// Other
			config.Moniker = args[0]
			config.SetRoot(clientCtx.HomeDir)

			// Get bip39 mnemonic and initialize node validator files
			var mnemonic string
			var err error
			recover, _ := cmd.Flags().GetBool(FlagRecover)
			if recover {
				inBuf := bufio.NewReader(cmd.InOrStdin())
				mnemonic, err = input.GetString("Enter your bip39 mnemonic", inBuf)
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

			genFilePath := config.GenesisFile()
			chainID, _ := cmd.Flags().GetString(flags.FlagChainID)
			overwrite, _ := cmd.Flags().GetBool(FlagOverwrite)

			if !overwrite && tmos.FileExists(genFilePath) {
				return fmt.Errorf("genesis.json file already exists: %v", genFilePath)
			}

			var toPrint printInfo
			isMainnet := chainID == "" || chainID == "osmosis-1"
			genesisFileDownloadFailed := false

			if isMainnet {
				// If the chainID is blank or osmosis-1, prep this as a mainnet node

				// Attempt to download the genesis file from the Osmosis GitHub repository
				// If fail, generate a new genesis file
				err := downloadGenesis(config)
				if err != nil {
					// TODO: Maybe we should just fail in this case?
					fmt.Println("Failed to download genesis file, using a random chain ID and genesis file for local testing")
					genesisFileDownloadFailed = true
					chainID = fmt.Sprintf("test-chain-%v", tmrand.Str(6))
				} else {
					// Set chainID to osmosis-1 in the case of a blank chainID
					chainID = "osmosis-1"

					// We dont print the app state for mainnet nodes because it's massive
					fmt.Println("Not printing app state for mainnet node due to verbosity")
					toPrint = newPrintInfo(config.Moniker, chainID, nodeID, "", nil)
				}
			}

			// If this is not a mainnet node, or the genesis file download failed, generate a new genesis file
			if genesisFileDownloadFailed || !isMainnet {
				// If the chainID is not blank or genesis file download failed, generate a new genesis file
				var genDoc genutiltypes.AppGenesis

				appState, err := json.MarshalIndent(mbm.DefaultGenesis(cdc), "", " ")
				if err != nil {
					return errors.Wrap(err, "Failed to marshall default genesis state")
				}

				if _, err := os.Stat(genFilePath); err != nil {
					if !os.IsNotExist(err) {
						return err
					}
				} else {
					_, genDocFromFile, err := genutiltypes.GenesisStateFromGenFile(genFilePath)
					if err != nil {
						return fmt.Errorf("failed to unmarshal genesis state: %w", err)
					}
					genDoc = *genDocFromFile
				}

				genDoc.Consensus = &genutiltypes.ConsensusGenesis{}
				genDoc.Consensus.Params = cmttypes.DefaultConsensusParams()

				genDoc.ChainID = chainID

				genDoc.Consensus.Validators = nil
				genDoc.AppState = appState
				if err = genutil.ExportGenesisFile(&genDoc, genFilePath); err != nil {
					return errors.Wrap(err, "Failed to export genesis file")
				}

				toPrint = newPrintInfo(config.Moniker, chainID, nodeID, "", genDoc.AppState)
			}

			// Write both app.toml and config.toml to the app's config directory
			tmcfg.WriteConfigFile(filepath.Join(config.RootDir, "config", "config.toml"), config)
			err = writeConfigToFile(filepath.Join(config.RootDir, "config", "client.toml"), nil)
			if err != nil {
				return errors.Wrap(err, "Failed to write client.toml file")
			}

			// Create .env file if FlagSetEnv is true
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

// downloadGenesis downloads the genesis file from a predefined URL and writes it to the genesis file path specified in the config.
// It creates an HTTP client to send a GET request to the genesis file URL. If the request is successful, it reads the response body
// and writes it to the destination genesis file path. If any step in this process fails, it generates the default genesis.
//
// Parameters:
// - config: A pointer to a tmcfg.Config object that contains the configuration, including the genesis file path.
//
// Returns:
// - An error if the download or file writing fails, otherwise nil.
func downloadGenesis(config *tmcfg.Config) error {
	// URL of the genesis file to download
	genesisURL := "https://github.com/osmosis-labs/osmosis/raw/main/networks/osmosis-1/genesis.json?download"

	// Determine the destination path for the genesis file
	genFilePath := config.GenesisFile()

	// Create a new HTTP client with a 30-second timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create a new GET request
	req, err := http.NewRequest("GET", genesisURL, nil)
	if err != nil {
		return errors.Wrap(err, "failed to create HTTP request for genesis file")
	}

	// Send the request
	fmt.Println("Downloading genesis file from", genesisURL)
	fmt.Println("If the download is not successful in 30 seconds, we will gracefully continue and the default genesis file will be used")
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to download genesis file")
	}
	defer resp.Body.Close()

	// Check if the HTTP request was successful
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("failed to download genesis file: HTTP status %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "failed to read genesis file response body")
	}

	// Write the body to the destination genesis file
	err = os.WriteFile(genFilePath, body, 0644)
	if err != nil {
		return errors.Wrap(err, "failed to write genesis file to destination")
	}

	return nil
}
