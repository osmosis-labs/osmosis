package valsetprefcli_test

import (
	gocontext "context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	valPref "github.com/osmosis-labs/osmosis/v27/x/valset-pref"
	"github.com/osmosis-labs/osmosis/v27/x/valset-pref/client/queryproto"
	"github.com/osmosis-labs/osmosis/v27/x/valset-pref/types"
)

type QueryTestSuite struct {
	apptesting.KeeperTestHelper
	queryClient queryproto.QueryClient
}

func (s *QueryTestSuite) SetupSuite() {
	s.Setup()
	s.queryClient = queryproto.NewQueryClient(s.QueryHelper)

	// setup 3 validators
	valAddrs := s.SetupMultipleValidators(3)
	delegator := sdk.AccAddress([]byte("addr1---------------"))
	preferences := []types.ValidatorPreference{
		{
			ValOperAddress: valAddrs[0],
			Weight:         osmomath.NewDecWithPrec(5, 1),
		},
		{
			ValOperAddress: valAddrs[1],
			Weight:         osmomath.NewDecWithPrec(3, 1),
		},
		{
			ValOperAddress: valAddrs[2],
			Weight:         osmomath.NewDecWithPrec(2, 1),
		},
	}

	// setup message server
	msgServer := valPref.NewMsgServerImpl(s.App.ValidatorSetPreferenceKeeper)
	c := s.Ctx

	// call the create validator set preference
	_, err := msgServer.SetValidatorSetPreference(c, types.NewMsgSetValidatorSetPreference(delegator, preferences))
	s.Require().NoError(err)

	// creates a test context like blockheader, blockheight and more
	s.Commit()
}

func (s *QueryTestSuite) TestQueriesNeverAlterState() {
	testCases := []struct {
		name   string
		query  string
		input  interface{}
		output interface{}
	}{
		{
			"Query delegators validator set",
			"/osmosis.valsetpref.v1beta1.Query/UserValidatorPreferences",
			&queryproto.UserValidatorPreferencesRequest{Address: sdk.AccAddress([]byte("addr1---------------")).String()},
			&queryproto.UserValidatorPreferencesResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.SetupSuite()

		s.Run(tc.name, func() {
			err := s.QueryHelper.Invoke(gocontext.Background(), tc.query, tc.input, tc.output)
			s.Require().NoError(err)
			s.StateNotAltered()
		})
	}
}

func TestQueryTestSuite(t *testing.T) {
	suite.Run(t, new(QueryTestSuite))
}
