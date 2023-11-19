package main

import (
	"context"
	"log"
	"os/user"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ignite/cli/ignite/pkg/cosmosaccount"
	"github.com/ignite/cli/ignite/pkg/cosmosclient"

	"github.com/osmosis-labs/osmosis/v20/app"
	"github.com/osmosis-labs/osmosis/v20/app/params"
)

const (
	addressPrefix            = "osmo"
	localosmosisFromHomePath = "/.osmosisd-local"
	consensusFee             = "3000uosmo"
)

var (
	baseFeeMin = sdk.MustNewDecFromStr("0.01")
)

type nonEIPBlockData struct {
	height      uint64
	gasWanted   uint64
	avgGasPrice sdk.Dec
}

type nonEIPValidatorData struct {
	consensusAddress string
	nonEIPBlockData  []nonEIPBlockData
}

func main() {
	ctx := context.Background()

	clientHome := getClientHomePath()

	// Create a Cosmos igniteClient instance
	igniteClient, err := cosmosclient.New(
		ctx,
		cosmosclient.WithAddressPrefix(addressPrefix),
		cosmosclient.WithKeyringBackend(cosmosaccount.KeyringTest),
		cosmosclient.WithHome(clientHome),
		// cosmosclient.WithNodeAddress("https://osmosis-rpc.polkachu.com:443"),
		cosmosclient.WithNodeAddress("http://65.109.20.216:26657"),
		// cosmosclient.WithNodeAddress("https://rpc.archive.osmosis.zone:443"),
	)
	if err != nil {
		log.Fatal(err)
	}
	igniteClient.TxFactory = igniteClient.TxFactory.WithGas(300000).WithGasAdjustment(1.3).WithFees(consensusFee)

	params.SetAddressPrefixes()

	// config := sdk.GetConfig()
	// config.SetBech32PrefixForValidator(addressPrefix+"valoper", addressPrefix+"valoperpub")

	// sdk.GetConfig().SetBech32PrefixForValidator(addressPrefix+"valoper", addressPrefix+"valoperpub")

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

	// https://www.mintscan.io/osmosis/block/12402004)
	startHeight := int64(12402004)
	// https://www.mintscan.io/osmosis/block/12402500
	endHeight := int64(12402500)

	consensusParams, err := igniteClient.RPC.ConsensusParams(ctx, &statusResp.SyncInfo.LatestBlockHeight)
	if err != nil {
		log.Fatal(err)
	}

	// Make it 90% of the actual max gas parameter
	maxGas := consensusParams.ConsensusParams.Block.MaxGas
	maxGasThreshold := maxGas * 90 / 100
	log.Println("max gas", maxGas, "max gas threshold", maxGasThreshold)

	// validatorToMonikerMap := make(map[string]string)

	stakingClient := stakingtypes.NewQueryClient(igniteClient.Context())

	_, err = stakingClient.Params(ctx, &stakingtypes.QueryParamsRequest{})
	if err != nil {
		log.Fatal(err)
	}

	// validatorResponse, err := stakingClient.Validator(ctx, &stakingtypes.QueryValidatorRequest{
	// 	ValidatorAddr: valAddressStr,
	// })

	log.Println("queried staking params")

	nonEIPPatchValidatorMap := make(map[string]nonEIPValidatorData)
	txDecoder := app.GetEncodingConfig().TxConfig.TxDecoder()

	for height := startHeight; height <= endHeight; height++ {
		// log some feedback every once in a while
		if height%100 == 0 {
			log.Printf("block %d\n", height)
		}

		blockTXs, err := igniteClient.GetBlockTXs(ctx, height)
		if err != nil {
			log.Fatal(err)
		}

		gasWanted := int64(0)
		gasUsed := int64(0)

		// Query fot the validator who proposed this block
		block, err := igniteClient.RPC.Block(ctx, &height)
		if err != nil {
			log.Fatal(err)
		}

		gasPriceSum := sdk.ZeroDec()

		for _, blockTx := range blockTXs {
			gasUsed += blockTx.Raw.TxResult.GasUsed
			gasWanted += blockTx.Raw.TxResult.GasWanted

			tx, err := txDecoder(blockTx.Raw.Tx)
			if err != nil {
				log.Fatal(err)
			}

			feeTx, ok := tx.(sdk.FeeTx)
			if !ok {
				log.Fatal("tx is not a FeeTx")
			}

			txFee := feeTx.GetFee().AmountOf("uosmo")

			// Skip this TX because the fee is paid in another token
			// We only focus on uosmo
			if txFee.IsZero() {
				continue
			}

			txGas := feeTx.GetGas()

			gasPrice := txFee.ToLegacyDec().Quo(sdk.NewDec(int64(txGas)))

			gasPriceSum.AddMut(gasPrice)
		}

		// Skip empty block.
		if len(blockTXs) == 0 {
			continue
		}

		averageGasPrice := gasPriceSum.QuoInt64(int64(len(blockTXs)))

		if gasWanted >= maxGasThreshold && averageGasPrice.LT(baseFeeMin) {
			log.Println("full block + gas price below min", "proposer address", block.Block.Header.ProposerAddress, "gas wanted", gasWanted, "height", height, "avg gas price", averageGasPrice)
			proposerAddress := block.Block.Header.ProposerAddress.String()

			blockData := nonEIPBlockData{
				height:      uint64(height),
				gasWanted:   uint64(gasWanted),
				avgGasPrice: averageGasPrice,
			}

			if existingValData, ok := nonEIPPatchValidatorMap[proposerAddress]; ok {
				existingValData.nonEIPBlockData = append(existingValData.nonEIPBlockData, blockData)
				nonEIPPatchValidatorMap[proposerAddress] = existingValData
			} else {
				newValData := nonEIPValidatorData{
					consensusAddress: proposerAddress,
					nonEIPBlockData:  []nonEIPBlockData{blockData},
				}
				nonEIPPatchValidatorMap[proposerAddress] = newValData
			}
		}
	}

	log.Println("Summary of non-EIP patch validators")
	for _, valData := range nonEIPPatchValidatorMap {
		log.Println("consensus address", valData.consensusAddress)
		for _, blockData := range valData.nonEIPBlockData {
			log.Println("height", blockData.height, "gas wanted", blockData.gasWanted, "avg gas price", blockData.avgGasPrice)
		}
		log.Printf("\n\n")
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
