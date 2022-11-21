package cli

import (
	"errors"
	"strconv"

	flag "github.com/spf13/pflag"

	"github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
)

func swapAmountInRoutes(fs *flag.FlagSet) ([]types.SwapAmountInRoute, error) {
	swapRoutePoolIds, err := fs.GetStringArray(FlagSwapRoutePoolIds)
	if err != nil {
		return nil, err
	}

	swapRouteDenoms, err := fs.GetStringArray(FlagSwapRouteDenoms)
	if err != nil {
		return nil, err
	}

	if len(swapRoutePoolIds) != len(swapRouteDenoms) {
		return nil, errors.New("swap route pool ids and denoms mismatch")
	}

	routes := []types.SwapAmountInRoute{}
	for index, poolIDStr := range swapRoutePoolIds {
		pID, err := strconv.Atoi(poolIDStr)
		if err != nil {
			return nil, err
		}
		routes = append(routes, types.SwapAmountInRoute{
			PoolId:        uint64(pID),
			TokenOutDenom: swapRouteDenoms[index],
		})
	}
	return routes, nil
}

func swapAmountOutRoutes(fs *flag.FlagSet) ([]types.SwapAmountOutRoute, error) {
	swapRoutePoolIds, err := fs.GetStringArray(FlagSwapRoutePoolIds)
	if err != nil {
		return nil, err
	}

	swapRouteDenoms, err := fs.GetStringArray(FlagSwapRouteDenoms)
	if err != nil {
		return nil, err
	}

	if len(swapRoutePoolIds) != len(swapRouteDenoms) {
		return nil, errors.New("swap route pool ids and denoms mismatch")
	}

	routes := []types.SwapAmountOutRoute{}
	for index, poolIDStr := range swapRoutePoolIds {
		pID, err := strconv.Atoi(poolIDStr)
		if err != nil {
			return nil, err
		}
		routes = append(routes, types.SwapAmountOutRoute{
			PoolId:       uint64(pID),
			TokenInDenom: swapRouteDenoms[index],
		})
	}
	return routes, nil
}
