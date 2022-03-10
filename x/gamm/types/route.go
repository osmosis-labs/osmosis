package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	appparams "github.com/osmosis-labs/osmosis/v7/app/params"
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

func (routes SwapAmountInRoutes) IsOsmoRoutedMultihop() bool {
	return len(routes) == 2 && (routes[0].TokenOutDenom == appparams.BaseCoinUnit)
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

func (routes SwapAmountOutRoutes) IsOsmoRoutedMultihop() bool {
	return len(routes) == 2 && (routes[1].TokenInDenom == appparams.BaseCoinUnit)
}
