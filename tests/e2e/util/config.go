package util

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (	
	// common
	OsmoDenom     = "uosmo"
	StakeDenom    = "stake"
	IbcDenom      = "ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518"
	MinGasPrice   = "0.000"
	IbcSendAmount = 3300000000
	// chainA
	ChainAID      = "osmo-test-a"
	OsmoBalanceA  = 200000000000
	StakeBalanceA = 110000000000
	StakeAmountA  = 100000000000
	// chainB
	ChainBID      = "osmo-test-b"
	OsmoBalanceB  = 500000000000
	StakeBalanceB = 440000000000
	StakeAmountB  = 400000000000
)

var (
	StakeAmountIntA  = sdk.NewInt(StakeAmountA)
	StakeAmountCoinA = sdk.NewCoin(StakeDenom, StakeAmountIntA)
	StakeAmountIntB  = sdk.NewInt(StakeAmountB)
	StakeAmountCoinB = sdk.NewCoin(StakeDenom, StakeAmountIntB)
)
