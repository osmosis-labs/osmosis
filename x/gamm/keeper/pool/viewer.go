package pool

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Viewer interface{}

type viewer struct {
	cdc      codec.BinaryMarshaler
	storeKey sdk.StoreKey
}
