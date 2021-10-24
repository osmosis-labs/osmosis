package keeper

// Shadow lockup spec
// - Shadow lockup uses same denom as prefix ({origin_denom}/staked_{validator_id})
// - Shadow lockup addition, deletion, state transition to unbonding should be called by external modules
// - Shadow lockup should follow the changes of native lockups
// - Shadow lockup has reference to native lockup ID
// - AccumulationStore should be managed for shadow lockups as another denom

// Scenario
// - Distribution module distribute rewards to shadow lockups using accumulation store I guess
// - If a user begin unlock the lockup, shadow lockup automatically move to unlocking lockup if exist.
// (Staking module or superfluid module should make following actions for this for voting power change etc.)
// - If unlock of the lockup finishes and lockup is deleted, shadow lockup should be deleted together. (Do it via hooks? or do directly?)
//// - Superfluid module create shadow lockup if a user want to use his lockup for superfluid staking
//// - Superfluid module start unbonding of shadow lockup if a user don't want to do superfluid staking
//// - Superfluid module add unbonding shadow lockup if the user redelegate to another validator
//// Shadow lockup could exist more than one per denom, and if suffix is same, only one could exist.
//// - Should be able to get native lockup ID from shadow and from native to shadows

func (k Keeper) setShadowLockup(lockID uint64, shadow string) {

}

func (k Keeper) GetShadowLockup(lockID uint64, shadow string) {

}

func (k Keeper) GetAllShadowsByLockup(lockID uint64) {

}

func (k Keeper) GetAllShadows(lockID uint64) {

}

// CreateShadowLockup create shadow of lockup with lock id and shadow(denom suffix)
func (k Keeper) CreateShadowLockup(lockID uint64, shadow string) {

}

// CreateShadowLockup delete shadow of lockup with lock id and shadow(denom suffix)
func (k Keeper) DeleteShadowLockup(lockID uint64, shadow string) {

}

// DeleteAllShadowByLockup delete all the shadows of lockup by id
func (k Keeper) DeleteAllShadowsByLockup(lockID uint64) {

}

// BeginUnbondShadowLockup begin unbonding for shadow lockup
func (k Keeper) BeginUnbondShadowLockup(lockID uint64, shadow string) {

}
