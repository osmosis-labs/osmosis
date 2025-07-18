package types

import (
	"github.com/osmosis-labs/osmosis/osmomath"
)

// ConsensusMinFee is a governance set parameter from prop 354 (https://www.mintscan.io/osmosis/proposals/354)
// It was intended to be .0025 uosmo / gas
// In v30, we set it to 0.01 uosmo / gas
var ConsensusMinFee osmomath.Dec = osmomath.NewDecWithPrec(1, 2)
