package commondomain

import (
	storetypes "cosmossdk.io/store/types"
)

// BlockPoolUpdateTracker is an interface that defines the methods for the block pool update tracker.
type BlockUpdateProcessUtilsI interface {
	// ProcessBlockChangeSet processes the change set and notifies the write listeners.
	ProcessBlockChangeSet() error

	// SetChangeSet sets the change set on the block update process utils.
	SetChangeSet(changeSet []*storetypes.StoreKVPair)
}

// BlockUpdateProcessUtils is a struct that implements BlockUpdateProcessUtilsI
// and contains the necessary data to process the block change set.
type BlockUpdateProcessUtils struct {
	WriteListeners map[storetypes.StoreKey][]WriteListener
	StoreKeyMap    map[string]storetypes.StoreKey
	ChangeSet      []*storetypes.StoreKVPair
}

var _ BlockUpdateProcessUtilsI = &BlockUpdateProcessUtils{}

// ProcessBlockChangeSet implements BlockUpdateProcessUtilsI.
func (b *BlockUpdateProcessUtils) ProcessBlockChangeSet() error {
	if b.ChangeSet == nil {
		return nil
	}

	for _, kv := range b.ChangeSet {
		for _, listener := range b.WriteListeners[b.StoreKeyMap[kv.StoreKey]] {
			if err := listener.OnWrite(b.StoreKeyMap[kv.StoreKey], kv.Key, kv.Value, kv.Delete); err != nil {
				return err
			}
		}
	}

	return nil
}

// SetChangeSet implements BlockUpdateProcessUtilsI.
func (b *BlockUpdateProcessUtils) SetChangeSet(changeSet []*storetypes.StoreKVPair) {
	b.ChangeSet = changeSet
}
