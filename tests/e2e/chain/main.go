package chain

func Init(id string) (*Chain, error) {
	chain, err := new(id)
	if err != nil {
		return nil, err
	}
	if err := initNodes(chain); err != nil {
		return nil, err
	}
	if err := initGenesis(chain); err != nil {
		return nil, err
	}
	if err := initValidatorConfigs(chain); err != nil {
		return nil, err
	}
	return chain, nil
}
