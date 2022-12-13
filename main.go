package main

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func main() {
	valAddr, err := sdk.ValAddressFromBech32("cosmosvaloper1sjllsnramtg3ewxqwwrwjxfgc4n4ef9u2lcnj0")
	if err != nil {
		fmt.Println("ERROR: ", err)
		return
	}

	fmt.Println(valAddr)
}
