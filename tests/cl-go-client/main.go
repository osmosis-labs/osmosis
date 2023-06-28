package main

import (
	"context"
	"flag"
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

	clqueryproto "github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/client/queryproto"
	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/model"
	cltypes "github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/types"
	incentivestypes "github.com/osmosis-labs/osmosis/v16/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v16/x/lockup/types"
	poolmanagerqueryproto "github.com/osmosis-labs/osmosis/v16/x/poolmanager/client/queryproto"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v16/x/poolmanager/types"
	epochstypes "github.com/osmosis-labs/osmosis/x/epochs/types"
)

// operation defines the desired operation to be run by this script.
type operation int

const (
	// createPositions creates positions in the CL pool with id expectedPoolId.
	createPositions operation = iota

	// makeManySmallSwaps makes many swaps in the CL pool with id expectedPoolId.
	makeManySmallSwaps

	// makeManyLargeSwaps makes many large swaps in the CL pool with id expectedPoolId.
	// it takes one large amount and swaps it into the pool. Then, takes output token
	// and swaps it back while accounting for the spread factor. This is done to
	// ensure that we cross ticks while minimizing the chance of running out of funds or liquidity.
	makeManyInvertibleLargeSwaps

	// createExternalCLIncentives creates external CL incentives.
	createExternalCLIncentives

	// createPoolOperation creates a pool with expectedPoolId.
	createPoolOperation
)

const (
	expectedPoolId           uint64 = 1
	addressPrefix                   = "osmo"
	localosmosisFromHomePath        = "/.osmosisd-local"
	consensusFee                    = "1500uosmo"
	denom0                          = "uosmo"
	denom1                          = "uusdc"
	tickSpacing              int64  = 100
	accountNamePrefix               = "lo-test"
	// Note, this is localosmosis-specific.
	expectedEpochIdentifier = "hour"
	numPositions            = 100
	numSwaps                = 100
	minAmountDeposited      = int64(1_000_000)
	randSeed                = 1
	maxAmountDeposited      = 1_00_000_000
	maxAmountSingleSwap     = 1_000_000
	largeSwapAmount         = 90_000_000_000
)

var (
	defaultAccountName  = fmt.Sprintf("%s%d", accountNamePrefix, 1)
	defaultMinAmount    = sdk.ZeroInt()
	defaultSpreadFactor = sdk.MustNewDecFromStr("0.001")
	externalGaugeCoins  = sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(1000_000_000)))
	accountMutex        sync.Mutex
)

func main() {
	var (
		desiredOperation int
	)

	flag.IntVar(&desiredOperation, "operation", 0, fmt.Sprintf("operation to run:\ncreate positions: %v, make many swaps: %v", createPositions, makeManySmallSwaps))

	flag.Parse()

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

	// Print warnings with common problems
	log.Printf("\n\n\nWARNING 1: your localosmosis and client home are assummed to be %s. Run 'osmosisd get-env' and confirm it matches the path you see printed here\n\n\n", clientHome)

	log.Printf("\n\n\nWARNING 2: you are attempting to interact with pool id %d.\nConfirm that the pool exists. if this is not the pool you want to interact with, please change the expectedPoolId variable in the code\n\n\n", expectedPoolId)

	log.Println("\n\n\nWARNING 3: sometimes the script hangs when just started. In that case, kill it and restart")

	// Check if need to create pool before every opperation.
	if operation(desiredOperation) != createPoolOperation {
		createPoolOp(igniteClient)
	}

	rand.Seed(randSeed)

	switch operation(desiredOperation) {
	case createPositions:
		createManyRandomPositions(igniteClient, expectedPoolId, numPositions)
		return
	case makeManySmallSwaps:
		swapRandomSmallAmountsContinuously(igniteClient, expectedPoolId, numSwaps)
		return
	case makeManyInvertibleLargeSwaps:
		swapGivenLargeAmountsBothDirections(igniteClient, expectedPoolId, numSwaps, largeSwapAmount)
	case createExternalCLIncentives:
		createExternalCLIncentive(igniteClient, expectedPoolId, externalGaugeCoins, expectedEpochIdentifier)
	case createPoolOperation:
		createPoolOp(igniteClient)
	default:
		log.Fatalf("invalid operation: %d", desiredOperation)
	}
}

func createManyRandomPositions(igniteClient cosmosclient.Client, poolId uint64, numPositions int) {
	minTick, maxTick := cltypes.MinInitializedTick, cltypes.MaxTick
	log.Println(minTick, " ", maxTick)
	for i := 0; i < numPositions; i++ {
		var (
			// 1 to 9. These are localosmosis keyring test accounts with names such as:
			// lo-test1
			// lo-test2
			// ...
			randAccountNum = rand.Intn(8) + 1
			accountName    = fmt.Sprintf("%s%d", accountNamePrefix, randAccountNum)
			// minTick <= lowerTick <= upperTick
			lowerTick = roundTickDown(rand.Int63n(maxTick-minTick+1)+minTick, tickSpacing)
			// lowerTick <= upperTick <= maxTick
			upperTick = roundTickDown(maxTick-rand.Int63n(int64(math.Abs(float64(maxTick-lowerTick)))), tickSpacing)

			tokenDesired0 = sdk.NewCoin(denom0, sdk.NewInt(rand.Int63n(maxAmountDeposited)))
			tokenDesired1 = sdk.NewCoin(denom1, sdk.NewInt(rand.Int63n(maxAmountDeposited)))
			tokensDesired = sdk.NewCoins(tokenDesired0, tokenDesired1)
		)

		runMessageWithRetries(func() error {
			_, _, _, err := createPosition(igniteClient, expectedPoolId, accountName, lowerTick, upperTick, tokensDesired, defaultMinAmount, defaultMinAmount)
			return err
		})
	}
}

func swapRandomSmallAmountsContinuously(igniteClient cosmosclient.Client, poolId uint64, numSwaps int) {
	for i := 0; i < numSwaps; i++ {
		var (
			randAccountNum = rand.Intn(8) + 1
			accountName    = fmt.Sprintf("%s%d", accountNamePrefix, randAccountNum)

			isToken0In = rand.Intn(2) == 0

			tokenOutMinAmount = sdk.OneInt()
		)

		tokenInDenom := denom0
		tokenOutDenom := denom1
		if !isToken0In {
			tokenInDenom = denom1
			tokenOutDenom = denom0
		}
		tokenInCoin := sdk.NewCoin(tokenInDenom, sdk.NewInt(rand.Int63n(maxAmountSingleSwap)))

		runMessageWithRetries(func() error {
			_, err := makeSwap(igniteClient, expectedPoolId, accountName, tokenInCoin, tokenOutDenom, tokenOutMinAmount)
			return err
		})
	}

	log.Println("finished swapping, num swaps done", numSwaps)
}

func swapGivenLargeAmountsBothDirections(igniteClient cosmosclient.Client, poolId uint64, numSwaps int, largeStartAmount int64) {
	var (
		randAccountNum = rand.Intn(8) + 1
		accountName    = fmt.Sprintf("%s%d", accountNamePrefix, randAccountNum)

		isToken0In = rand.Intn(2) == 0

		tokenOutMinAmount = sdk.OneInt()
	)

	tokenInDenom := denom0
	tokenOutDenom := denom1
	if !isToken0In {
		tokenInDenom = denom1
		tokenOutDenom = denom0
	}

	tokenInCoin := sdk.NewCoin(tokenInDenom, sdk.NewInt(largeStartAmount))

	for i := 0; i < numSwaps; i++ {
		runMessageWithRetries(func() error {
			tokenOut, err := makeSwap(igniteClient, expectedPoolId, accountName, tokenInCoin, tokenOutDenom, tokenOutMinAmount)

			if err == nil {
				// Swap the resulting amount out back while accounting for spread factor.
				// This is to make sure we can continue swapping back and forth and not run
				// out of funds or liquidity.
				tempTokenInDenom := tokenInCoin.Denom
				// new token in = token out / (1 - spread factor)
				tokenInCoin = sdk.NewCoin(tokenOutDenom, tokenOut.ToDec().Quo(sdk.OneDec().Sub(defaultSpreadFactor)).RoundInt())
				tokenOutDenom = tempTokenInDenom
			}

			return err
		})
	}

	log.Println("finished swapping, num swaps done", numSwaps)
}

func createExternalCLIncentive(igniteClient cosmosclient.Client, poolId uint64, gaugeCoins sdk.Coins, expectedEpochIdentifier string) {
	var (
		randAccountNum = rand.Intn(8) + 1
		accountName    = fmt.Sprintf("%s%d", accountNamePrefix, randAccountNum)
	)

	epochsQueryClient := epochstypes.NewQueryClient(igniteClient.Context())
	currentEpochResponse, err := epochsQueryClient.CurrentEpoch(context.Background(), &epochstypes.QueryCurrentEpochRequest{
		expectedEpochIdentifier,
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("current epoch", currentEpochResponse.CurrentEpoch, "epoch identifier", expectedEpochIdentifier)

	log.Println("querying epoch info. Note that incentives are distributed at the end of epoch. That's why it matters.")
	epochInfosRespose, err := epochsQueryClient.EpochInfos(context.Background(), &epochstypes.QueryEpochsInfoRequest{})
	if err != nil {
		log.Fatal(err)
	}

	if len(epochInfosRespose.Epochs) > 0 {
		lastEpochInfo := epochInfosRespose.Epochs[len(epochInfosRespose.Epochs)-1]
		log.Println("epoch duration", lastEpochInfo, "next epoch start time", lastEpochInfo.StartTime.Add(lastEpochInfo.Duration))
	} else {
		log.Println("could not find information about previous epoch. If duration is too long, this test might be infeasible")
	}

	//.Create gauge
	runMessageWithRetries(func() error {
		return createGauge(igniteClient, expectedPoolId, accountName, gaugeCoins)
	})

	epochAfterGaugeCreation := int64(-1)
	for {
		// Wait for 1 epoch to pass
		currentEpochResponse, err = epochsQueryClient.CurrentEpoch(context.Background(), &epochstypes.QueryCurrentEpochRequest{
			expectedEpochIdentifier,
		})
		if err != nil {
			log.Fatal(err)
		}
		if epochAfterGaugeCreation == -1 {
			log.Println("current epoch after gauge creation", currentEpochResponse.CurrentEpoch)
			log.Println("waiting for next epoch...")
			epochAfterGaugeCreation = currentEpochResponse.CurrentEpoch
			continue
		}

		// One epoch after gauge creation has passed
		if epochAfterGaugeCreation+1 == currentEpochResponse.CurrentEpoch {
			log.Println("next epoch reached, checking incentive records...")
			break
		}

		log.Println("current epoch", currentEpochResponse.CurrentEpoch)
		time.Sleep(5 * time.Second)
	}

	clQueryClient := clqueryproto.NewQueryClient(igniteClient.Context())

	incentiveRecordsResponse, err := clQueryClient.IncentiveRecords(context.Background(), &clqueryproto.IncentiveRecordsRequest{
		PoolId: expectedPoolId,
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("incentive records. If empty, something probably went wrong", incentiveRecordsResponse.IncentiveRecords)
}

func createPool(igniteClient cosmosclient.Client, accountName string) uint64 {
	msg := &model.MsgCreateConcentratedPool{
		Sender:       getAccountAddressFromKeyring(igniteClient, accountName),
		Denom1:       denom0,
		Denom0:       denom1,
		TickSpacing:  1,
		SpreadFactor: defaultSpreadFactor,
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

func createPosition(client cosmosclient.Client, poolId uint64, senderKeyringAccountName string, lowerTick int64, upperTick int64, tokensProvided sdk.Coins, tokenMinAmount0, tokenMinAmount1 sdk.Int) (amountCreated0, amountCreated1 sdk.Int, liquidityCreated sdk.Dec, err error) {
	accountMutex.Lock() // Lock access to getAccountAddressFromKeyring
	senderAddress := getAccountAddressFromKeyring(client, senderKeyringAccountName)
	accountMutex.Unlock() // Unlock access to getAccountAddressFromKeyring

	log.Println("creating position: pool id", expectedPoolId, "accountName", senderKeyringAccountName, "lowerTick", lowerTick, "upperTick", upperTick, "token0Desired", tokensProvided[0], "tokenDesired1", tokensProvided[1], "defaultMinAmount", defaultMinAmount)

	msg := &cltypes.MsgCreatePosition{
		PoolId:          poolId,
		Sender:          senderAddress,
		LowerTick:       lowerTick,
		UpperTick:       upperTick,
		TokensProvided:  tokensProvided,
		TokenMinAmount0: tokenMinAmount0,
		TokenMinAmount1: tokenMinAmount1,
	}
	txResp, err := client.BroadcastTx(senderKeyringAccountName, msg)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, err
	}
	resp := cltypes.MsgCreatePositionResponse{}
	if err := txResp.Decode(&resp); err != nil {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, err
	}
	log.Println("created position: amt0", resp.Amount0, "amt1", resp.Amount1, "liquidity", resp.LiquidityCreated)
	return resp.Amount0, resp.Amount1, resp.LiquidityCreated, nil
}

func makeSwap(client cosmosclient.Client, poolId uint64, senderKeyringAccountName string, tokenInCoin sdk.Coin, tokenOutDenom string, tokenOutMinAmount sdk.Int) (sdk.Int, error) {
	accountMutex.Lock() // Lock access to getAccountAddressFromKeyring
	senderAddress := getAccountAddressFromKeyring(client, senderKeyringAccountName)
	accountMutex.Unlock() // Unlock access to getAccountAddressFromKeyring

	log.Println("making swap in: pool id", expectedPoolId, "tokenIn", tokenInCoin, "tokenOutDenom", tokenOutDenom, "tokenOutMinAmount", tokenOutMinAmount, "from", senderKeyringAccountName)

	msg := &poolmanagertypes.MsgSwapExactAmountIn{
		Sender: senderAddress,
		Routes: []poolmanagertypes.SwapAmountInRoute{
			{
				PoolId:        expectedPoolId,
				TokenOutDenom: tokenOutDenom,
			},
		},
		TokenIn:           tokenInCoin,
		TokenOutMinAmount: tokenOutMinAmount,
	}
	txResp, err := client.BroadcastTx(senderKeyringAccountName, msg)
	if err != nil {
		return sdk.Int{}, err
	}
	resp := poolmanagertypes.MsgSwapExactAmountInResponse{}
	if err := txResp.Decode(&resp); err != nil {
		return sdk.Int{}, err
	}

	log.Println("swap made, token out amount: ", resp.TokenOutAmount)
	return resp.TokenOutAmount, nil
}

func createGauge(client cosmosclient.Client, poolId uint64, senderKeyringAccountName string, gaugeCoins sdk.Coins) error {
	accountMutex.Lock() // Lock access to getAccountAddressFromKeyring
	senderAddress := getAccountAddressFromKeyring(client, senderKeyringAccountName)
	accountMutex.Unlock() // Unlock access to getAccountAddressFromKeyring

	log.Println("creating CL gauge for pool id", expectedPoolId, "gaugeCoins", gaugeCoins)

	msg := &incentivestypes.MsgCreateGauge{
		IsPerpetual: false,
		Owner:       senderAddress,
		DistributeTo: lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.NoLock,
		},
		StartTime:         time.Now(),
		Coins:             gaugeCoins,
		NumEpochsPaidOver: 5,
		PoolId:            expectedPoolId,
	}
	txResp, err := client.BroadcastTx(senderKeyringAccountName, msg)
	if err != nil {
		return err
	}
	resp := &incentivestypes.MsgCreateGaugeResponse{}
	if err := txResp.Decode(resp); err != nil {
		return err
	}

	log.Println("gauge created")
	return nil
}

func createPoolOp(igniteClient cosmosclient.Client) {
	// Instantiate a query client
	poolManagerClient := poolmanagerqueryproto.NewQueryClient(igniteClient.Context())

	// Query pool with id 1 and create new if does not exist.
	_, err := poolManagerClient.Pool(context.Background(), &poolmanagerqueryproto.PoolRequest{PoolId: expectedPoolId})
	if err != nil {
		if !strings.Contains(err.Error(), poolmanagertypes.FailedToFindRouteError{PoolId: expectedPoolId}.Error()) {
			log.Fatal(err)
		}
		createdPoolId := createPool(igniteClient, defaultAccountName)
		if createdPoolId != expectedPoolId {
			log.Fatalf("created pool id (%d), expected pool id (%d)", createdPoolId, expectedPoolId)
		}
	} else {
		log.Println("pool already exists. Tweak expectedPoolId variable if you want another pool, current expectedPoolId", expectedPoolId)
	}
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

func runMessageWithRetries(runMsg func() error) {
	maxRetries := 100
	var err error
	for j := 0; j < maxRetries; j++ {
		err := runMsg()
		if err != nil {
			log.Println("retrying, error occurred while running message: ", err)
			time.Sleep(8 * time.Second)
		} else {
			time.Sleep(200 * time.Millisecond)
			break
		}
	}
	if err != nil {
		log.Fatal(err)
	}
}

func roundTickDown(tickIndex int64, tickSpacing int64) int64 {
	// Round the tick index down to the nearest tick spacing if the tickIndex is in between authorized tick values
	// Note that this is Euclidean modulus.
	// The difference from default Go modulus is that Go default results
	// in a negative remainder when the dividend is negative.
	// Consider example tickIndex = -17, tickSpacing = 10
	// tickIndexModulus = tickIndex % tickSpacing = -7
	// tickIndexModulus = -7 + 10 = 3
	// tickIndex = -17 - 3 = -20
	tickIndexModulus := tickIndex % tickSpacing
	if tickIndexModulus < 0 {
		tickIndexModulus += tickSpacing
	}

	if tickIndexModulus != 0 {
		tickIndex = tickIndex - tickIndexModulus
	}
	return tickIndex
}
