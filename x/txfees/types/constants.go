package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
)

// ConsensusMinFee is a governance set parameter from prop 354 (https://www.mintscan.io/osmosis/proposals/354)
// It was intended to be .0025 uosmo / gas
// In v30, we set it to 0.01 uosmo / gas
var ConsensusMinFee osmomath.Dec = osmomath.NewDecWithPrec(1, 2)

// DefaultNullAddress is the default value for the null account. Portion of tx fees are sent to this account to effectively burn the tokens.
// This address is equivalent to osmo1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqmcn030.
var DefaultNullAddress = sdk.AccAddress([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
