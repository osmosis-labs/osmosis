package mocks

import (
	"cosmossdk.io/store/types"

	commondomain "github.com/osmosis-labs/osmosis/v31/ingest/common/domain"
)

type BlockUpdateProcessUtilsMock struct {
	ProcessBlockReturn error
	LastSetChangeSet   []*types.StoreKVPair
}

var _ commondomain.BlockUpdateProcessUtilsI = &BlockUpdateProcessUtilsMock{}

// ProcessBlockChangeSet implements domain.BlockUpdateProcessUtilsI.
func (b *BlockUpdateProcessUtilsMock) ProcessBlockChangeSet() error {
	return b.ProcessBlockReturn
}

// SetChangeSet implements domain.BlockUpdateProcessUtilsI.
func (b *BlockUpdateProcessUtilsMock) SetChangeSet(changeSet []*types.StoreKVPair) {
	b.LastSetChangeSet = changeSet
}
