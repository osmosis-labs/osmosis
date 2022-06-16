package v10

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v9/app/keepers"
)

func RunForkLogic(ctx sdk.Context, appKeepers *keepers.AppKeepers) {
	for i := 0; i < 100; i++ {
		fmt.Println("Switching to v10 code")
	}
}
