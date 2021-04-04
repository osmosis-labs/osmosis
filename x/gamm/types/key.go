package types

import (
	"fmt"
	"regexp"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	ModuleName = "gamm"

	StoreKey = ModuleName

	RouterKey = ModuleName

	QuerierRoute = ModuleName
)

var (
	PoolAddressPrefix = []byte("gmm_liquidity_pool")
	GlobalPoolNumber  = []byte("gmm_global_pool_number")
	// Used for querying to paginate the registered pool numbers.
	PaginationPoolNumbers = []byte("gmm_pool_numbers_pagination")
)

func GetPoolShareDenom(poolId uint64) string {
	return fmt.Sprintf("gamm/pool/%d", poolId)
}

func GetPoolIdFromShareDenom(denom string) (uint64, bool) {
	r, err := regexp.Compile(`(gamm/pool/)([\d]+)$`)
	if err != nil {
		return 0, false
	}

	split := r.FindSubmatch([]byte(denom))
	if len(split) != 3 {
		return 0, false
	}

	poolId, err := strconv.ParseUint(string(split[2]), 10, 64)
	if err != nil {
		return 0, false
	}

	return poolId, true
}

func GetKeyPaginationPoolNumbers(poolId uint64) []byte {
	return append(PaginationPoolNumbers, sdk.Uint64ToBigEndian(poolId)...)
}
