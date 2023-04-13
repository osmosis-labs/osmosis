package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v15/app"
	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	cl "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/math"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	cltypes "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	clgenesis "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types/genesis"
)

type BigBangPosition struct {
	Address    string `json:"address"`
	JoinTime   string `json:"join_time"`
	Liquidity  string `json:"liquidity"`
	LowerTick  string `json:"lower_tick"`
	PoolID     string `json:"pool_id"`
	PositionID string `json:"position_id"`
	UpperTick  string `json:"upper_tick"`
}

type BigBangPositions struct {
	Positions []BigBangPosition `json:"positions"`
}

type OsmosisApp struct {
	App         *app.OsmosisApp
	Ctx         sdk.Context
	QueryHelper *baseapp.QueryServiceTestHelper
	TestAccs    []sdk.AccAddress
}

var (
	uniV3TickBase    = osmomath.MustNewDecFromStr("1.0001")
	osmosisPrecision = 6
)

func ReadSubgraphDataFromDisk() []Position {
	// read in the data from file
	data, err := ioutil.ReadFile(pathToFilesFromRoot + positionsFileName)
	if err != nil {
		fmt.Println("Error reading file:", err)
		os.Exit(1)
	}

	// unmarshal the data into a slice of Position structs
	var positions []Position
	err = json.Unmarshal(data, &positions)
	if err != nil {
		fmt.Println("Error unmarshalling data:", err)
		os.Exit(1)
	}

	return positions
}

func ConvertUniswapToOsmosis(localKeyringAccounts []sdk.AccAddress) *clgenesis.GenesisState {
	positions := ReadSubgraphDataFromDisk()

	osmosis := apptesting.KeeperTestHelper{}
	osmosis.Setup()

	if len(localKeyringAccounts) > 0 {
		fmt.Println("Using local keyring accounts")
		osmosis.TestAccs = make([]sdk.AccAddress, len(localKeyringAccounts))
		for i := 0; i < len(localKeyringAccounts); i++ {
			osmosis.TestAccs[i] = localKeyringAccounts[i]
			fmt.Println(osmosis.TestAccs[i].String())
		}
	} else {
		fmt.Println("Using default osmosis testing accounts")
	}

	initAmounts := sdk.NewCoins(
		sdk.NewCoin(denom0, sdk.NewInt(1000000000000000000)),
		sdk.NewCoin(denom1, sdk.NewInt(1000000000000000000)),
		sdk.NewCoin("uosmo", sdk.NewInt(1000000000000000000)),
	)

	// fund all accounts
	for _, acc := range osmosis.TestAccs {
		err := simapp.FundAccount(osmosis.App.BankKeeper, osmosis.Ctx, acc, initAmounts)
		if err != nil {
			panic(err)
		}
	}

	msgCreatePool := model.MsgCreateConcentratedPool{
		Sender:             osmosis.TestAccs[0].String(),
		Denom0:             denom0,
		Denom1:             denom1,
		TickSpacing:        1,
		ExponentAtPriceOne: sdk.OneInt().Neg(),
		SwapFee:            sdk.MustNewDecFromStr("0.0005"),
	}

	poolId, err := osmosis.App.PoolManagerKeeper.CreatePool(osmosis.Ctx, msgCreatePool)
	if err != nil {
		panic(err)
	}

	fmt.Println(poolId)

	pool, err := osmosis.App.ConcentratedLiquidityKeeper.GetPool(osmosis.Ctx, poolId)
	if err != nil {
		panic(err)
	}

	// Initialize first positon to be 1:1 price
	// this is because the first position must have non-zero token0 and token1 to initialize the price
	// however, our data has first position with non-zero amount.
	_, _, _, _, _, _ = osmosis.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(osmosis.Ctx, pool.(cltypes.ConcentratedPoolExtension), osmosis.TestAccs[0], sdk.NewCoins(sdk.NewCoin(msgCreatePool.Denom0, sdk.NewInt(100)), sdk.NewCoin(msgCreatePool.Denom1, sdk.NewInt(100))))
	if err != nil {
		panic(err)
	}

	clMsgServer := cl.NewMsgServerImpl(osmosis.App.ConcentratedLiquidityKeeper)

	numberOfSuccesfulPositions := 0

	bigBangPositions := make([]BigBangPosition, 0)

	for _, uniV3Position := range positions {

		lowerPrice := parsePrice(uniV3Position.TickLower.Price0)
		upperPrice := parsePrice(uniV3Position.TickUpper.Price0)
		if err != nil {
			panic(err)
		}

		if lowerPrice.GTE(upperPrice) {
			fmt.Printf("lowerPrice (%s) >= upperPrice (%s), skipping", lowerPrice, upperPrice)
			continue
		}

		lowerTickOsmosis, err := math.PriceToTick(lowerPrice, msgCreatePool.ExponentAtPriceOne)
		if err != nil {
			panic(err)
		}

		upperTickOsmosis, err := math.PriceToTick(upperPrice, msgCreatePool.ExponentAtPriceOne)
		if err != nil {
			panic(err)
		}

		if lowerTickOsmosis.GT(upperTickOsmosis) {
			fmt.Printf("lowerTickOsmosis (%s) > upperTickOsmosis (%s), skipping", lowerTickOsmosis, upperTickOsmosis)
			continue
		}

		if lowerTickOsmosis.Equal(upperTickOsmosis) {
			// bump up the upper tick by one. We don't care about having exactly the same tick range
			// Just a roughly similar breakdown
			upperTickOsmosis = upperTickOsmosis.Add(sdk.OneInt())
		}

		depositedAmount0, failedParsing := parseStringToInt(uniV3Position.DepositedToken0)
		if failedParsing {
			fmt.Printf("Failed parsing %s, skipping", uniV3Position.DepositedToken0)
			continue
		}

		depositedAmount1, failedParsing := parseStringToInt(uniV3Position.DepositedToken1)
		if failedParsing {
			fmt.Printf("Failed parsing %s, skipping", uniV3Position.DepositedToken0)
			continue
		}

		randomCreator := osmosis.TestAccs[rand.Intn(len(osmosis.TestAccs))]

		position, err := clMsgServer.CreatePosition(sdk.WrapSDKContext(osmosis.Ctx), &cltypes.MsgCreatePosition{
			PoolId:          poolId,
			Sender:          randomCreator.String(),
			LowerTick:       lowerTickOsmosis.Int64(),
			UpperTick:       upperTickOsmosis.Int64(),
			TokenDesired0:   sdk.NewCoin(msgCreatePool.Denom0, depositedAmount0),
			TokenDesired1:   sdk.NewCoin(msgCreatePool.Denom1, depositedAmount1),
			TokenMinAmount0: sdk.ZeroInt(),
			TokenMinAmount1: sdk.ZeroInt(),
		})

		if err != nil {
			fmt.Printf("\n\n\nWARNING: Failed to create position: %v\n\n\n", err)
			fmt.Printf("attempted creation between ticks (%s) and (%s), desired amount 0: (%s), desired amount 1 (%s)\n", lowerTickOsmosis, upperTickOsmosis, depositedAmount0, depositedAmount1)
			fmt.Println("\n\n")
			continue
		}

		fmt.Printf("created position with liquidity (%s) between ticks (%s) and (%s)\n", position.LiquidityCreated, lowerTickOsmosis, upperTickOsmosis)
		numberOfSuccesfulPositions++

		bigBangPositions = append(bigBangPositions, BigBangPosition{
			Address:    randomCreator.String(),
			PoolID:     strconv.FormatUint(poolId, 10),
			JoinTime:   osmosis.Ctx.BlockTime().String(),
			Liquidity:  position.LiquidityCreated.String(),
			PositionID: strconv.FormatUint(position.PositionId, 10),
			LowerTick:  lowerTickOsmosis.String(),
			UpperTick:  upperTickOsmosis.String(),
		})
	}

	fmt.Printf("\nout of %d uniswap positions, %d were successfully created\n", len(positions), numberOfSuccesfulPositions)

	if writeBigBangConfigToDisk {
		writeBigBangPositionsToState(bigBangPositions)
	}

	if writeGenesisToDisk {
		state := osmosis.App.ExportState(osmosis.Ctx)
		writeStateToDisk(state)
	}

	clGenesis := osmosis.App.ConcentratedLiquidityKeeper.ExportGenesis(osmosis.Ctx)
	return clGenesis
}

func parsePrice(strPrice string) (result sdk.Dec) {
	defer func() {
		r := recover()
		if r != nil {
			fmt.Printf("Recovered in price parsing %s, %s\n", r, strPrice)
			if strPrice[0] == '0' {
				result = cltypes.MinSpotPrice
			} else {
				result = cltypes.MaxSpotPrice
			}
		}

		if result.GT(cltypes.MaxSpotPrice) {
			result = cltypes.MaxSpotPrice
		}

		if result.LT(cltypes.MinSpotPrice) {
			result = cltypes.MinSpotPrice
		}
	}()
	result = osmomath.MustNewDecFromStr(strPrice).SDKDec()
	return result
}

func parseStringToInt(strInt string) (result sdk.Int, failedParsing bool) {
	defer func() {
		r := recover()
		if r != nil {
			fmt.Printf("Recovered in int parsing %s, %s\n", r, strInt)
			failedParsing = true
		}
	}()
	result = osmomath.MustNewDecFromStr(strInt).SDKDec().MulInt64(int64(osmosisPrecision)).TruncateInt()
	return result, failedParsing
}

func writeStateToDisk(state map[string]json.RawMessage) {
	stateBz, err := json.MarshalIndent(state, "", "    ")
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(pathToFilesFromRoot+osmosisStateFileName, stateBz, 0644)
	if err != nil {
		panic(err)
	}
}

func writeBigBangPositionsToState(positions []BigBangPosition) {
	fmt.Println("writing big bang positions to disk")
	positionsBytes, err := json.MarshalIndent(BigBangPositions{
		Positions: positions,
	}, "", "    ")
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(pathToFilesFromRoot+bigbangPosiionsFileName, positionsBytes, 0644)
	if err != nil {
		panic(err)
	}
}
