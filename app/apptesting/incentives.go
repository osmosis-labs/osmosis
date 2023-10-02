package apptesting

// validates that Group and group Gauge exist
func (s *KeeperTestHelper) ValidateGroupExists(gaugeID uint64) {
	_, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, gaugeID)
	s.Require().NoError(err)

	_, err = s.App.IncentivesKeeper.GetGroupByGaugeID(s.Ctx, gaugeID)
	s.Require().NoError(err)
}
