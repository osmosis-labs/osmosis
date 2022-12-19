package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
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
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			newEnv := args[0]
			if newEnv != EnvMainnet && newEnv != EnvLocalnet {
				return fmt.Errorf("invalid environment variable")
			}

			userHomeDir, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			envPath := filepath.Join(userHomeDir, ".osmosisd/.env")

			err = godotenv.Load(envPath)
			if err != nil {
				return err
			}

			m := make(map[string]string)
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
