package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	viper "github.com/spf13/viper"
	tmcli "github.com/tendermint/tendermint/libs/cli"
)

type OsmosisCustomClient struct {
	ChainID        string `mapstructure:"chain-id" json:"chain-id"`
	KeyringBackend string `mapstructure:"keyring-backend" json:"keyring-backend"`
	Output         string `mapstructure:"output" json:"output"`
	Node           string `mapstructure:"node" json:"node"`
	BroadcastMode  string `mapstructure:"broadcast-mode" json:"broadcast-mode"`
	Gas            string `mapstructure:"gas" json:"gas"`
	GasPrices      string `mapstructure:"gas-prices" json:"gas-prices"`
	GasAdjustment  string `mapstructure:"gas-adjustment" json:"gas-adjustment"`
	Fees           string `mapstructure:"fees" json:"fees"`
}

// Override sdk ConfigCmd func
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

	switch len(args) {
	case 0:
		s, err := json.MarshalIndent(conf, "", "\t")
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
		case flags.FlagGas:
			cmd.Println(conf.Gas)
		case flags.FlagGasPrices:
			cmd.Println(conf.GasPrices)
		case flags.FlagGasAdjustment:
			cmd.Println(conf.GasAdjustment)
		case flags.FlagFees:
			cmd.Println(conf.Fees)
		default:
			err := errUnknownConfigKey(key)
			return fmt.Errorf("couldn't get the value for the key: %v, error:  %v", key, err)
		}

	case 2:
		// it's set
		key, value := args[0], args[1]

		switch key {
		case flags.FlagChainID:
			conf.ChainID = value
		case flags.FlagKeyringBackend:
			conf.KeyringBackend = value
		case tmcli.OutputFlag:
			conf.Output = value
		case flags.FlagNode:
			conf.Node = value
		case flags.FlagBroadcastMode:
			conf.BroadcastMode = value
		case flags.FlagGas:
			conf.Gas = value
		case flags.FlagGasPrices:
			conf.GasPrices = value
		case flags.FlagGasAdjustment:
			conf.GasAdjustment = value
		case flags.FlagFees:
			conf.Fees = value
		default:
			return errUnknownConfigKey(key)
		}

		confFile := filepath.Join(configPath, "client.toml")
		if err := writeConfigToFile(confFile, conf); err != nil {
			return fmt.Errorf("could not write client config to the file: %v", err)
		}

	default:
		panic("cound not execute config command")
	}

	return nil
}

const defaultConfigTemplate = `# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

###############################################################################
###                           Client Configuration                            ###
###############################################################################

# The network chain ID
chain-id = "{{ .ChainID }}"
# The keyring's backend, where the keys are stored (os|file|kwallet|pass|test|memory)
keyring-backend = "{{ .KeyringBackend }}"
# CLI output format (text|json)
output = "{{ .Output }}"
# <host>:<port> to Tendermint RPC interface for this chain
node = "{{ .Node }}"
# Transaction broadcasting mode (sync|async)
broadcast-mode = "{{ .BroadcastMode }}"


###############################################################################
###                          Osmosis Tx Configuration                       ###
###############################################################################
# Amount of gas per transaction
gas = "{{ .Gas }}"
# Price per unit of gas (ex: 0.005uosmo)
gas-prices = "{{ .GasPrices }}"
gas-adjustment = "{{ .GasAdjustment }}"
fees = "{{ .Fees }}"
`

// writeConfigToFile parses defaultConfigTemplate, renders config using the template and writes it to
// configFilePath.
func writeConfigToFile(configFilePath string, config *OsmosisCustomClient) error {
	var buffer bytes.Buffer

	tmpl := template.New("clientConfigFileTemplate")
	configTemplate, err := tmpl.Parse(defaultConfigTemplate)
	if err != nil {
		return err
	}

	if err := configTemplate.Execute(&buffer, config); err != nil {
		return err
	}

	return os.WriteFile(configFilePath, buffer.Bytes(), 0o600)
}

// getClientConfig reads values from client.toml file and unmarshalls them into ClientConfig
func getClientConfig(configPath string, v *viper.Viper) (*OsmosisCustomClient, error) {
	v.AddConfigPath(configPath)
	v.SetConfigName("client")
	v.SetConfigType("toml")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	conf := new(OsmosisCustomClient)
	if err := v.Unmarshal(conf); err != nil {
		return nil, err
	}

	return conf, nil
}

// Reads the custom extra values in the config.toml file if set.
// If they are, then use them.
func SetCustomEnvVariablesFromClientToml(ctx client.Context) {
	configFilePath := filepath.Join(ctx.HomeDir, "config", "client.toml")

	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		return
	}

	viper := ctx.Viper
	viper.SetConfigFile(configFilePath)

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	setEnvFromConfig := func(key string, envVar string) {
		// if the user sets the env key manually, then we don't want to override it
		if os.Getenv(envVar) != "" {
			return
		}

		// reads from the config file
		val := viper.GetString(key)
		if val != "" {
			// Sets the env for this instance of the app only.
			os.Setenv(envVar, val)
		}
	}

	// Bound custom flags to environment variable
	// gas
	setEnvFromConfig("gas", "OSMOSISD_GAS")
	setEnvFromConfig("gas-prices", "OSMOSISD_GAS_PRICES")
	setEnvFromConfig("gas-adjustment", "OSMOSISD_GAS_ADJUSTMENT")
	// fees
	setEnvFromConfig("fees", "OSMOSISD_FEES")
}

func errUnknownConfigKey(key string) error {
	return fmt.Errorf("unknown configuration key: %q", key)
}
