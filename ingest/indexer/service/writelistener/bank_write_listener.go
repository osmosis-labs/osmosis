package writelistener

import (
	"bytes"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

var _ storetypes.WriteListener = (*bankWriteListener)(nil)

type bankWriteListener struct {
}

func NewBank() storetypes.WriteListener {
	return &bankWriteListener{}
}

// OnWrite implements types.WriteListener.
func (s *bankWriteListener) OnWrite(storeKey storetypes.StoreKey, key []byte, value []byte, delete bool) error {
	// Track updated supplies.
	if len(key) > 0 && bytes.Equal(banktypes.SupplyKey, key[:1]) {
		// TODO: deal with supply updates.
	}

	return nil
}
