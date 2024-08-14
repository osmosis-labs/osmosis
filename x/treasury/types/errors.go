package types

import (
	errorsmod "cosmossdk.io/errors"
)

var ErrNoSuchBurnTaxExemptionAddress = errorsmod.Register(ModuleName, 1, "no such address in extemption list")
