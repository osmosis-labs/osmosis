package types

const (
	ModuleName = "gamm"

	StoreKey = ModuleName

	RouterKey = ModuleName

	QuerierRoute = ModuleName
)

var (
	PoolPrefix       = []byte("gmm_liquidity_pool")
	GlobalPoolNumber = []byte("gmm_global_pool_number")
)
