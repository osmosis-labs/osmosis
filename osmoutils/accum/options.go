package accum

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var one = sdk.OneDec()

// TODO: test
func (o *Options) validate() error {
	if o == nil {
		return nil
	}

	if o.InflationRate.IsZero() {
		return fmt.Errorf("inflation rate cannot be zero. If you desire regular position, please provide nil in place of Options")
	}

	if o.InflationRate.GTE(one) {
		return fmt.Errorf("inflation rate cannot be greater than or equal to 1")
	}

	return nil
}
