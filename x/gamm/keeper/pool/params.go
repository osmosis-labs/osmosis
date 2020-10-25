package pool

import sdk "github.com/cosmos/cosmos-sdk/types"

// TODO: handle this constanct in the param space
var maxInRatio, _ = sdk.NewDecFromStr("0.5")
var maxOutRatio, _ = sdk.NewDecFromStr("0.34")
