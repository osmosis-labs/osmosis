package keeper_test

func (suite *KeeperTestSuite) TestLockReferencesManagement() {

	key1 := []byte{0x11}
	key2 := []byte{0x12}

	suite.SetupTest()
	suite.app.LockupKeeper.AddLockRefByKey(suite.ctx, key1, 1)
	suite.app.LockupKeeper.AddLockRefByKey(suite.ctx, key2, 1)
	suite.app.LockupKeeper.AddLockRefByKey(suite.ctx, key1, 2)
	suite.app.LockupKeeper.AddLockRefByKey(suite.ctx, key2, 2)
	suite.app.LockupKeeper.AddLockRefByKey(suite.ctx, key2, 3)

	timeLock1 := suite.app.LockupKeeper.GetLockRefs(suite.ctx, key1)
	suite.Require().Equal(len(timeLock1.IDs), 2)
	timeLock2 := suite.app.LockupKeeper.GetLockRefs(suite.ctx, key2)
	suite.Require().Equal(len(timeLock2.IDs), 3)

	suite.app.LockupKeeper.DeleteLockRefByKey(suite.ctx, key2, 1)
	timeLock2 = suite.app.LockupKeeper.GetLockRefs(suite.ctx, key2)
	suite.Require().Equal(len(timeLock2.IDs), 2)
}
