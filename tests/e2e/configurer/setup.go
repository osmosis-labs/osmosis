package configurer

import (
	"fmt"
)

type setupFn func(configurer Configurer) error

func baseSetup(configurer Configurer) error {
	if err := configurer.RunValidators(); err != nil {
		return err
	}
	return nil
}

func withIBC(setupHandler setupFn) setupFn {
	return func(configurer Configurer) error {
		if err := setupHandler(configurer); err != nil {
			return err
		}

		if err := configurer.RunIBC(); err != nil {
			return err
		}

		return nil
	}
}

func withUpgrade(setupHandler setupFn) setupFn {
	return func(configurer Configurer) error {
		if err := setupHandler(configurer); err != nil {
			return err
		}

		upgradeConfigurer, ok := configurer.(*UpgradeConfigurer)
		if !ok {
			return fmt.Errorf("to run with upgrade, %v must be set during initialization", &UpgradeConfigurer{})
		}

		if err := upgradeConfigurer.CreatePreUpgradeState(); err != nil {
			return err
		}

		if err := upgradeConfigurer.RunUpgrade(); err != nil {
			return err
		}

		return nil
	}
}
