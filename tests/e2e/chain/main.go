package chain

func Init(id string) (*Chain, error) {
	chain, err := New(id)
	if err != nil {
		return nil, err
	}
	if err := InitNodes(chain); err != nil {
		return nil, err
	}
	if err := InitGenesis(chain); err != nil {
		return nil, err
	}
	return chain, nil
}
