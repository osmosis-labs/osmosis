package keeper_test

import (
	txKeeper "github.com/osmosis-labs/osmosis/v7/x/txfees/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestSybil() {
	suite.SetupTest(false)

	createSybil := func(after func(s txKeeper.Sybil) txKeeper.Sybil) txKeeper.Sybil {
		properSybil := txKeeper.Sybil{
			GasPrice: sdk.MustNewDecFromStr("0.01"),
			FeesPaid: sdk.NewCoin("test", sdk.NewInt(1)),
		}

		return after(properSybil)
	}

	sybilStruct := createSybil(func(s txKeeper.Sybil) txKeeper.Sybil {
		// do nothing
		return s
	})

	suite.Require().Equal(sdk.MustNewDecFromStr("0.01"), sybilStruct.GasPrice)
	suite.Require().Equal(sdk.NewCoin("test", sdk.NewInt(1)), sybilStruct.FeesPaid)

	tests := []struct {
		name       string
		gasPrice   sdk.Dec
		feesPaid   sdk.Coin
		expectPass bool
	}{
		{
			name:       "proper sybil",
			gasPrice:   sdk.MustNewDecFromStr("0.01"),
			feesPaid:   sdk.NewCoin("test", sdk.NewInt(100)),
			expectPass: true,
		},
		{
			name:       "sybil with zero fees paid",
			gasPrice:   sdk.MustNewDecFromStr("0.01"),
			feesPaid:   sdk.NewCoin("test", sdk.ZeroInt()),
			expectPass: true,
		},
		{
			name:       "sybil with zeo gas price",
			gasPrice:   sdk.ZeroDec(),
			feesPaid:   sdk.NewCoin("test", sdk.NewInt(100)),
			expectPass: true,
		},
		{
			name:       "sybil with zero gas price and zero fees paid",
			gasPrice:   sdk.ZeroDec(),
			feesPaid:   sdk.NewCoin("test", sdk.ZeroInt()),
			expectPass: true,
		},
		{
			name:       "sybil with negative gas price",
			gasPrice:   sdk.MustNewDecFromStr("-100.0"),
			feesPaid:   sdk.NewCoin("test", sdk.NewInt(100)),
			expectPass: false,
		},
	}

	for _, test := range tests {
		sybil, err := txKeeper.NewSybil(test.gasPrice, test.feesPaid)

		if test.expectPass {
			suite.Require().NoError(err, "test: %v", test.name)
			suite.Require().Equal(test.gasPrice, sybil.GasPrice, "test: %v", test.name)
			suite.Require().Equal(test.feesPaid, sybil.FeesPaid, "test: %v", test.name)

			// store original fees paid														= x
			fp := sybil.FeesPaid
			// check that feesPaid can be added to by adding 100 test coin using method	 	= x + 100
			sybil.AddToFeesPaid(sdk.NewCoin("test", sdk.NewInt(100)))
			// check that the difference is 100		 										= 100
			diff := sybil.FeesPaid.Sub(fp)
			suite.Require().Equal(diff, sdk.NewCoin("test", sdk.NewInt(100)))

		} else {
			suite.Require().Error(err, "test: %v", test.name)
		}
	}

}
