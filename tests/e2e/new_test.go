package e2e

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type E2ETest struct {
	name       string
	fundCoins  sdk.Coins
	logic      func(s *IntegrationTestSuite, t *E2ETest)
	walletAddr sdk.AccAddress
}
