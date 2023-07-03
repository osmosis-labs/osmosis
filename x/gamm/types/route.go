package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	poolmanagertypes "github.com/osmosis-labs/osmosis/v16/x/poolmanager/types"
)

type SwapAmountInRoutes []poolmanagertypes.SwapAmountInRoute

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

type SwapAmountOutRoutes []poolmanagertypes.SwapAmountOutRoute

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
