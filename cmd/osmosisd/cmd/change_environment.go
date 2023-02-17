package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/v14/app"
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
			if _, err := os.Stat(newEnv); os.IsNotExist(err) {
				return fmt.Errorf("directory %s does not exist", newEnv)
			}

			userHomeDir, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			envPath := filepath.Join(userHomeDir, ".osmosisd/.env")

			err = godotenv.Load(envPath)
			if err != nil {
				if err = CreateEnvFile(cmd); err != nil {
					return err
				}
			}
			m := make(map[string]string)

			if newEnv == EnvMainnet {
				newEnv = app.DefaultNodeHome
			} else if newEnv == EnvLocalnet {
				newEnv = filepath.Join(userHomeDir, ".osmosisd-local")
			}
			m[EnvVariable] = newEnv

			err = godotenv.Write(m, envPath)
			if err != nil {
				return err
			}

			return nil
		},
	}
	return cmd
}
