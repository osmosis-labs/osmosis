package v17

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v17/app/keepers"
)

func FlipTwapSpotPriceRecords(ctx sdk.Context, poolIds []uint64, keepers *keepers.AppKeepers) error {
	for _, poolId := range poolIds {
		// check that this is a cl pool
		_, err := keepers.ConcentratedLiquidityKeeper.GetConcentratedPoolById(ctx, poolId)
		if err != nil {
			return err
		}

		// check that the twap record exists
		clPoolTwapRecords, err := keepers.TwapKeeper.GetAllMostRecentRecordsForPool(ctx, poolId)
		if err != nil {
			return err
		}

		fmt.Println("BEFORE FLIP", clPoolTwapRecords)

		for _, twapRecord := range clPoolTwapRecords {
			twapRecord.LastErrorTime = time.Time{}
			oldAsset0Denom := twapRecord.Asset0Denom
			oldAsset1Denom := twapRecord.Asset1Denom
			oldSpotPrice0 := twapRecord.P0LastSpotPrice
			oldSpotPrice1 := twapRecord.P1LastSpotPrice

			twapRecord.Asset0Denom = oldAsset1Denom
			twapRecord.Asset1Denom = oldAsset0Denom
			twapRecord.P0LastSpotPrice = oldSpotPrice1
			twapRecord.P1LastSpotPrice = oldSpotPrice0
			keepers.TwapKeeper.StoreNewRecord(ctx, oldAsset0Denom, oldAsset1Denom, twapRecord)
		}

		newRecord, err := keepers.TwapKeeper.GetAllMostRecentRecordsForPool(ctx, poolId)
		if err != nil {
			return err
		}

		fmt.Println("AFTER FLIP", newRecord)
	}
	return nil
}
