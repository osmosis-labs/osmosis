package main

import (
	"context"
	"log"
	"os/user"

	"github.com/ignite/cli/ignite/pkg/cosmosaccount"
	"github.com/ignite/cli/ignite/pkg/cosmosclient"
)

const (
	addressPrefix            = "osmo"
	localosmosisFromHomePath = "/.osmosisd-local"
	consensusFee             = "3000uosmo"
)

func main() {
	ctx := context.Background()

	clientHome := getClientHomePath()

	// Create a Cosmos igniteClient instance
	igniteClient, err := cosmosclient.New(
		ctx,
		cosmosclient.WithAddressPrefix(addressPrefix),
		cosmosclient.WithKeyringBackend(cosmosaccount.KeyringTest),
		cosmosclient.WithHome(clientHome),
		cosmosclient.WithNodeAddress("https://rpc.archive.osmosis.zone:443"),
	)
	if err != nil {
		log.Fatal(err)
	}
	igniteClient.TxFactory = igniteClient.TxFactory.WithGas(300000).WithGasAdjustment(1.3).WithFees(consensusFee)

	statusResp, err := igniteClient.Status(ctx)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("connected to: ", "chain-id", statusResp.NodeInfo.Network, "height", statusResp.SyncInfo.LatestBlockHeight)

	// Start Block
	// https://www.mintscan.io/osmosis/block/12300000
	// ~ 11 AM Nov 11 ET
	// ~ 1pm Nov 11 Levana messaged about the issue
	//
	// End Block
	// 	// https://www.mintscan.io/osmosis/block/12302500
	//
	//
	// Strategy
	// - Iterate over all blocks between startHeight and endHeight
	//

	startHeight := int64(12300000)
	endHeight := int64(12302500)

	consensusParams, err := igniteClient.RPC.ConsensusParams(ctx, &startHeight)
	if err != nil {
		log.Fatal(err)
	}

	for height := startHeight; height <= endHeight; height++ {
		// log some feedback every once in a while
		if height%100 == 0 {
			log.Printf("block %d\n", height)
		}

		blockTXs, err := igniteClient.GetBlockTXs(ctx, height)
		if err != nil {
			log.Fatal(err)
		}

		// log.Printf("number of transactions in block %d is %d\n", height, len(transactions))

		maxGas := consensusParams.ConsensusParams.Block.MaxGas

		gasWanted := int64(0)
		gasUsed := int64(0)

		for _, blockTx := range blockTXs {
			gasUsed += blockTx.Raw.TxResult.GasUsed
			gasWanted += blockTx.Raw.TxResult.GasWanted
		}

		// log.Printf("block %d gas used (%d) / gas wanted (%d)\n", height, gasUsed, gasWanted)

		if gasUsed >= maxGas {
			log.Printf("block %d gas used (%d) >= max gas (%d)\n", height, gasUsed, maxGas)
		}

		if gasWanted >= maxGas {
			log.Printf("block %d gas wanted (%d) >= max gas (%d)\n", height, gasUsed, maxGas)
		}
	}

}

func getClientHomePath() string {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatal(err)
		return ""
	}

	return currentUser.HomeDir + localosmosisFromHomePath
}
