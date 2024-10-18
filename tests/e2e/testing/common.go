package e2eTesting

import (
	"fmt"
	"strconv"
	"strings"

	math "cosmossdk.io/math"
	wasmKeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetStringEventAttribute returns TX response event attribute string value by type and attribute key.
func GetStringEventAttribute(events []abci.Event, eventType, attrKey string) string {
	for _, event := range events {
		if event.Type != eventType {
			continue
		}

		for _, attr := range event.Attributes {
			if attr.Key != attrKey {
				continue
			}

			attrValue := attr.Value
			if valueUnquoted, err := strconv.Unquote(attr.Value); err == nil {
				attrValue = valueUnquoted
			}

			return attrValue
		}
	}

	return ""
}

// GenAccounts generates a list of accounts and private keys for them.
func GenAccounts(num uint) ([]sdk.AccAddress, []cryptotypes.PrivKey) {
	addrs := make([]sdk.AccAddress, 0, num)
	privKeys := make([]cryptotypes.PrivKey, 0, num)

	for i := 0; i < cap(addrs); i++ {
		privKey := secp256k1.GenPrivKey()

		addrs = append(addrs, sdk.AccAddress(privKey.PubKey().Address()))
		privKeys = append(privKeys, privKey)
	}

	return addrs, privKeys
}

// GenContractAddresses generates a list of contract addresses (codeID and instanceID are sequential).
func GenContractAddresses(num uint) []sdk.AccAddress {
	addrs := make([]sdk.AccAddress, 0, num)

	for i := 0; i < cap(addrs); i++ {
		contractAddr := wasmKeeper.BuildContractAddressClassic(uint64(i), uint64(i))
		addrs = append(addrs, contractAddr)
	}

	return addrs
}

// HumanizeCoins returns the sdk.Coins string representation with a number of decimals specified.
// 1123000stake -> 1.123stake with 6 decimals (3 numbers after the dot is hardcoded).
func HumanizeCoins(decimals uint8, coins ...sdk.Coin) string {
	baseDec := math.LegacyNewDecWithPrec(1, int64(decimals))

	strs := make([]string, 0, len(coins))
	for _, coin := range coins {
		amtDec := math.LegacyNewDecFromInt(coin.Amount).Mul(baseDec)
		amtFloat, _ := amtDec.Float64()

		strs = append(strs, fmt.Sprintf("%.03f%s", amtFloat, coin.Denom))
	}

	return strings.Join(strs, ",")
}

// HumanizeDecCoins returns the sdk.DecCoins string representation.
// 1000.123456789stake -> 1.123456stake with 3 decimals (6 numbers after the dot is hardcoded).
func HumanizeDecCoins(decimals uint8, coins ...sdk.DecCoin) string {
	baseDec := math.LegacyNewDecWithPrec(1, int64(decimals))

	strs := make([]string, 0, len(coins))
	for _, coin := range coins {
		amtDec := coin.Amount.Mul(baseDec)
		amtFloat, _ := amtDec.Float64()

		strs = append(strs, fmt.Sprintf("%.06f%s", amtFloat, coin.Denom))
	}

	return strings.Join(strs, ",")
}
