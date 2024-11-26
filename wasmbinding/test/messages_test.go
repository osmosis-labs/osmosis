package wasmbinding

import (
	"fmt"
	"os"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	"github.com/osmosis-labs/osmosis/v27/wasmbinding"
	"github.com/osmosis-labs/osmosis/v27/wasmbinding/bindings"

	"github.com/stretchr/testify/require"
)

func TestCreateDenom(t *testing.T) {
	apptesting.SkipIfWSL(t)
	actor := RandomAccountAddress()
	osmosis, ctx, homeDir := SetupCustomApp(t, actor)
	defer os.RemoveAll(homeDir)

	specs := map[string]struct {
		createDenom *bindings.CreateDenom
		expErr      bool
	}{
		"valid sub-denom": {
			createDenom: &bindings.CreateDenom{
				Subdenom: "MOON",
			},
		},
		// UNFORKINGNOTE: store now panics when attempting to search for nil key on bank keeper
		// "empty sub-denom": {
		// 	createDenom: &bindings.CreateDenom{
		// 		Subdenom: "",
		// 	},
		// 	expErr: false,
		// },
		"invalid sub-denom": {
			createDenom: &bindings.CreateDenom{
				Subdenom: "subdenom2!",
			},
			expErr: true,
		},
		"null create denom": {
			createDenom: nil,
			expErr:      true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			// when
			gotErr := wasmbinding.PerformCreateDenom(osmosis.TokenFactoryKeeper, osmosis.BankKeeper, ctx, actor, spec.createDenom)
			// then
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
		})
	}
}

func TestChangeAdmin(t *testing.T) {
	apptesting.SkipIfWSL(t)
	const validDenom = "validdenom"

	tokenCreator := RandomAccountAddress()

	specs := map[string]struct {
		actor       sdk.AccAddress
		changeAdmin *bindings.ChangeAdmin

		expErrMsg string
	}{
		"valid": {
			changeAdmin: &bindings.ChangeAdmin{
				Denom:           fmt.Sprintf("factory/%s/%s", tokenCreator.String(), validDenom),
				NewAdminAddress: RandomBech32AccountAddress(),
			},
			actor: tokenCreator,
		},
		"typo in factory in denom name": {
			changeAdmin: &bindings.ChangeAdmin{
				Denom:           fmt.Sprintf("facory/%s/%s", tokenCreator.String(), validDenom),
				NewAdminAddress: RandomBech32AccountAddress(),
			},
			actor:     tokenCreator,
			expErrMsg: "denom prefix is incorrect. Is: facory.  Should be: factory: invalid denom",
		},
		"invalid address in denom": {
			changeAdmin: &bindings.ChangeAdmin{
				Denom:           fmt.Sprintf("factory/%s/%s", RandomBech32AccountAddress(), validDenom),
				NewAdminAddress: RandomBech32AccountAddress(),
			},
			actor:     tokenCreator,
			expErrMsg: "failed changing admin from message: unauthorized account",
		},
		"other denom name in 3 part name": {
			changeAdmin: &bindings.ChangeAdmin{
				Denom:           fmt.Sprintf("factory/%s/%s", tokenCreator.String(), "invalid denom"),
				NewAdminAddress: RandomBech32AccountAddress(),
			},
			actor:     tokenCreator,
			expErrMsg: fmt.Sprintf("invalid denom: factory/%s/invalid denom", tokenCreator.String()),
		},
		"empty denom": {
			changeAdmin: &bindings.ChangeAdmin{
				Denom:           "",
				NewAdminAddress: RandomBech32AccountAddress(),
			},
			actor:     tokenCreator,
			expErrMsg: "invalid denom: ",
		},
		"empty address": {
			changeAdmin: &bindings.ChangeAdmin{
				Denom:           fmt.Sprintf("factory/%s/%s", tokenCreator.String(), validDenom),
				NewAdminAddress: "",
			},
			actor:     tokenCreator,
			expErrMsg: "address from bech32: empty address string is not allowed",
		},
		"creator is a different address": {
			changeAdmin: &bindings.ChangeAdmin{
				Denom:           fmt.Sprintf("factory/%s/%s", tokenCreator.String(), validDenom),
				NewAdminAddress: RandomBech32AccountAddress(),
			},
			actor:     RandomAccountAddress(),
			expErrMsg: "failed changing admin from message: unauthorized account",
		},
		"change to the same address": {
			changeAdmin: &bindings.ChangeAdmin{
				Denom:           fmt.Sprintf("factory/%s/%s", tokenCreator.String(), validDenom),
				NewAdminAddress: tokenCreator.String(),
			},
			actor: tokenCreator,
		},
		"nil binding": {
			actor:     tokenCreator,
			expErrMsg: "invalid request: changeAdmin is nil - original request: ",
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			// Setup
			osmosis, ctx, homeDir := SetupCustomApp(t, tokenCreator)
			defer os.RemoveAll(homeDir)

			err := wasmbinding.PerformCreateDenom(osmosis.TokenFactoryKeeper, osmosis.BankKeeper, ctx, tokenCreator, &bindings.CreateDenom{
				Subdenom: validDenom,
			})
			require.NoError(t, err)

			err = wasmbinding.ChangeAdmin(osmosis.TokenFactoryKeeper, ctx, spec.actor, spec.changeAdmin)
			if len(spec.expErrMsg) > 0 {
				require.Error(t, err)
				actualErrMsg := err.Error()
				require.Equal(t, spec.expErrMsg, actualErrMsg)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestMint(t *testing.T) {
	apptesting.SkipIfWSL(t)
	creator := RandomAccountAddress()
	osmosis, ctx, homeDir := SetupCustomApp(t, creator)
	defer os.RemoveAll(homeDir)

	// Create denoms for valid mint tests
	validDenom := bindings.CreateDenom{
		Subdenom: "MOON",
	}
	err := wasmbinding.PerformCreateDenom(osmosis.TokenFactoryKeeper, osmosis.BankKeeper, ctx, creator, &validDenom)
	require.NoError(t, err)

	// UNFORKINGNOTE: store now panics when attempting to search for nil key on bank keeper
	// emptyDenom := bindings.CreateDenom{
	// 	Subdenom: "",
	// }
	// err = wasmbinding.PerformCreateDenom(osmosis.TokenFactoryKeeper, &osmosis.BankKeeper, ctx, creator, &emptyDenom)
	// require.NoError(t, err)

	validDenomStr := fmt.Sprintf("factory/%s/%s", creator.String(), validDenom.Subdenom)
	// UNFORKINGNOTE: store now panics when attempting to search for nil key on bank keeper
	// emptyDenomStr := fmt.Sprintf("factory/%s/%s", creator.String(), emptyDenom.Subdenom)

	lucky := RandomAccountAddress()

	// lucky was broke
	balances := osmosis.BankKeeper.GetAllBalances(ctx, lucky)
	require.Empty(t, balances)

	amount, ok := osmomath.NewIntFromString("8080")
	require.True(t, ok)

	specs := map[string]struct {
		mint   *bindings.MintTokens
		expErr bool
	}{
		"valid mint": {
			mint: &bindings.MintTokens{
				Denom:         validDenomStr,
				Amount:        amount,
				MintToAddress: lucky.String(),
			},
		},
		// UNFORKINGNOTE: store now panics when attempting to search for nil key on bank keeper
		// "empty sub-denom": {
		// 	mint: &bindings.MintTokens{
		// 		Denom:         emptyDenomStr,
		// 		Amount:        amount,
		// 		MintToAddress: lucky.String(),
		// 	},
		// 	expErr: false,
		// },
		"nonexistent sub-denom": {
			mint: &bindings.MintTokens{
				Denom:         fmt.Sprintf("factory/%s/%s", creator.String(), "SUN"),
				Amount:        amount,
				MintToAddress: lucky.String(),
			},
			expErr: true,
		},
		"invalid sub-denom": {
			mint: &bindings.MintTokens{
				Denom:         "subdenom2!",
				Amount:        amount,
				MintToAddress: lucky.String(),
			},
			expErr: true,
		},
		"zero amount": {
			mint: &bindings.MintTokens{
				Denom:         validDenomStr,
				Amount:        osmomath.ZeroInt(),
				MintToAddress: lucky.String(),
			},
			expErr: true,
		},
		"negative amount": {
			mint: &bindings.MintTokens{
				Denom:         validDenomStr,
				Amount:        amount.Neg(),
				MintToAddress: lucky.String(),
			},
			expErr: true,
		},
		"empty recipient": {
			mint: &bindings.MintTokens{
				Denom:         validDenomStr,
				Amount:        amount,
				MintToAddress: "",
			},
			expErr: true,
		},
		"invalid recipient": {
			mint: &bindings.MintTokens{
				Denom:         validDenomStr,
				Amount:        amount,
				MintToAddress: "invalid",
			},
			expErr: true,
		},
		"null mint": {
			mint:   nil,
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			// when
			gotErr := wasmbinding.PerformMint(osmosis.TokenFactoryKeeper, osmosis.BankKeeper, ctx, creator, spec.mint)
			// then
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
		})
	}
}

func TestBurn(t *testing.T) {
	apptesting.SkipIfWSL(t)
	creator := RandomAccountAddress()
	osmosis, ctx, homeDir := SetupCustomApp(t, creator)
	defer os.RemoveAll(homeDir)

	// Create denoms for valid burn tests
	validDenom := bindings.CreateDenom{
		Subdenom: "MOON",
	}
	err := wasmbinding.PerformCreateDenom(osmosis.TokenFactoryKeeper, osmosis.BankKeeper, ctx, creator, &validDenom)
	require.NoError(t, err)

	// UNFORKINGNOTE: store now panics when attempting to search for nil key on bank keeper
	// emptyDenom := bindings.CreateDenom{
	// 	Subdenom: "",
	// }
	// err = wasmbinding.PerformCreateDenom(osmosis.TokenFactoryKeeper, &osmosis.BankKeeper, ctx, creator, &emptyDenom)
	// require.NoError(t, err)

	lucky := RandomAccountAddress()

	// lucky was broke
	balances := osmosis.BankKeeper.GetAllBalances(ctx, lucky)
	require.Empty(t, balances)

	validDenomStr := fmt.Sprintf("factory/%s/%s", creator.String(), validDenom.Subdenom)
	// UNFORKINGNOTE: store now panics when attempting to search for nil key on bank keeper
	//emptyDenomStr := fmt.Sprintf("factory/%s/%s", creator.String(), emptyDenom.Subdenom)

	mintAmount, ok := osmomath.NewIntFromString("8080")
	require.True(t, ok)

	specs := map[string]struct {
		burn   *bindings.BurnTokens
		expErr bool
	}{
		"valid burn": {
			burn: &bindings.BurnTokens{
				Denom:           validDenomStr,
				Amount:          mintAmount,
				BurnFromAddress: creator.String(),
			},
			expErr: false,
		},
		"non admin address": {
			burn: &bindings.BurnTokens{
				Denom:           validDenomStr,
				Amount:          mintAmount,
				BurnFromAddress: lucky.String(),
			},
			expErr: false,
		},
		// UNFORKINGNOTE: store now panics when attempting to search for nil key on bank keeper
		// "empty sub-denom": {
		// 	burn: &bindings.BurnTokens{
		// 		Denom:           emptyDenomStr,
		// 		Amount:          mintAmount,
		// 		BurnFromAddress: creator.String(),
		// 	},
		// 	expErr: false,
		// },
		"invalid sub-denom": {
			burn: &bindings.BurnTokens{
				Denom:           "sub-denom_2",
				Amount:          mintAmount,
				BurnFromAddress: creator.String(),
			},
			expErr: true,
		},
		"non-minted denom": {
			burn: &bindings.BurnTokens{
				Denom:           fmt.Sprintf("factory/%s/%s", creator.String(), "SUN"),
				Amount:          mintAmount,
				BurnFromAddress: creator.String(),
			},
			expErr: true,
		},
		"zero amount": {
			burn: &bindings.BurnTokens{
				Denom:           validDenomStr,
				Amount:          osmomath.ZeroInt(),
				BurnFromAddress: creator.String(),
			},
			expErr: true,
		},
		"negative amount": {
			burn:   nil,
			expErr: true,
		},
		"null burn": {
			burn: &bindings.BurnTokens{
				Denom:           validDenomStr,
				Amount:          mintAmount.Neg(),
				BurnFromAddress: creator.String(),
			},
			expErr: true,
		},
	}

	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			// Mint valid denom str and empty denom string for burn test
			mintBinding := &bindings.MintTokens{
				Denom:         validDenomStr,
				Amount:        mintAmount,
				MintToAddress: creator.String(),
			}
			err := wasmbinding.PerformMint(osmosis.TokenFactoryKeeper, osmosis.BankKeeper, ctx, creator, mintBinding)
			require.NoError(t, err)

			// UNFORKINGNOTE: store now panics when attempting to search for nil key on bank keeper
			// emptyDenomMintBinding := &bindings.MintTokens{
			// 	Denom:         emptyDenomStr,
			// 	Amount:        mintAmount,
			// 	MintToAddress: creator.String(),
			// }
			// err = wasmbinding.PerformMint(osmosis.TokenFactoryKeeper, &osmosis.BankKeeper, ctx, creator, emptyDenomMintBinding)
			// require.NoError(t, err)

			// when
			gotErr := wasmbinding.PerformBurn(osmosis.TokenFactoryKeeper, ctx, creator, spec.burn)
			// then
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
		})
	}
}
