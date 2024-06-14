package types

import (
	"bytes"
	"fmt"
)

const (
	ModuleName   = "tradingtiers"
	KeySeparator = "|"
)

var (
	AccountDailyOsmoVolumePrefix = []byte{0x01}

	AccountRollingWindowUSDVolumePrefix = []byte{0x02}

	OsmoUSDValuePrefix = []byte{0x03}

	AccountTierOptInPrefix = []byte{0x04}
)

func FormatAccountDailyOsmoVolumeKey(day, addr string) []byte {
	var buffer bytes.Buffer
	fmt.Fprintf(&buffer, "%s%s%s%s%s", AccountDailyOsmoVolumePrefix, KeySeparator, day, KeySeparator, addr)
	return buffer.Bytes()
}

func FormatAccountRollingWindowUSDVolumeKey(addr, tier string) []byte {
	var buffer bytes.Buffer
	fmt.Fprintf(&buffer, "%s%s%s%s%s", AccountRollingWindowUSDVolumePrefix, KeySeparator, addr, KeySeparator, tier)
	return buffer.Bytes()
}

func FormatOsmoUSDValueKey(epochNum string) []byte {
	var buffer bytes.Buffer
	fmt.Fprintf(&buffer, "%s%s%s", OsmoUSDValuePrefix, KeySeparator, epochNum)
	return buffer.Bytes()
}

func FormatAccountTierOptInKey(addr string) []byte {
	var buffer bytes.Buffer
	fmt.Fprintf(&buffer, "%s%s%s", AccountTierOptInPrefix, KeySeparator, addr)
	return buffer.Bytes()
}
