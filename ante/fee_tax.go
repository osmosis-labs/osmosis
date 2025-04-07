package ante

import (
	"github.com/osmosis-labs/osmosis/osmomath"
	"regexp"
	"strings"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authz "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	marketexported "github.com/osmosis-labs/osmosis/v27/x/market/exported"
	oracleexported "github.com/osmosis-labs/osmosis/v27/x/oracle/exported"
)

var IBCRegexp = regexp.MustCompile("^ibc/[a-fA-F0-9]{64}$")

func isIBCDenom(denom string) bool {
	return IBCRegexp.MatchString(strings.ToLower(denom))
}

// FilterMsgAndComputeTax computes the stability tax on messages.
func FilterMsgAndComputeTax(ctx sdk.Context, tk TreasuryKeeper, msgs ...sdk.Msg) sdk.Coins {
	taxes := sdk.Coins{}

	for _, msg := range msgs {
		switch msg := msg.(type) {
		case *banktypes.MsgSend:
			taxes = taxes.Add(computeTax(ctx, tk, msg.Amount)...)

		case *banktypes.MsgMultiSend:
			for _, input := range msg.Inputs {
				taxes = taxes.Add(computeTax(ctx, tk, input.Coins)...)
			}

		case *marketexported.MsgSwapSend:
			taxes = taxes.Add(computeTax(ctx, tk, sdk.NewCoins(msg.OfferCoin))...)

		case *wasmtypes.MsgInstantiateContract:
			taxes = taxes.Add(computeTax(ctx, tk, msg.Funds)...)

		case *wasmtypes.MsgInstantiateContract2:
			taxes = taxes.Add(computeTax(ctx, tk, msg.Funds)...)

		case *wasmtypes.MsgExecuteContract:
			taxes = taxes.Add(computeTax(ctx, tk, msg.Funds)...)

		case *authz.MsgExec:
			messages, err := msg.GetMessages()
			if err == nil {
				taxes = taxes.Add(FilterMsgAndComputeTax(ctx, tk, messages...)...)
			}
		}
	}

	return taxes
}

// computes the stability tax according to tax-rate and tax-cap
func computeTax(ctx sdk.Context, tk TreasuryKeeper, principal sdk.Coins) sdk.Coins {
	taxRate := tk.GetTaxRate(ctx)
	if taxRate.Equal(osmomath.ZeroDec()) {
		return sdk.Coins{}
	}

	taxes := sdk.Coins{}

	for _, coin := range principal {
		if coin.Denom == sdk.DefaultBondDenom {
			continue
		}

		if isIBCDenom(coin.Denom) {
			continue
		}

		taxDue := osmomath.NewDecFromInt(coin.Amount).Mul(taxRate).QuoInt64(100).TruncateInt()

		// If tax due is greater than the tax cap, cap!
		//taxCap := tk.GetTaxCap(ctx, coin.Denom)
		//if taxDue.GT(taxCap) {
		//	taxDue = taxCap
		//}

		if taxDue.Equal(osmomath.ZeroInt()) {
			continue
		}

		taxes = taxes.Add(sdk.NewCoin(coin.Denom, taxDue))
	}

	return taxes
}

func isOracleTx(msgs []sdk.Msg) bool {
	for _, msg := range msgs {
		switch msg.(type) {
		case *oracleexported.MsgAggregateExchangeRatePrevote:
			continue
		case *oracleexported.MsgAggregateExchangeRateVote:
			continue
		default:
			return false
		}
	}

	return true
}
