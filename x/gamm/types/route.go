package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	swaproutertypes "github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
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

type MultihopRoute interface {
	Length() int
	PoolIds() []uint64
	IntermediateDenoms() []string
}

// ConvertAmountInRoutes converts gamm swap exact amount in routes to swaprouter routes.
// This is a temporary function to be used until we make the route protos be shared between
// x/gamm and x/swaprouter instead of duplicating them in each module.
func ConvertAmountInRoutes(gammRoutes []SwapAmountInRoute) []swaproutertypes.SwapAmountInRoute {
	swaprouterRoutes := make([]swaproutertypes.SwapAmountInRoute, 0, len(gammRoutes))
	for _, route := range gammRoutes {
		swaprouterRoutes = append(swaprouterRoutes, swaproutertypes.SwapAmountInRoute{
			PoolId:        route.PoolId,
			TokenOutDenom: route.TokenOutDenom,
		})
	}
	return swaprouterRoutes
}

// ConvertAmountOutRoutes converts gamm swap exact amount out routes to swaprouter routes.
// This is a temporary function to be used until we make the route protos be shared between
// x/gamm and x/swaprouter instead of duplicating them in each module.
func ConvertAmountOutRoutes(gammRoutes []SwapAmountOutRoute) []swaproutertypes.SwapAmountOutRoute {
	swaprouterRoutes := make([]swaproutertypes.SwapAmountOutRoute, 0, len(gammRoutes))
	for _, route := range gammRoutes {
		swaprouterRoutes = append(swaprouterRoutes, swaproutertypes.SwapAmountOutRoute{
			PoolId:       route.PoolId,
			TokenInDenom: route.TokenInDenom,
		})
	}
	return swaprouterRoutes
}
