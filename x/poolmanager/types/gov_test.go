package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v19/x/poolmanager/types"
)

func TestDenomPairTakerFeeProposalMarshalUnmarshal(t *testing.T) {
	records := []types.DenomPairTakerFee{
		{
			Denom0:   "uion",
			Denom1:   "uosmo",
			TakerFee: sdk.MustNewDecFromStr("0.0013"),
		},
		{
			Denom0:   "stake",
			Denom1:   "uosmo",
			TakerFee: sdk.MustNewDecFromStr("0.0016"),
		},
		{
			Denom0:   "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
			Denom1:   "uosmo",
			TakerFee: sdk.MustNewDecFromStr("0.0017"),
		},
	}

	tests := []struct {
		proposal *types.DenomPairTakerFeeProposal
	}{
		{ // empty title
			proposal: &types.DenomPairTakerFeeProposal{
				Title:       "",
				Description: "proposal to add denom pair taker fee records",
			},
		},
		{ // empty description
			proposal: &types.DenomPairTakerFeeProposal{
				Title:       "title",
				Description: "",
			},
		},
		{ // happy path
			proposal: &types.DenomPairTakerFeeProposal{
				Title:             "title",
				Description:       "proposal to add denom pair taker fee records",
				DenomPairTakerFee: records,
			},
		},
	}

	for _, test := range tests {
		bz, err := proto.Marshal(test.proposal)
		require.NoError(t, err)
		decoded := types.DenomPairTakerFeeProposal{}
		err = proto.Unmarshal(bz, &decoded)
		require.NoError(t, err)
		require.Equal(t, *test.proposal, decoded)
	}
}

func TestDenomPairTakerFeeProposal_ValidateBasic(t *testing.T) {
	baseRecord := types.DenomPairTakerFee{
		Denom0:   "uion",
		Denom1:   "uosmo",
		TakerFee: sdk.MustNewDecFromStr("0.0013"),
	}

	withSameDenom := func(record types.DenomPairTakerFee) types.DenomPairTakerFee {
		record.Denom1 = record.Denom0
		return record
	}

	withInvalidDenom0 := func(record types.DenomPairTakerFee) types.DenomPairTakerFee {
		record.Denom0 = "0"
		return record
	}

	withInvalidDenom1 := func(record types.DenomPairTakerFee) types.DenomPairTakerFee {
		record.Denom1 = "1"
		return record
	}

	withInvalidTakerFee := func(record types.DenomPairTakerFee) types.DenomPairTakerFee {
		record.TakerFee = sdk.MustNewDecFromStr("1.01")
		return record
	}

	withInvalidRecord := func(record types.DenomPairTakerFee) types.DenomPairTakerFee {
		record = types.DenomPairTakerFee{}
		return record
	}

	tests := []struct {
		name       string
		modifyFunc func(types.DenomPairTakerFee) types.DenomPairTakerFee
		expectPass bool
	}{
		{
			name:       "proper msg",
			modifyFunc: func(record types.DenomPairTakerFee) types.DenomPairTakerFee { return record },
			expectPass: true,
		},
		{
			name:       "invalid denom pair",
			modifyFunc: withSameDenom,
			expectPass: false,
		},
		{
			name:       "invalid denom0",
			modifyFunc: withInvalidDenom0,
			expectPass: false,
		},
		{
			name:       "invalid denom1",
			modifyFunc: withInvalidDenom1,
			expectPass: false,
		},
		{
			name:       "invalid taker fee",
			modifyFunc: withInvalidTakerFee,
			expectPass: false,
		},
		{
			name:       "invalid record",
			modifyFunc: withInvalidRecord,
			expectPass: false,
		},
	}

	for _, test := range tests {
		records := []types.DenomPairTakerFee{test.modifyFunc(baseRecord)}

		denomPairTakerFeeProposal := types.NewDenomPairTakerFeeProposal("title", "description", records)

		if test.expectPass {
			require.NoError(t, denomPairTakerFeeProposal.ValidateBasic(), "test: %v", test.name)
		} else {
			require.Error(t, denomPairTakerFeeProposal.ValidateBasic(), "test: %v", test.name)
		}
	}
}
