package testnetify

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/osmosis-labs/osmosis/v8/app"
	"github.com/spf13/cobra"
)

// get cmd to convert any bech32 address to an osmo prefix.
func StateExportToTestnetGenesis() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "testnetify [state_export_path] -p [testnet_params]",
		Short: "Convert state export to be a testnet",
		Long: `Fill this out plz
	`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			testnetParams, err := loadTestnetParams(cmd)
			if err != nil {
				return err
			}

			state_export_path := args[0]
			file, err := ioutil.ReadFile(state_export_path)
			if err != nil {
				// failed to read file
				return err
			}
			var genesis app.GenesisState
			err = json.Unmarshal(file, &genesis)
			if err != nil {
				return err
			}

			replaceValidatorDetails(genesis, testnetParams)
			updateChainId(genesis)
			clearIBC(genesis)

			cmd.Println("Writing new genesis to: ", testnetParams.OutputExportFilepath)
			err = writeGenesis(genesis, testnetParams.OutputExportFilepath)
			if err != nil {
				// failed to read file
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringP(flagTestnetParams, "p", "params", "Testnet params json")

	return cmd
}

func writeGenesis(genesis app.GenesisState, path string) error {
	file, err := os.OpenFile(path, os.O_CREATE, os.ModePerm)
	if err != nil {
		// failed to read file
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	return encoder.Encode(genesis)
}
