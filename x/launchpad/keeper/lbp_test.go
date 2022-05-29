package keeper

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/x/launchpad/api"
)

func TestSuites(t *testing.T) {
	suite.Run(t, new(LBPSuite))
	suite.Run(t, new(TwoBuyersSuite))
}

type LBPSuite struct {
	suite.Suite
	treasury sdk.AccAddress
	accs     []sdk.AccAddress

	before, before2, start, end, after time.Time
}

func (s *LBPSuite) SetupSuite() {
	s.treasury = sdk.AccAddress([]byte("treasury"))
	s.accs = []sdk.AccAddress{
		[]byte("acc1"),
		[]byte("acc2"),
		[]byte("acc3"),
	}
	t0 := time.Unix(0, 0)
	s.before = t0
	s.before2 = t0.Add(api.ROUND)
	s.start = t0.Add(api.ROUND * 10)
	s.end = t0.Add(api.ROUND * 20)
	s.after = t0.Add(api.ROUND * 25)
}

func (s *LBPSuite) createLBP() *api.LBP {
	p := newLBP(s.treasury.String(), 1, "t_in", "t_out", s.start, s.end, sdk.NewInt(12_000))
	return &p
}

func (s *LBPSuite) TestNBuyers() {
	tcs := []struct {
		n int
	}{
		{1},
		{2},
	}
	for i, tc := range tcs {
		s.Run(fmt.Sprint("test: ", i), func() {
			s.testNBuyers(tc.n)
		})
	}
}

func (s *LBPSuite) testNBuyers(n int) {
	// require := s.Require()
	p := s.createLBP()
	users := make([]*api.UserPosition, n)
	stakedAmount := sdk.NewInt(24_000)
	// zero := sdk.ZeroInt()

	for i := 0; i < n; i++ {
		u := newUserPosition()
		users[i] = &u
		subscribe(p, &u, stakedAmount, s.before)
	}
}
