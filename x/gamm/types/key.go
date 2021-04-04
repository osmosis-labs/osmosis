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

func GetPoolIdFromShareDenom(denom string) (uint64, error) {
	r, err := regexp.Compile(`(gamm/pool/)([\d]+)$`)
	if err != nil {
		return 0, err
	}

	split := r.FindSubmatch([]byte(denom))
	if len(split) != 3 {
		return 0, fmt.Errorf("invalid pool share denom")
	}

	return strconv.ParseUint(string(split[2]), 10, 64)
}

func GetKeyPaginationPoolNumbers(poolId uint64) []byte {
	return append(PaginationPoolNumbers, sdk.Uint64ToBigEndian(poolId)...)
}
