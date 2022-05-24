package main

import (
	"os"
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

	if err := os.Chown(tmCfgPath, os.Getuid(), os.Getgid()); err != nil {
		return err
	}

	if err := os.Chmod(tmCfgPath, 0777); err != nil {
		return err
	}

	tmconfig.WriteConfigFile(tmCfgPath, valConfig)
	return nil
}
