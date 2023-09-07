package osmoassert

var diffTypesErrorMessage = "cannot compare variables of different types"

func failNowIfNot(s testSuite, ok bool) {
	if !ok {
		s.Require().FailNow(diffTypesErrorMessage)
	}
}
