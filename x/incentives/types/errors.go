package types

import (
	fmt "fmt"

	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

type UnsupportedPoolToDistribute struct {
	PoolId   uint64
	PoolType poolmanagertypes.PoolType
}

func (e UnsupportedPoolToDistribute) Error() string {
	return fmt.Sprintf("pool id (%d) is not supported type (%s) to distribute", e.PoolId, poolmanagertypes.PoolType_name[int32(e.PoolType)])
}
