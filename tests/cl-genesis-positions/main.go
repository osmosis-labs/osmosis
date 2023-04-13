package main

import sdk "github.com/cosmos/cosmos-sdk/types"

type operation int

const (
	getData operation = iota
	convertPositions
	setupKeyringAccounts
	localOsmosisGenesisEdit
)

const (
	pathToFilesFromRoot = "tests/cl-genesis-positions/"

	positionsFileName       = "subgraph_positions.json"
	osmosisStateFileName    = "genesis.json"
	bigbangPosiionsFileName = "bigbang_positions.json"

	localOsmosisHomePath = "/osmosis/.osmosisd/"

	denom0 = "uusdc"
	denom1 = "uweth"

	useKeyringAccounts = true

	writeGenesisToDisk = true

	writeBigBangConfigToDisk = true
)

func main() {
	desiredOperation := localOsmosisGenesisEdit

	switch desiredOperation {
	case getData:
		GetUniV3SubgraphData()
		break
	case convertPositions:
		var localKeyringAccounts []sdk.AccAddress
		if useKeyringAccounts {
			localKeyringAccounts = getLocalKeyringAccounts()
		}

		ConvertUniswapToOsmosis(localKeyringAccounts)

		break
	case setupKeyringAccounts:
		getLocalKeyringAccounts()
		break
	case localOsmosisGenesisEdit:
		localKeyringAccounts := getLocalKeyringAccounts()

		state := ConvertUniswapToOsmosis(localKeyringAccounts)

		EditLocalOsmosisGenesis(state)
	default:
		panic("Invalid operation")
	}
}
