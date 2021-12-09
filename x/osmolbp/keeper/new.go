package keeper

import (
	"github.com/cosmos/cosmos-sdk/types"
)

// moja mama to
type S struct {
	Num int
}

var _ types.Msg = S{}

// s *S
// moja mama to
