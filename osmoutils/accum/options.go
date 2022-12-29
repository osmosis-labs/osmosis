package accum

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var one = sdk.OneDec()

// validate returns nil if Options are valid.
// Error otherwise. Note, that, currently,
// options do not contain any fields. We
// create them to be able to extend in the
// future with auto-compounding logic.
// As a result, this always returns nil.
func (o *Options) validate() error {
	if o == nil {
		return nil
	}
	return nil
}
