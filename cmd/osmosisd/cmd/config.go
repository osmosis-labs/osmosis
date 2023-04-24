package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	clientconfig "github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/flags"
	viper "github.com/spf13/viper"
)

type OsmosisCustomClient struct {
	clientconfig.ClientConfig
	Gas           string `mapstructure:"gas" json:"gas"`
	GasPrices     string `mapstructure:"gas-prices" json:"gas-prices"`
	GasAdjustment string `mapstructure:"gas-adjustment" json:"gas-adjustment"`
}

// ConfigCmd returns a CLI command to interactively create an application CLI
// config file.
func ConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config <key> [value]",
		Short: "Create or query an application CLI configuration file",
		RunE:  runConfigCmd,
		Args:  cobra.RangeArgs(0, 2),
	}
	return cmd
}

func runConfigCmd(cmd *cobra.Command, args []string) error {
	clientCtx := client.GetClientContextFromCmd(cmd)
	configPath := filepath.Join(clientCtx.HomeDir, "config")

	conf, err := getClientConfig(configPath, clientCtx.Viper)
	if err != nil {
		return fmt.Errorf("couldn't get client config: %v", err)
	}

	jcc := OsmosisCustomClient{
		*conf,
		os.Getenv("JUNOD_GAS"),
		os.Getenv("JUNOD_GAS_PRICES"),
		os.Getenv("JUNOD_GAS_ADJUSTMENT"),
	}

	switch len(args) {
	case 0:
		s, err := json.MarshalIndent(jcc, "", "\t")
		if err != nil {
			return err
		}

		cmd.Println(string(s))

	case 1:
		// it's a get
		key := args[0]

		switch key {
		case flags.FlagChainID:
			cmd.Println(conf.ChainID)
		case flags.FlagKeyringBackend:
			cmd.Println(conf.KeyringBackend)
		case tmcli.OutputFlag:
			cmd.Println(conf.Output)
		case flags.FlagNode:
			cmd.Println(conf.Node)
		case flags.FlagBroadcastMode:
			cmd.Println(conf.BroadcastMode)

		// Custom flags
		case flags.FlagGas:
			cmd.Println(jcc.Gas)
		case flags.FlagGasPrices:
			cmd.Println(jcc.GasPrices)
		case flags.FlagGasAdjustment:
			cmd.Println(jcc.GasAdjustment)
		default:
			err := errUnknownConfigKey(key)
			return fmt.Errorf("couldn't get the value for the key: %v, error:  %v", key, err)
		}

	case 2:
		// it's set
		key, value := args[0], args[1]

		switch key {
		case flags.FlagChainID:
			jcc.ChainID = value
		case flags.FlagKeyringBackend:
			jcc.KeyringBackend = value
		case tmcli.OutputFlag:
			jcc.Output = value
		case flags.FlagNode:
			jcc.Node = value
		case flags.FlagBroadcastMode:
			jcc.BroadcastMode = value
		case flags.FlagGas:
			jcc.Gas = value
		case flags.FlagGasPrices:
			jcc.GasPrices = value
			jcc.Fees = "" // resets since we can only use 1 at a time
		case flags.FlagGasAdjustment:
			jcc.GasAdjustment = value
		case flags.FlagFees:
			jcc.Fees = value
			jcc.GasPrices = "" // resets since we can only use 1 at a time
		case flags.FlagFeeAccount:
			jcc.FeeAccount = value
		case flags.FlagNote:
			jcc.Note = value
		default:
			return errUnknownConfigKey(key)
		}

		confFile := filepath.Join(configPath, "client.toml")
		if err := writeConfigToFile(confFile, &jcc); err != nil {
			return fmt.Errorf("could not write client config to the file: %v", err)
		}

	default:
		panic("cound not execute config command")
	}

	return nil
}

// getClientConfig reads values from client.toml file and unmarshalls them into ClientConfig
func getClientConfig(configPath string, v *viper.Viper) (*scconfig.ClientConfig, error) {
	v.AddConfigPath(configPath)
	v.SetConfigName("client")
	v.SetConfigType("toml")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	conf := new(clientconfig.ClientConfig)
	if err := v.Unmarshal(conf); err != nil {
		return nil, err
	}

	return conf, nil
}

func errUnknownConfigKey(key string) error {
	return fmt.Errorf("unknown configuration key: %q", key)
}