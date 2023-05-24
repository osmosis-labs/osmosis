package apptesting

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func (s *KeeperTestHelper) OverrideErrorCheck(b *testing.B) {
	s.b = b
}

func (s *KeeperTestHelper) RNoError(err error) {
	if s.b == nil {
		s.Require().NoError(err)
	} else {
		require.NoError(s.b, err)
	}
}

func (s *KeeperTestHelper) RTrue(b bool) {
	if s.b == nil {
		s.Require().True(b)
	} else {
		require.True(s.b, b)
	}
}

func (s *KeeperTestHelper) RNotNil(v interface{}) {
	if s.b == nil {
		s.Require().NotNil(v)
	} else {
		require.NotNil(s.b, v)
	}
}
