package keeper_test

func (suite *KeeperTestSuite) TestLockReferencesManagement() {
	key1 := []byte{0x11}
	key2 := []byte{0x12}

	suite.SetupTest()
	_ = suite.App.LockupKeeper.AddLockRefByKey(suite.Ctx, key1, 1)
	_ = suite.App.LockupKeeper.AddLockRefByKey(suite.Ctx, key2, 1)
	_ = suite.App.LockupKeeper.AddLockRefByKey(suite.Ctx, key1, 2)
	_ = suite.App.LockupKeeper.AddLockRefByKey(suite.Ctx, key2, 2)
	_ = suite.App.LockupKeeper.AddLockRefByKey(suite.Ctx, key2, 3)

	lockIDs1 := suite.App.LockupKeeper.GetLockRefs(suite.Ctx, key1)
	suite.Require().Equal(len(lockIDs1), 2)
	lockIDs2 := suite.App.LockupKeeper.GetLockRefs(suite.Ctx, key2)
	suite.Require().Equal(len(lockIDs2), 3)

	suite.App.LockupKeeper.DeleteLockRefByKey(suite.Ctx, key2, 1)
	lockIDs2 = suite.App.LockupKeeper.GetLockRefs(suite.Ctx, key2)
	suite.Require().Equal(len(lockIDs2), 2)
}
