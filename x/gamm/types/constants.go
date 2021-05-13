package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	MIN_POOL_ASSETS = 2
	MAX_POOL_ASSETS = 8

	BONE_EXPONENT = 18
)

var (
	// BONE term is borrowed from Balancer Bronze Codebase
	// source: https://github.com/balancer-labs/balancer-core/blob/master/contracts/BConst.sol#L19
	// We assume it stands for Balancer_ONE, but it's funny enough that we decided to use it too.
	BONE = sdk.NewIntWithDecimal(1, BONE_EXPONENT)

	// INIT_POOL_SUPPLY is the amount of new shares to initialize a pool with
	INIT_POOL_SUPPLY = BONE.MulRaw(100)
)
