package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/v19/app"
)

const (
	EnvVariable = "OSMOSISD_ENVIRONMENT"
	EnvMainnet  = "mainnet"
	EnvTestnet  = "testnet"
	EnvLocalnet = "localnet"
)

// ExportAirdropSnapshotCmd generates a snapshot.json from a provided exported genesis.json.
func ChangeEnvironmentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-env [new env]",
		Short: "Set home environment variables for commands",
		Long: `Set home environment variables for commands
Example:
	osmosisd set-env mainnet
	osmosisd set-env testnet
	osmosisd set-env localnet [optional-chain-id]
	osmosisd set-env $HOME/.custom-dir
`,
		Args: customArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Note: If we are calling this method, the environment file has already been set in
			// NewRootCmd() when creating the rootCmd. We do this because order of operations
			// dictates this as a requirement. If we changed the env file here, the osmosis
			// daemon would not initialize the folder we are intending to set to.
			newEnv := args[0]
			chainId := ""
			if len(args) > 1 {
				chainId = args[1]
			}
			return clientSettingsFromEnv(cmd, newEnv, chainId)
		},
	}
	return cmd
}

// PrintEnvironmentCmd prints the current environment.
func PrintEnvironmentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-env",
		Short: "Prints the current environment",
		Long: `Prints the current environment
Example:
	osmosisd get-env'

	Returns one of:
	- mainnet implying $HOME/.osmosisd
	- testnet implying $HOME/.osmosisd-test
	- localosmosis implying $HOME/.osmosisd-local
	- custom path`,
		RunE: func(cmd *cobra.Command, args []string) error {
			environment := getHomeEnvironment()
			path, err := environmentNameToPath(environment)
			if err != nil {
				return err
			}

			fmt.Println("Environment name: ", environment)
			fmt.Println("Environment path: ", path)
			return nil
		},
	}
	return cmd
}

func environmentNameToPath(environmentName string) (string, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch environmentName {
	case EnvMainnet:
		return app.DefaultNodeHome, nil
	case EnvTestnet:
		return filepath.Join(userHomeDir, ".osmosisd-test/"), nil
	case EnvLocalnet:
		return filepath.Join(userHomeDir, ".osmosisd-local/"), nil
	default:
		_, err := os.Stat(environmentName)
		if os.IsNotExist(err) {
			// Creating new environment directory
			if err := os.Mkdir(environmentName, os.ModePerm); err != nil {
				return "", err
			}
		}
		return environmentName, nil
	}
}

// clientSettingsFromEnv takes the env name (mainnet, testnet, localnet, etc) and sets the
// client.toml settings to commonly used values for that environment.
func clientSettingsFromEnv(cmd *cobra.Command, environmentName, chainId string) error {
	envConfigs := map[string]map[string]string{
		EnvMainnet: {
			flags.FlagChainID:       "osmosis-1",
			flags.FlagNode:          "https://rpc.osmosis.zone:443",
			flags.FlagBroadcastMode: "block",
		},
		EnvTestnet: {
			flags.FlagChainID:       "osmo-test-5",
			flags.FlagNode:          "https://rpc.testnet.osmosis.zone:443",
			flags.FlagBroadcastMode: "block",
		},
		EnvLocalnet: {
			flags.FlagChainID:       "localosmosis",
			flags.FlagBroadcastMode: "block",
		},
	}

	configs, ok := envConfigs[environmentName]
	if !ok {
		return nil
	}

	// Update the ChainID if environmentName is EnvLocalnet and chainId is provided
	if environmentName == EnvLocalnet && chainId != "" {
		configs[flags.FlagChainID] = chainId
	}

	for flag, value := range configs {
		if err := runConfigCmd(cmd, []string{flag, value}); err != nil {
			return err
		}
	}
	return nil
}

// changeEnvironment takes the given environment name and changes the .env file to reflect it.
func changeEnvironment(args []string) error {
	newEnv := args[0]

	currentEnvironment := getHomeEnvironment()
	fmt.Println("Current environment: ", currentEnvironment)

	if _, err := environmentNameToPath(newEnv); err != nil {
		return err
	}

	fmt.Println("New environment: ", newEnv)

	envMap := make(map[string]string)
	envMap[EnvVariable] = newEnv
	err := godotenv.Write(envMap, filepath.Join(app.DefaultNodeHome, ".env"))
	if err != nil {
		return err
	}

	return nil
}

// createHomeDirIfNotExist creates the home directory if it does not exist and writes a blank
// .env file. This is used for the first time setup of the osmosisd home directory.
func createHomeDirIfNotExist(homeDir string) error {
	if _, err := os.Stat(homeDir); os.IsNotExist(err) {
		err := os.MkdirAll(homeDir, 0755)
		if err != nil {
			return err
		}
	}

	envFilePath := filepath.Join(homeDir, ".env")
	if _, err := os.Stat(envFilePath); os.IsNotExist(err) {
		file, err := os.Create(envFilePath)
		if err != nil {
			return err
		}
		file.Close()
	}

	return nil
}

// changeEnvPriorToSetup changes the env file to reflect the desired environment the user wants to change to.
// If this is not called in NewRootCmd(), the environment change will happen **after** all relevant setup actions
// happen (e.g., the .env will be read in as the previous value in the setup and not the new value).
func changeEnvPriorToSetup(cmd *cobra.Command, initClientCtx *client.Context, args []string, homeDir string) error {
	if cmd.Name() == "set-env" {
		err := createHomeDirIfNotExist(homeDir)
		if err != nil {
			return err
		}

		err = changeEnvironment(args)
		if err != nil {
			return err
		}

		homeEnvironment := getHomeEnvironment()
		homeDir, err := environmentNameToPath(homeEnvironment)
		if err != nil {
			// Failed to convert home environment to home path, using default home
			homeDir = app.DefaultNodeHome
		}
		*initClientCtx = initClientCtx.WithHomeDir(homeDir)
	}
	return nil
}

// customArgs accepts one arg, but if the first arg is "localnet", then it accepts two args.
func customArgs(cmd *cobra.Command, args []string) error {
	if len(args) < 1 || len(args) > 2 {
		return errors.New("set-env requires 1 or 2 arguments")
	}
	if args[0] == "localnet" {
		return nil
	}
	if len(args) == 2 {
		return errors.New("only 'set-env localnet' accepts a second argument")
	}
	return nil
}
