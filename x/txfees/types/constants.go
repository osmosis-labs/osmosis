package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// ConsensusMinFee is a governance set parameter from prop 354 (https://www.mintscan.io/osmosis/proposals/354)
// Its intended to be .0025 uosmo / gas
var ConsensusMinFee sdk.Dec = sdk.NewDecWithPrec(25, 4)
