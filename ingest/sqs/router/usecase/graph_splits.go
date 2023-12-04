package usecase

import "github.com/osmosis-labs/osmosis/v21/ingest/sqs/router/usecase/route"

// TODO:
// - consider constructing double linked list
//    * slice forward
//    * slice backward
//    * if slice backward contains a pool, there is a cycle so we must wait

// CONTRACT: there are no cycles
func routesToGraph(routes []route.RouteImpl) []route.RouteImpl {

	// Create pool adjacency matrix
	adjacencyMatrixPools := map[uint64][]uint64{}

	// Keep track of start pools
	startPoolIDs := map[uint64]struct{}{}

	hasDiffered := false

	for _, route := range routes {

		pools := route.GetPools()

		// TODO: err check
		firstPool := pools[0]

		startPoolIDs[firstPool.GetId()] = struct{}{}

		for i := 0; i < len(pools)-1; i++ {
			curPool := pools[i]
			nextPool := pools[i+1]

			adjacency, ok := adjacencyMatrixPools[curPool.GetId()]
			if ok {
				adjacencyMatrixPools[curPool.GetId()] = append(adjacency, nextPool.GetId())
			} else {
				adjacencyMatrixPools[curPool.GetId()] = []uint64{nextPool.GetId()}
			}
		}
	}

	// for _, poolID := range startPoolIDs {

	// }

	return nil
}

func keepIterating(routes []route.RouteImpl) []route.RouteImpl {

	// Create pool adjacency matrix
	adjacencyMatrixPools := map[uint64][]uint64{}

	// Keep track of start pools
	startPoolIDs := map[uint64]struct{}{}

	hasDiffered := false

	for _, route := range routes {

		pools := route.GetPools()

		// TODO: err check
		firstPool := pools[0]

		startPoolIDs[firstPool.GetId()] = struct{}{}

		for i := 0; i < len(pools)-1; i++ {
			curPool := pools[i]
			nextPool := pools[i+1]

			adjacency, ok := adjacencyMatrixPools[curPool.GetId()]
			if ok {
				adjacencyMatrixPools[curPool.GetId()] = append(adjacency, nextPool.GetId())
			} else {
				adjacencyMatrixPools[curPool.GetId()] = []uint64{nextPool.GetId()}
			}
		}
	}

	// for _, poolID := range startPoolIDs {

	// }

	return nil
}
