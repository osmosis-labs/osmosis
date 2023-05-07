package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os/user"
	"strings"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ignite/cli/ignite/pkg/cosmosaccount"
	"github.com/ignite/cli/ignite/pkg/cosmosclient"

	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	cltypes "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	poolmanagerqueryproto "github.com/osmosis-labs/osmosis/v15/x/poolmanager/client/queryproto"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

const (
	expectedPoolId           uint64 = 1
	addressPrefix                   = "osmo"
	localosmosisFromHomePath        = "/.osmosisd-local"
	consensusFee                    = "1500uosmo"
	denom0                          = "uusdc"
	denom1                          = "uosmo"
	accountNamePrefix               = "lo-test"
	numPositions                    = 1_000
	minAmountDeposited              = int64(1_000_000)
	randSeed                        = 1
	maxAmountDeposited              = 1_00_000_000
)

var (
	defaultAccountName = fmt.Sprintf("%s%d", accountNamePrefix, 1)
	defaultMinAmount   = sdk.ZeroInt()
	accountMutex       sync.Mutex
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
	)
	if err != nil {
		log.Fatal(err)
	}
	igniteClient.Factory = igniteClient.Factory.WithGas(300000).WithGasAdjustment(1.3).WithFees(consensusFee)

	statusResp, err := igniteClient.Status(ctx)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("connected to: ", "chain-id", statusResp.NodeInfo.Network, "height", statusResp.SyncInfo.LatestBlockHeight)

	// Instantiate a query client
	clQueryClient := poolmanagerqueryproto.NewQueryClient(igniteClient.Context())

	// Print warnings with common problems
	log.Printf("\n\n\nWARNING 1: your localosmosis and client home are assummed to be %s. Run 'osmosisd get-env' and confirm it matches the path you see printed here\n\n\n", clientHome)

	log.Printf("\n\n\nWARNING 2: you are attempting to interact with pool id %d.\nConfirm that the pool exists. if this is not the pool you want to interact with, please change the expectedPoolId variable in the code\n\n\n", expectedPoolId)

	log.Println("\n\n\nWARNING 3: sometimes the script hangs when just started. In that case, kill it and restart")

	// Query pool with id 1 and create new if does not exist.
	_, err = clQueryClient.Pool(ctx, &poolmanagerqueryproto.PoolRequest{PoolId: expectedPoolId})
	if err != nil {
		if !strings.Contains(err.Error(), poolmanagertypes.FailedToFindRouteError{PoolId: expectedPoolId}.Error()) {
			log.Fatal(err)
		}
		createdPoolId := createPool(igniteClient, defaultAccountName)
		if createdPoolId != expectedPoolId {
			log.Fatalf("created pool id (%d), expected pool id (%d)", createdPoolId, expectedPoolId)
		}
	}

	minTick, maxTick := cltypes.MinTick, cltypes.MaxTick
	log.Println(minTick, " ", maxTick)

	rand.Seed(randSeed)

	for i := 0; i < numPositions; i++ {
		var (
			// 1 to 9. These are localosmosis keyring test accounts with names such as:
			// lo-test1
			// lo-test2
			// ...
			randAccountNum = rand.Intn(8) + 1
			accountName    = fmt.Sprintf("%s%d", accountNamePrefix, randAccountNum)
			// minTick <= lowerTick <= upperTick
			lowerTick = rand.Int63n(maxTick-minTick+1) + minTick
			// lowerTick <= upperTick <= maxTick
			upperTick = maxTick - rand.Int63n(int64(math.Abs(float64(maxTick-lowerTick))))

			tokenDesired0 = sdk.NewCoin(denom0, sdk.NewInt(rand.Int63n(maxAmountDeposited)))
			tokenDesired1 = sdk.NewCoin(denom1, sdk.NewInt(rand.Int63n(maxAmountDeposited)))
		)

		log.Println("creating position: pool id", expectedPoolId, "accountName", accountName, "lowerTick", lowerTick, "upperTick", upperTick, "token0Desired", tokenDesired0, "tokenDesired1", tokenDesired1, "defaultMinAmount", defaultMinAmount)

		maxRetries := 100
		for j := 0; j < maxRetries; j++ {
			amt0, amt1, liquidity := createPosition(igniteClient, expectedPoolId, accountName, lowerTick, upperTick, tokenDesired0, tokenDesired1, defaultMinAmount, defaultMinAmount)
			if err == nil {
				log.Println("created position: amt0", amt0, "amt1", amt1, "liquidity", liquidity)
				break
			}
			time.Sleep(8 * time.Second)
		}
	}
}

func createPool(igniteClient cosmosclient.Client, accountName string) uint64 {
	msg := &model.MsgCreateConcentratedPool{
		Sender:      getAccountAddressFromKeyring(igniteClient, accountName),
		Denom1:      denom0,
		Denom0:      denom1,
		TickSpacing: 1,
		SwapFee:     sdk.ZeroDec(),
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
	accountMutex.Lock() // Lock access to getAccountAddressFromKeyring
	senderAddress := getAccountAddressFromKeyring(client, senderKeyringAccountName)
	accountMutex.Unlock() // Unlock access to getAccountAddressFromKeyring

	msg := &cltypes.MsgCreatePosition{
		PoolId:          poolId,
		Sender:          senderAddress,
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
	resp := cltypes.MsgCreatePositionResponse{}
	if err := txResp.Decode(&resp); err != nil {
		log.Fatal(err)
	}
	return resp.Amount0, resp.Amount1, resp.LiquidityCreated
}

func getAccountAddressFromKeyring(igniteClient cosmosclient.Client, accountName string) string {
	account, err := igniteClient.Account(accountName)
	if err != nil {
		log.Fatal(fmt.Errorf("did not find account with name (%s) in the keyring: %w", accountName, err))
	}

	address := account.Address(addressPrefix)
	if err != nil {
		log.Fatal(err)
	}
	return address
}

func getClientHomePath() string {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatal(err)
		return ""
	}

	return currentUser.HomeDir + localosmosisFromHomePath
}
