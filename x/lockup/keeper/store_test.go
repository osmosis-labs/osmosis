package keeper_test

func (suite *KeeperTestSuite) TestLockReferencesManagement() {

	key1 := []byte{0x11}
	key2 := []byte{0x12}

	suite.SetupTest()
	_ = suite.app.LockupKeeper.AddLockRefByKey(suite.ctx, key1, 1)
	_ = suite.app.LockupKeeper.AddLockRefByKey(suite.ctx, key2, 1)
	_ = suite.app.LockupKeeper.AddLockRefByKey(suite.ctx, key1, 2)
	_ = suite.app.LockupKeeper.AddLockRefByKey(suite.ctx, key2, 2)
	_ = suite.app.LockupKeeper.AddLockRefByKey(suite.ctx, key2, 3)

	lockIDs1 := suite.app.LockupKeeper.GetLockRefs(suite.ctx, key1)
	suite.Require().Equal(len(lockIDs1), 2)
	lockIDs2 := suite.app.LockupKeeper.GetLockRefs(suite.ctx, key2)
	suite.Require().Equal(len(lockIDs2), 3)

	err := suite.app.LockupKeeper.DeleteLockRefByKey(suite.ctx, key2, 1)
	suite.Require().NoError(err)
	lockIDs2 = suite.app.LockupKeeper.GetLockRefs(suite.ctx, key2)
	suite.Require().Equal(len(lockIDs2), 2)
}
