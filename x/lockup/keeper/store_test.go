package keeper_test

func (s *KeeperTestSuite) TestLockReferencesManagement() {
	key1 := []byte{0x11}
	key2 := []byte{0x12}

	s.SetupTest()
	_ = s.App.LockupKeeper.AddLockRefByKey(s.Ctx, key1, 1)
	_ = s.App.LockupKeeper.AddLockRefByKey(s.Ctx, key2, 1)
	_ = s.App.LockupKeeper.AddLockRefByKey(s.Ctx, key1, 2)
	_ = s.App.LockupKeeper.AddLockRefByKey(s.Ctx, key2, 2)
	_ = s.App.LockupKeeper.AddLockRefByKey(s.Ctx, key2, 3)

	lockIDs1 := s.App.LockupKeeper.GetLockRefs(s.Ctx, key1)
	s.Require().Equal(len(lockIDs1), 2)
	lockIDs2 := s.App.LockupKeeper.GetLockRefs(s.Ctx, key2)
	s.Require().Equal(len(lockIDs2), 3)

	s.App.LockupKeeper.DeleteLockRefByKey(s.Ctx, key2, 1)
	lockIDs2 = s.App.LockupKeeper.GetLockRefs(s.Ctx, key2)
	s.Require().Equal(len(lockIDs2), 2)
}
