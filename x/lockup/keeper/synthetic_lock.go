package keeper

// func (k Keeper) DeleteAllMaturedSyntheticLocks(ctx sdk.Context) {
// 	iterator := k.iteratorBeforeTime(ctx, combineKeys(types.KeyPrefixSyntheticLockTimestamp), ctx.BlockTime())
// 	defer iterator.Close()
// 	for ; iterator.Valid(); iterator.Next() {
// 		synthLock := types.SyntheticLock{}
// 		err := proto.Unmarshal(iterator.Value(), &synthLock)
// 		if err != nil {
// 			panic(err)
// 		}
// 		err = k.DeleteSyntheticLockup(ctx, synthLock.UnderlyingLockId, synthLock.SynthDenom)
// 		if err != nil {
// 			// TODO: When underlying lock is deleted for a reason while synthetic lockup exists, panic could happen
// 			panic(err)
// 		}
// 	}
// }
