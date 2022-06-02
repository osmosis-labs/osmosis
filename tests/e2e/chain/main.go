package chain

import "time"

func Init(id, dataDir string, validatorConfigs []*ValidatorConfig, votingPeriod time.Duration) (*Chain, error) {
	chain, err := new(id, dataDir)
	if err != nil {
		return nil, err
	}
	if err := initNodes(chain, len(validatorConfigs)); err != nil {
		return nil, err
	}
	if err := initGenesis(chain, votingPeriod); err != nil {
		return nil, err
	}
	if err := initValidatorConfigs(chain, validatorConfigs); err != nil {
		return nil, err
	}
	return chain.export(), nil
}
