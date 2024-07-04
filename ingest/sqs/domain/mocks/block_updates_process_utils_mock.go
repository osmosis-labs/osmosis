package mocks

import (
	"cosmossdk.io/store/types"

	"github.com/osmosis-labs/osmosis/v25/ingest/sqs/domain"
)

type BlockUpdateProcessUtilsMock struct {
	ProcessBlockReturn error
	LastSetChangeSet   []*types.StoreKVPair
}

var _ domain.BlockUpdateProcessUtilsI = &BlockUpdateProcessUtilsMock{}

// ProcessBlockChangeSet implements domain.BlockUpdateProcessUtilsI.
func (b *BlockUpdateProcessUtilsMock) ProcessBlockChangeSet() error {
	return b.ProcessBlockReturn
}

// SetChangeSet implements domain.BlockUpdateProcessUtilsI.
func (b *BlockUpdateProcessUtilsMock) SetChangeSet(changeSet []*types.StoreKVPair) {
	b.LastSetChangeSet = changeSet
}
