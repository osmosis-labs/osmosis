package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ignite/cli/ignite/pkg/cosmosaccount"
	"github.com/ignite/cli/ignite/pkg/cosmosclient"

	cl "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

const (
	expectedPoolId        uint64 = 1
	addressPrefix                = "osmo"
	defaultClientHomePath        = "~/.osmosisd-local"
	consensusFee                 = "1500uosmo"
	denom0                       = "uosmo"
	denom1                       = "uion"
	accountNamePrefix            = "lo-test"
	numPositions                 = 1_000
	minAmountDeposited           = int64(1_000_000)
	randSeed                     = 1
	maxAmountDeposited           = 1_00_000_000
)

var (
	defaultAccountName = fmt.Sprintf("%s%d", accountNamePrefix, 1)
	exponentAtPriceOne = sdk.OneInt().Neg()
	defaultMinAmount   = sdk.ZeroInt()
)

func main() {
	ctx := context.Background()

	clientHomePath := defaultClientHomePath
	if len(os.Args) > 2 {
		clientHomePath = os.Args[1]
	}

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
	igniteClient.Factory = igniteClient.Factory.WithGas(300000).WithGasAdjustment(1.3).WithFees(consensusFee)

	list, err := igniteClient.AccountRegistry.List()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("initial listing")
	fmt.Println(list)
	a, err := igniteClient.AccountRegistry.Keyring.List()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(a)
	// _, err = igniteClient.AccountRegistry.Import("lo-test1", "notice oak worry limit wrap speak medal online prefer cluster roof addict wrist behave treat actual wasp year salad speed social layer crew genius", "1234")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	statusResp, err := igniteClient.Status(ctx)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("connected to: ", "chain-id", statusResp.NodeInfo.Network, "height", statusResp.SyncInfo.LatestBlockHeight)

	// Instantiate a query client for your `blog` blockchain
	// clQueryClient := types.NewQueryClient(igniteClient.Context())

	// // // Query pool with id 1 and create new if does not exist.
	// // _, err = clQueryClient.Pool(ctx, &types.QueryPoolRequest{PoolId: expectedPoolId})
	// // if err != nil {
	// // 	if !strings.Contains(err.Error(), types.PoolNotFoundError{PoolId: expectedPoolId}.Error()) {
	// // 		log.Fatal(err)
	// // 	}
	// // 	createdPoolId := createPool(igniteClient, defaultAccountName)
	// // 	if createdPoolId != expectedPoolId {
	// // 		log.Fatalf("created pool id (%d), expected pool id (%d)", createdPoolId, expectedPoolId)
	// // 	}
	// // }

	minTick, maxTick := cl.GetMinAndMaxTicksFromExponentAtPriceOne(exponentAtPriceOne)
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
			upperTick = maxTick - rand.Int63n(int64(math.Abs(float64(maxTick-lowerTick)))+1)

			tokenDesired0 = sdk.NewCoin(denom0, sdk.NewInt(rand.Int63n(maxAmountDeposited)))
			tokenDesired1 = sdk.NewCoin(denom1, sdk.NewInt(rand.Int63n(maxAmountDeposited)))
		)

		log.Println("creating position: pool id", expectedPoolId, "accountName", accountName, "lowerTick", lowerTick, "upperTick", upperTick, "token0Desired", tokenDesired0, "tokenDesired1", tokenDesired1, "defaultMinAmount", defaultMinAmount)
		amt0, amt1, liquidity := createPosition(igniteClient, expectedPoolId, accountName, lowerTick, upperTick, tokenDesired0, tokenDesired1, defaultMinAmount, defaultMinAmount)
		log.Println("created position: amt0", amt0, "amt1", amt1, "liquidity", liquidity)
	}
}

func createPool(igniteClient cosmosclient.Client, accountName string) uint64 {
	msg := &model.MsgCreateConcentratedPool{
		Sender:                    getAccountAddressFromKeyring(igniteClient, accountName),
		Denom1:                    denom0,
		Denom0:                    denom1,
		TickSpacing:               1,
		PrecisionFactorAtPriceOne: exponentAtPriceOne,
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
	accAddr1 := "osmo1cyyzpxplxdzkeea7kwsydadg87357qnahakaks"
	acc1Name := "lo-test1"
	msg := &types.MsgCreatePosition{
		PoolId:          poolId,
		Sender:          accAddr1,
		LowerTick:       lowerTick,
		UpperTick:       upperTick,
		TokenDesired0:   tokenDesired0,
		TokenDesired1:   tokenDesired1,
		TokenMinAmount0: tokenMinAmount0,
		TokenMinAmount1: tokenMinAmount1,
	}
	txResp, err := client.BroadcastTx(acc1Name, msg)
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
	a, err := igniteClient.AccountRegistry.List()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(a)
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
