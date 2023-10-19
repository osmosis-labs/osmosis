package redis

import poolmanagertypes "github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"

const (
	CfmmPoolKey         = cfmmPoolKey
	ConcentratedPoolKey = concentratedPoolKey
	CosmWasmPoolKey     = cosmWasmPoolKey
)

func CfmmKeyFromPoolTypeAndID(poolType poolmanagertypes.PoolType, ID uint64) string {
	return cfmmKeyFromPoolTypeAndID(poolType, ID)
}
