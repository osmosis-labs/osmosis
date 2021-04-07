package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	ModuleName = "farm"

	StoreKey = ModuleName

	RouterKey = ModuleName

	QuerierRoute = ModuleName
)

var (
	GlobalFarmNumber = []byte("global_farm_number")

	FarmPrefix            = []byte{0x01}
	FarmerPrefix          = []byte{0x02}
	HistoricalEntryPrefix = []byte{0x03}
)

func GetFarmStoreKey(farmId uint64) []byte {
	return append(FarmPrefix, sdk.Uint64ToBigEndian(farmId)...)
}

func GetHistoricalEntryKey(farmId uint64, period uint64) []byte {
	return append(append(HistoricalEntryPrefix, sdk.Uint64ToBigEndian(farmId)...), sdk.Uint64ToBigEndian(period)...)
}

func GetFarmerStoreKey(farmId uint64, address sdk.AccAddress) []byte {
	return append(append(FarmerPrefix, sdk.Uint64ToBigEndian(farmId)...), address.Bytes()...)
}
