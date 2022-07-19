package twap

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v10/x/gamm/twap/types"
)

type twapkeeper struct {
	storeKey     sdk.StoreKey
	transientKey sdk.TransientStoreKey

	gammkeeper types.AmmInterface
}
