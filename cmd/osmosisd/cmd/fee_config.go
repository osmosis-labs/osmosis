package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"text/template"

	"github.com/cosmos/cosmos-sdk/client"
	viper "github.com/spf13/viper"
)

type FeeConfig struct {
	DefaultBaseFee string `mapstructure:"default-base-fee" json:"default-base-fee"`
}

const defaultFeeConfigTemplate = `Fee Config File
# This file contains configuration for fee related parameters.


`

// writeConfigToFile parses defaultConfigTemplate, renders config using the template and writes it to
// configFilePath. If nil is provided as config, the default config is used.
func writeFeeConfigToFile(configFilePath string, config *FeeConfig) error {
	var buffer bytes.Buffer
	defaultOsmosisCustomFeeConfig := &FeeConfig{
		DefaultBaseFee: "1000000", // TODO: change
	}

	tmpl := template.New("feeConfigFileTemplate")
	configTemplate, err := tmpl.Parse(defaultFeeConfigTemplate)
	if err != nil {
		return err
	}

	// Loop through the fields of the provided config and replace values in the default client
	if config != nil {
		configValue := reflect.ValueOf(config).Elem()
		defaultValue := reflect.ValueOf(defaultOsmosisCustomFeeConfig).Elem()

		for i := 0; i < configValue.NumField(); i++ {
			configField := configValue.Field(i)
			defaultField := defaultValue.Field(i)

			// Check if the field is a pointer type
			if configField.Kind() == reflect.Ptr {
				// If it's a pointer type, check if it's nil
				if !configField.IsNil() {
					defaultField.Set(configField.Elem())
				}
			} else {
				// For non-pointer types, check if the value is the zero value
				if !reflect.DeepEqual(configField.Interface(), reflect.Zero(configField.Type()).Interface()) {
					defaultField.Set(configField)
				}
			}
		}
	}

	if err := configTemplate.Execute(&buffer, defaultOsmosisCustomFeeConfig); err != nil {
		return err
	}

	return os.WriteFile(configFilePath, buffer.Bytes(), 0o600)
}

// getFeeConfig reads values from fee.toml file and unmarshalls them into FeeConfig
func getFeeConfig(configPath string, v *viper.Viper) (*FeeConfig, error) {
	v.AddConfigPath(configPath)
	v.SetConfigName("client")
	v.SetConfigType("toml")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	conf := new(FeeConfig)
	if err := v.Unmarshal(conf); err != nil {
		return nil, err
	}

	return conf, nil
}

func retrieveFeeConfig(ctx client.Context) (*FeeConfig, error) {
	configPath := filepath.Join(ctx.HomeDir, "config")
	configFilePath := filepath.Join(configPath, "fee.toml")

	// if config.toml file does not exist we create it and write default ClientConfig values into it.
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		if err := ensureConfigPath(configPath); err != nil {
			return nil, fmt.Errorf("couldn't make client config: %v", err)
		}

		feeConf := FeeConfig{
			DefaultBaseFee: "0.025uosmo",
		}

		if err := writeFeeConfigToFile(configFilePath, &feeConf); err != nil {
			return nil, fmt.Errorf("could not write client config to the file: %v", err)
		}
	}

	feeConf, err := getFeeConfig(configPath, ctx.Viper)
	if err != nil {
		return nil, err
	}

	return feeConf, nil
}
