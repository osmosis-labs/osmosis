package main

import (
	"context"
	"log"
	"os/user"

	"github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
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
	totalOsmosisValidators   = 150
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
	nonEIPBlockData []nonEIPBlockData
}

type validatorInfo struct {
	operatorAddress string
	moniker         string
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

	// Set correct Osmosis address prefix
	params.SetAddressPrefixes()

	statusResp, err := igniteClient.Status(ctx)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("connected to: ", "chain-id", statusResp.NodeInfo.Network, "height", statusResp.SyncInfo.LatestBlockHeight)

	consensusValidators, err := queryAllTendermintValidators(ctx, igniteClient, statusResp.SyncInfo.LatestBlockHeight)
	if err != nil {
		log.Fatal(err)
	}

	validatorsMap := make(map[string]validatorInfo)
	for _, validator := range consensusValidators {
		validatorsMap[string(validator.PubKey.Address())] = validatorInfo{}
	}

	stakingClient := stakingtypes.NewQueryClient(igniteClient.Context())
	stakingValidators, err := queryAllStakingValidators(ctx, stakingClient)
	if err != nil {
		log.Fatal(err)
	}

	encodingConfig := app.GetEncodingConfig()
	if err := stakingValidators.UnpackInterfaces(encodingConfig.Marshaler); err != nil {
		log.Fatal(err)
	}
	sdkValidators := stakingValidators.ToSDKValidators()

	for _, validator := range sdkValidators {
		consensusAddress, err := validator.GetConsAddr()
		if err != nil {
			log.Fatal(err)
		}

		if _, ok := validatorsMap[string(consensusAddress)]; ok {
			validatorsMap[string(consensusAddress)] = validatorInfo{
				operatorAddress: validator.GetOperator().String(),
				moniker:         validator.GetMoniker(),
			}
		} else {
			log.Println("did not find consensus validator", "moniker", validator.GetMoniker())
		}
	}

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

	_, err = stakingClient.Params(ctx, &stakingtypes.QueryParamsRequest{})
	if err != nil {
		log.Fatal(err)
	}

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
			proposerAddress := string(block.Block.Header.ProposerAddress)
			validatorInfo, ok := validatorsMap[proposerAddress]
			if !ok {
				log.Fatal("did not find validator with consensus address ", proposerAddress)
			}
			moniker := validatorInfo.moniker

			log.Println("full block + gas price below min", "moniker", moniker, "gas wanted", gasWanted, "height", height, "avg gas price", averageGasPrice)

			blockData := nonEIPBlockData{
				height:      uint64(height),
				gasWanted:   uint64(gasWanted),
				avgGasPrice: averageGasPrice,
			}

			if existingValData, ok := nonEIPPatchValidatorMap[moniker]; ok {
				existingValData.nonEIPBlockData = append(existingValData.nonEIPBlockData, blockData)
				nonEIPPatchValidatorMap[moniker] = existingValData
			} else {

				newValData := nonEIPValidatorData{
					nonEIPBlockData: []nonEIPBlockData{blockData},
				}
				nonEIPPatchValidatorMap[moniker] = newValData
			}
		}
	}

	log.Println("Summary of non-EIP patch validators")
	for moniker, valData := range nonEIPPatchValidatorMap {
		log.Println("moniker", moniker)
		for _, blockData := range valData.nonEIPBlockData {
			log.Println("height", blockData.height, "gas wanted", blockData.gasWanted, "avg gas price", blockData.avgGasPrice)
		}
		log.Printf("\n\n")
	}
}

func queryAllTendermintValidators(ctx context.Context, client cosmosclient.Client, height int64) ([]*types.Validator, error) {
	// Query tendermint validators
	var (
		page    int = 1
		perPage int = 100
	)

	result := make([]*types.Validator, 0, totalOsmosisValidators)

	validators, err := client.RPC.Validators(ctx, &height, &page, &perPage)
	if err != nil {
		return nil, err
	}

	page++
	result = append(result, validators.Validators...)

	validators, err = client.RPC.Validators(ctx, &height, &page, &perPage)
	if err != nil {
		return nil, err
	}

	result = append(result, validators.Validators...)

	return result, nil
}

// query staking module validators
func queryAllStakingValidators(ctx context.Context, client stakingtypes.QueryClient) (stakingtypes.Validators, error) {
	validatorResponse, err := client.Validators(ctx, &stakingtypes.QueryValidatorsRequest{
		Status: stakingtypes.BondStatusBonded,
	})
	if err != nil {
		log.Fatal(err)
	}

	result := make(stakingtypes.Validators, 0, totalOsmosisValidators)
	result = append(result, validatorResponse.Validators...)

	validatorResponse, err = client.Validators(ctx, &stakingtypes.QueryValidatorsRequest{
		Status: stakingtypes.BondStatusBonded,
		Pagination: &query.PageRequest{
			Key: validatorResponse.Pagination.NextKey,
		},
	})
	if err != nil {
		return nil, err
	}

	result = append(result, validatorResponse.Validators...)

	return result, nil
}

func getClientHomePath() string {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatal(err)
		return ""
	}

	return currentUser.HomeDir + localosmosisFromHomePath
}
