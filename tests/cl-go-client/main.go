package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ignite/cli/ignite/pkg/cosmosaccount"
	"github.com/ignite/cli/ignite/pkg/cosmosclient"

	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

const (
	expectedPoolId    uint64 = 1
	addressPrefix            = "osmo"
	clientHomePath           = "/root/.osmosisd-local"
	consensusFee             = "875uosmo"
	denom0                   = "uosmo"
	denom1                   = "uion"
	accountNamePrefix        = "lo-test"
)

var (
	defaultAccountName = fmt.Sprintf("%s%d", accountNamePrefix, 1)
)

func main() {

	ctx := context.Background()

	// Create a Cosmos igniteClient instance
	igniteClient, err := cosmosclient.New(
		ctx,
		cosmosclient.WithAddressPrefix(addressPrefix),
		cosmosclient.WithKeyringBackend(cosmosaccount.KeyringTest),
		cosmosclient.WithHome(clientHomePath),
	)
	if err != nil {
		log.Fatal(err)
	}
	igniteClient.Factory = igniteClient.Factory.WithFees(consensusFee)

	statusResp, err := igniteClient.Status(ctx)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("connected to: ", "chain-id", statusResp.NodeInfo.Network, "height", statusResp.SyncInfo.LatestBlockHeight)

	// Instantiate a query client for your `blog` blockchain
	clQueryClient := types.NewQueryClient(igniteClient.Context())

	// Query pool with id 1 and create new if does not exist.
	_, err = clQueryClient.Pool(ctx, &types.QueryPoolRequest{PoolId: expectedPoolId})
	if err != nil {
		if !strings.Contains(err.Error(), types.PoolNotFoundError{PoolId: expectedPoolId}.Error()) {
			log.Fatal(err)
		}
		createdPoolId := createPool(igniteClient, defaultAccountName)
		if createdPoolId != expectedPoolId {
			log.Fatalf("created pool id (%d), expected pool id (%d)", createdPoolId, expectedPoolId)
		}
	}

	var (
		// TODO: randomize params, use multiple accounts and many positions.
		accountName            = defaultAccountName
		lowerTick        int64 = -1000
		upperTick        int64 = 1000
		tokenDesired0          = sdk.NewCoin(denom0, sdk.NewInt(10000))
		tokenDesired1          = sdk.NewCoin(denom1, sdk.NewInt(10000))
		defaultMinAmount       = sdk.OneInt()
	)
	amt0, amt1, liquidity := createPosition(igniteClient, expectedPoolId, accountName, lowerTick, upperTick, tokenDesired0, tokenDesired1, defaultMinAmount, defaultMinAmount)
	log.Println("created position: amt0", amt0, "amt1", amt1, "liquidity", liquidity)
}

func createPool(igniteClient cosmosclient.Client, accountName string) uint64 {
	msg := &model.MsgCreateConcentratedPool{
		Sender:                    getAccountAddressFromKeyring(igniteClient, accountName),
		Denom1:                    denom0,
		Denom0:                    denom1,
		TickSpacing:               1,
		PrecisionFactorAtPriceOne: sdk.OneInt().Neg(),
		SwapFee:                   sdk.ZeroDec(),
	}
	txResp, err := igniteClient.BroadcastTx(accountName, msg)
	if err != nil {
		log.Fatal(err)
	}
	resp := model.MsgCreateConcentratedPoolResponse{}
	if err := txResp.Decode(&resp); err != nil {
		log.Fatal(err)
	}
	return resp.PoolID
}

func createPosition(client cosmosclient.Client, poolId uint64, senderKeyringAccountName string, lowerTick int64, upperTick int64, tokenDesired0, tokenDesired1 sdk.Coin, tokenMinAmount0, tokenMinAmount1 sdk.Int) (amountCreated0, amountCreated1 sdk.Int, liquidityCreated sdk.Dec) {
	msg := &types.MsgCreatePosition{
		PoolId:          poolId,
		Sender:          getAccountAddressFromKeyring(client, senderKeyringAccountName),
		LowerTick:       lowerTick,
		UpperTick:       upperTick,
		TokenDesired0:   tokenDesired0,
		TokenDesired1:   tokenDesired1,
		TokenMinAmount0: tokenMinAmount0,
		TokenMinAmount1: tokenMinAmount1,
	}
	txResp, err := client.BroadcastTx(senderKeyringAccountName, msg)
	if err != nil {
		log.Fatal(err)
	}
	resp := types.MsgCreatePositionResponse{}
	if err := txResp.Decode(&resp); err != nil {
		log.Fatal(err)
	}
	return resp.Amount0, resp.Amount1, resp.LiquidityCreated
}

func getAccountAddressFromKeyring(igniteClient cosmosclient.Client, accountName string) string {
	account, err := igniteClient.Account(accountName)
	if err != nil {
		log.Fatal(fmt.Errorf("did not fimf account with name (%s) in the keyring: %w", accountName, err))
	}

	address := account.Address(addressPrefix)
	if err != nil {
		log.Fatal(err)
	}
	return address
}
