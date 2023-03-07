package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/v15/app"
)

const (
	EnvVariable = "OSMOSISD_ENVIRONMENT"
	EnvMainnet  = "mainnet"
	EnvLocalnet = "localosmosis"
)

// ExportAirdropSnapshotCmd generates a snapshot.json from a provided exported genesis.json.
func ChangeEnvironmentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-env [new env]",
		Short: "Set home environment variables for commands",
		Long: `Set home environment variables for commands
Example:
	osmosisd set-env mainnet
	osmosisd set-env localosmosis
	osmosisd set-env $HOME/.custom-dir
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			newEnv := args[0]

			currentEnvironment := getHomeEnvironment()
			fmt.Println("Current environment: ", currentEnvironment)

			path_newEnv, err := environmentNameToPath(newEnv)
			if err != nil {
				return err
			}

			fmt.Println("New environment: ", newEnv)

			envMap, err := godotenv.Read(filepath.Join(app.DefaultNodeHome, ".env"))
			if err != nil {
				return err
			}

			envMap[EnvVariable] = newEnv
			envMap[newEnv] = path_newEnv
			err = godotenv.Write(envMap, filepath.Join(app.DefaultNodeHome, ".env"))
			if err != nil {
				return err
			}
			return nil
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
	- localosmosis implying $HOME/.osmosisd-local
	- localosmosis
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
	case EnvLocalnet:
		return filepath.Join(userHomeDir, ".osmosisd-local/"), nil
	default:
		osmosisdPathCustom := filepath.Join(userHomeDir, ".osmosisd-custom")
		if _, err = os.Stat(osmosisdPathCustom); err != nil {
			// Creating new $HOME/.osmosisd-custom
			if err = os.Mkdir(osmosisdPathCustom, os.ModePerm); err != nil {
				return "", err
			}
		}
		osmosisdPath := filepath.Join(osmosisdPathCustom, environmentName)

		_, err = os.Stat(osmosisdPath)
		if os.IsNotExist(err) {
			// Creating new environment directory
			if err := os.Mkdir(osmosisdPath, os.ModePerm); err != nil {
				return "", err
			}
		}
		return osmosisdPath, nil
	}
}
func PrintAllEnvironmentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-env ",
		Short: "listing all available environments.",
		Long: `listing all available environments.
Example:
	osmosisd list-env'

	Returns all EnvironmentCmd`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Mainnet
			path, err := environmentNameToPath(EnvMainnet)
			if err != nil {
				fmt.Printf("%s \n", err.Error())
			} else {
				fmt.Printf("Environment name: %s \n", EnvMainnet)
				fmt.Printf("Environment path: %s \n\n", path)
			}
			// Localnet
			path, err = environmentNameToPath(EnvLocalnet)
			if err != nil {
				fmt.Printf("%s \n", err.Error())
			} else {
				fmt.Printf("Environment name: %s \n", EnvLocalnet)
				fmt.Printf("Environment path: %s \n\n", path)
			}
			// Custom
			envMap, err := godotenv.Read(filepath.Join(app.DefaultNodeHome, ".env"))
			if err != nil {
				return err
			}
			for name, path := range envMap {
				if name == EnvVariable {
					continue
				}
				fmt.Printf("Environment name: %s \n", name)
				fmt.Printf("Environment path: %s \n\n", path)
			}

			return nil
		},
	}
	return cmd
}
