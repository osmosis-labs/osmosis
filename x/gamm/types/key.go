package types

const (
	ModuleName = "gamm"

	StoreKey = ModuleName

	RouterKey = ModuleName

	QuerierRoute = ModuleName
)

var (
	PoolAddressPrefix = []byte("gmm_liquidity_pool")
	PoolPrefix        = []byte("gmm_liquidity_pool")
	GlobalPoolNumber  = []byte("gmm_global_pool_number")
)
