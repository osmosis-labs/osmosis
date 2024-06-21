package types

import (
	"bytes"
	"fmt"
)

const (
	ModuleName   = "tradingtiers"
	StoreKey     = ModuleName
	KeySeparator = "|"
)

var (
	AccountDailyOsmoVolumePrefix = []byte{0x01}

	AccountRollingWindowUSDVolumePrefix = []byte{0x02}

	OsmoUSDValuePrefix = []byte{0x03}

	AccountTierOptInPrefix = []byte{0x04}
)

func FormatAccountDailyOsmoVolumeKey(epochNum int64, addr string) []byte {
	var buffer bytes.Buffer
	fmt.Fprintf(&buffer, "%s%s%d%s%s", AccountDailyOsmoVolumePrefix, KeySeparator, epochNum, KeySeparator, addr)
	return buffer.Bytes()
}

func FormatAccountDailyOsmoVolumeDayOnly(epochNum int64) []byte {
	var buffer bytes.Buffer
	fmt.Fprintf(&buffer, "%s%s%d%s%s", AccountDailyOsmoVolumePrefix, KeySeparator, epochNum)
	return buffer.Bytes()
}

func FormatAccountRollingWindowUSDVolumeKey(addr, tier string) []byte {
	var buffer bytes.Buffer
	fmt.Fprintf(&buffer, "%s%s%s%s%s", AccountRollingWindowUSDVolumePrefix, KeySeparator, addr, KeySeparator, tier)
	return buffer.Bytes()
}

func FormatOsmoUSDValueKey(epochNum int64) []byte {
	var buffer bytes.Buffer
	fmt.Fprintf(&buffer, "%s%s%d", OsmoUSDValuePrefix, KeySeparator, epochNum)
	return buffer.Bytes()
}

func FormatAccountTierOptInKey(addr string) []byte {
	var buffer bytes.Buffer
	fmt.Fprintf(&buffer, "%s%s%s", AccountTierOptInPrefix, KeySeparator, addr)
	return buffer.Bytes()
}
