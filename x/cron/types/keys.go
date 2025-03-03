package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "cron"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

var (
	CronJobKeyPrefix = []byte{0x11}
	LastCronIDKey    = []byte{0x12}
)

func CronKey(cronID uint64) []byte {
	return append(CronJobKeyPrefix, sdk.Uint64ToBigEndian(cronID)...)
}
