package types

import (
	"fmt"
)

const (
	ModuleName = "farm"

	StoreKey = ModuleName

	RouterKey = ModuleName

	QuerierRoute = ModuleName
)

var (
	GlobalFarmNumber = []byte("global_farm_number")
)

func GetFarmStoreKey(farmId uint64) []byte {
	return []byte(fmt.Sprintf("farm/%d", farmId))
}

func GetHistoricalRecord(farmId uint64, period int64) []byte {
	return []byte(fmt.Sprintf("farm/%d/records/%d", farmId, period))
}

func GetFarmerStoreKey(farmId uint64, address string) []byte {
	return []byte(fmt.Sprintf("farmer/%d/%s", farmId, address))
}
