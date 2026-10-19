package types

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type SwapAmountInRoutes []SwapAmountInRoute

func (routes SwapAmountInRoutes) Validate() error {
	if len(routes) == 0 {
		return ErrEmptyRoutes
	}

	for _, route := range routes {
		err := sdk.ValidateDenom(route.TokenOutDenom)
		if err != nil {
			return err
		}
	}

	return nil
}

func (routes SwapAmountInRoutes) IntermediateDenoms() []string {
	if len(routes) < 2 {
		return nil
	}
	intermediateDenoms := make([]string, 0, len(routes)-1)
	for _, route := range routes[:len(routes)-1] {
		intermediateDenoms = append(intermediateDenoms, route.TokenOutDenom)
	}

	return intermediateDenoms
}

func (routes SwapAmountInRoutes) PoolIds() []uint64 {
	poolIds := make([]uint64, 0, len(routes))
	for _, route := range routes {
		poolIds = append(poolIds, route.PoolId)
	}
	return poolIds
}

func (routes SwapAmountInRoutes) Length() int {
	return len(routes)
}

// Key returns an injective encoding of the route, used for duplicate detection.
// Each hop is encoded as uvarint(PoolId) || uvarint(len(TokenOutDenom)) || TokenOutDenom.
func (routes SwapAmountInRoutes) Key() string {
	buf := make([]byte, 0, len(routes)*routeKeyHopSizeHint)
	for _, route := range routes {
		buf = appendRouteKeyHop(buf, route.PoolId, route.TokenOutDenom)
	}
	return string(buf)
}

type SwapAmountOutRoutes []SwapAmountOutRoute

func (routes SwapAmountOutRoutes) Validate() error {
	if len(routes) == 0 {
		return ErrEmptyRoutes
	}

	for _, route := range routes {
		err := sdk.ValidateDenom(route.TokenInDenom)
		if err != nil {
			return err
		}
	}

	return nil
}

func (routes SwapAmountOutRoutes) IntermediateDenoms() []string {
	if len(routes) < 2 {
		return nil
	}
	intermediateDenoms := make([]string, 0, len(routes)-1)
	for _, route := range routes[1:] {
		intermediateDenoms = append(intermediateDenoms, route.TokenInDenom)
	}

	return intermediateDenoms
}

func (routes SwapAmountOutRoutes) PoolIds() []uint64 {
	poolIds := make([]uint64, 0, len(routes))
	for _, route := range routes {
		poolIds = append(poolIds, route.PoolId)
	}
	return poolIds
}

func (routes SwapAmountOutRoutes) Length() int {
	return len(routes)
}

// Key returns an injective encoding of the route, used for duplicate detection.
// Each hop is encoded as uvarint(PoolId) || uvarint(len(TokenInDenom)) || TokenInDenom.
func (routes SwapAmountOutRoutes) Key() string {
	buf := make([]byte, 0, len(routes)*routeKeyHopSizeHint)
	for _, route := range routes {
		buf = appendRouteKeyHop(buf, route.PoolId, route.TokenInDenom)
	}
	return string(buf)
}

// routeKeyHopSizeHint is the preallocated key bytes per hop: two uvarints plus a typical denom.
const routeKeyHopSizeHint = 32

// appendRouteKeyHop appends a length-prefixed hop encoding so that keys are injective
// regardless of which characters the denom regex permits.
func appendRouteKeyHop(buf []byte, poolId uint64, denom string) []byte {
	buf = binary.AppendUvarint(buf, poolId)
	buf = binary.AppendUvarint(buf, uint64(len(denom)))
	return append(buf, denom...)
}

// containsDuplicateRoute returns true if any two routes have the same key.
// It runs in time linear in the total encoded size of the routes.
func containsDuplicateRoute[R interface{ Key() string }](routes []R) bool {
	seen := make(map[string]struct{}, len(routes))
	for _, route := range routes {
		key := route.Key()
		if _, ok := seen[key]; ok {
			return true
		}
		seen[key] = struct{}{}
	}
	return false
}

// ValidateSwapAmountInSplitRoute validates a slice of SwapAmountInSplitRoute.
//
// returns an error if any of the following are true:
// - the slice is empty
// - any SwapAmountInRoute in the slice is invalid
// - the last TokenOutDenom of any SwapAmountInRoute in the slice does not match the TokenOutDenom of the previous SwapAmountInRoute in the slice
// - there are duplicate SwapAmountInRoutes in the slice
func ValidateSwapAmountInSplitRoute(splitRoutes []SwapAmountInSplitRoute) error {
	if len(splitRoutes) == 0 {
		return ErrEmptyRoutes
	}

	// validate every multihop path
	previousLastDenomOut := ""
	multihopRoutes := make([]SwapAmountInRoutes, 0, len(splitRoutes))
	for _, splitRoute := range splitRoutes {
		multihopRoute := splitRoute.Pools

		err := SwapAmountInRoutes(multihopRoute).Validate()
		if err != nil {
			return err
		}

		lastDenomOut := multihopRoute[len(multihopRoute)-1].TokenOutDenom

		if previousLastDenomOut != "" && lastDenomOut != previousLastDenomOut {
			return InvalidFinalTokenOutError{TokenOutGivenA: previousLastDenomOut, TokenOutGivenB: lastDenomOut}
		}

		previousLastDenomOut = lastDenomOut

		multihopRoutes = append(multihopRoutes, multihopRoute)
	}

	if containsDuplicateRoute(multihopRoutes) {
		return ErrDuplicateRoutesNotAllowed
	}

	return nil
}

// ValidateSwapAmountOutSplitRoute validates a slice of SwapAmountOutSplitRoute and returns an error if any of the following are true:
// - the slice is empty
// - any SwapAmountOutRoute in the slice is invalid
// - the first TokenInDenom of any SwapAmountOutRoute in the slice does not match the TokenInDenom of the previous SwapAmountOutRoute in the slice
// - there are duplicate SwapAmountOutRoutes in the slice
func ValidateSwapAmountOutSplitRoute(splitRoutes []SwapAmountOutSplitRoute) error {
	if len(splitRoutes) == 0 {
		return ErrEmptyRoutes
	}

	// validate every multihop path
	previousFirstDenomIn := ""
	multihopRoutes := make([]SwapAmountOutRoutes, 0, len(splitRoutes))
	for _, splitRoute := range splitRoutes {
		multihopRoute := splitRoute.Pools

		err := SwapAmountOutRoutes(multihopRoute).Validate()
		if err != nil {
			return err
		}

		firstDenomIn := multihopRoute[0].TokenInDenom

		if previousFirstDenomIn != "" && firstDenomIn != previousFirstDenomIn {
			return InvalidFinalTokenOutError{TokenOutGivenA: previousFirstDenomIn, TokenOutGivenB: firstDenomIn}
		}

		previousFirstDenomIn = firstDenomIn

		multihopRoutes = append(multihopRoutes, multihopRoute)
	}

	if containsDuplicateRoute(multihopRoutes) {
		return ErrDuplicateRoutesNotAllowed
	}

	return nil
}
