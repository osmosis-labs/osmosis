package chain

func Init(id, dataDir string, numVal int, pruning []string, pruningKeepRecent []string, pruningInterval []string, snapshotInterval []uint64, snapshotKeepRecent []uint32) (*Chain, error) {
	chain, err := new(id, dataDir)
	if err != nil {
		return nil, err
	}
	if err := initNodes(chain, numVal); err != nil {
		return nil, err
	}
	if err := initGenesis(chain); err != nil {
		return nil, err
	}
	if err := initValidatorConfigs(chain, pruning, pruningKeepRecent, pruningInterval, snapshotInterval, snapshotKeepRecent); err != nil {
		return nil, err
	}
	return chain.export(), nil
}
