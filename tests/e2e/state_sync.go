package e2e

import (
	"path/filepath"

	"github.com/spf13/viper"
	tmconfig "github.com/tendermint/tendermint/config"
)

func configureNodeForStateSync(stateSyncValidatorConfigDir string, trustHeight int64, trustHash string) error {
	tmCfgPath := filepath.Join(stateSyncValidatorConfigDir, "config", "config.toml")

	vpr := viper.New()
	vpr.SetConfigFile(tmCfgPath)
	if err := vpr.ReadInConfig(); err != nil {
		return err
	}

	valConfig := &tmconfig.Config{}
	if err := vpr.Unmarshal(valConfig); err != nil {
		return err
	}
	valConfig.StateSync.Enable = true
	valConfig.StateSync.TrustHeight = trustHeight
	valConfig.StateSync.TrustHash = trustHash
	// configBytes.=

	// valConfig.StateSync = tmconfig.DefaultStateSyncConfig()
	// valConfig.StateSync.Enable = true
	// valConfig.StateSync.RPCServers =
	return nil
}
