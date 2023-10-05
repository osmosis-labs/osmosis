package main

import (
	"flag"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// operation defines the desired operation to be run by this script.
type operation int

const (
	// getData retrieves the data from the Uniswap subgraph and writes it to disk
	// under pathToFilesFromRoot + positionsFileName path.
	getData operation = iota
	// convertPositions converts the data from the Uniswap subgraph into Osmosis
	// genesis. It reads pathToFilesFromRoot + positionsFileName path
	// run Osmosis app via apptesting, creates positions and writes the genesis
	// under pathToFilesFromRoot + osmosisStateFileName path.
	convertPositions
	// mergeSubgraphAndLocalOsmosisGenesis merges the genesis created from the subgraph data
	// with the localosmosis genesis. This command is meant to be called inside the localosmosis
	// container during setup (see setup.sh). It reads the existing genesis from localosmosisHomePath,
	// updates the concentrated liquidity section to append the CL pool created from the subgraph data,
	// its positions, ticks and accumulators.
	mergeSubgraphAndLocalOsmosisGenesis
)

const (
	pathToFilesFromRoot = "tests/cl-genesis-positions/"

	positionsFileName       = "subgraph_positions.json"
	osmosisGenesisFileName  = "genesis.json"
	bigbangPosiionsFileName = "bigbang_positions.json"

	localOsmosisHomePath = "/osmosis/.osmosisd/"

	denom0 = "uusdc"
	denom1 = "uosmo"
)

var (
	// This is lo-test1 address in localosmosis
	defaultCreatorAddresses = []sdk.AccAddress{sdk.MustAccAddressFromBech32("osmo1cyyzpxplxdzkeea7kwsydadg87357qnahakaks"), sdk.MustAccAddressFromBech32("osmo18s5lynnmx37hq4wlrw9gdn68sg2uxp5rgk26vv")}

	useKeyringAccounts bool

	writeGenesisToDisk bool

	writeBigBangConfigToDisk bool
)

func main() {
	var (
		desiredOperation int
		isLocalOsmosis   bool
	)

	flag.BoolVar(&writeBigBangConfigToDisk, "big-bang", false, fmt.Sprintf("flag indicating whether to write the big bang config to disk at path %s", bigbangPosiionsFileName))
	flag.BoolVar(&writeGenesisToDisk, "genesis", false, fmt.Sprintf("flag indicating whether to write the genesis file to disk at path %s", osmosisGenesisFileName))
	flag.BoolVar(&useKeyringAccounts, "keyring", false, "flag indicating whether to use local test keyring accounts")
	flag.BoolVar(&isLocalOsmosis, "localosmosis", false, "flag indicating whether this is being run inside the localosmosis container")
	flag.IntVar(&desiredOperation, "operation", 0, fmt.Sprintf("operation to run:\nget subgraph data: %v, convert subgraph positions to osmo genesis: %v\nmerge converted subgraph genesis and localosmosis genesis: %v", getData, convertPositions, mergeSubgraphAndLocalOsmosisGenesis))

	flag.Parse()

	fmt.Println("isLocalOsmosis:", isLocalOsmosis)

	pathToSaveFilesAt := pathToFilesFromRoot
	if isLocalOsmosis {
		pathToSaveFilesAt = ""
	}

	// Set this to one of the 'operation' values
	switch operation(desiredOperation) {
	// See definition for more info.
	case getData:
		fmt.Println("Getting data from Uniswap subgraph...")

		GetUniV3SubgraphData(pathToSaveFilesAt + positionsFileName)
		// See definition for more info.
	case convertPositions:
		fmt.Println("Converting positions from subgraph data to Osmosis genesis...")

		var creatorAddresses []sdk.AccAddress
		if useKeyringAccounts {
			fmt.Println("Using local keyring addresses as creators")
			creatorAddresses = GetLocalKeyringAccounts()
		} else {
			fmt.Println("Using default creator addresses")
			creatorAddresses = defaultCreatorAddresses
		}

		ConvertSubgraphToOsmosisGenesis(creatorAddresses, pathToSaveFilesAt+positionsFileName)
		// See definition for more info.
	case mergeSubgraphAndLocalOsmosisGenesis:
		fmt.Println("Merging subgraph and local Osmosis genesis...")
		clState, bankState := ConvertSubgraphToOsmosisGenesis(defaultCreatorAddresses, pathToSaveFilesAt+positionsFileName)

		EditLocalOsmosisGenesis(clState, bankState)
	default:
		panic("Invalid operation")
	}
}
