package recoverylock_test

// App level tests of the recovery lock with real Osmosis keepers, the real
// ante chain and the real message routers. Test keys: private key =
// sha256(seed), matching listed.txt and params.json (test values).

import (
	"crypto/sha256"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/suite"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/keeper"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v32/app/apptesting"
	"github.com/osmosis-labs/osmosis/v32/app/recoverylock/sdk50"
	cltypes "github.com/osmosis-labs/osmosis/v32/x/concentrated-liquidity/types"
	gammtypes "github.com/osmosis-labs/osmosis/v32/x/gamm/types"
	lockupkeeper "github.com/osmosis-labs/osmosis/v32/x/lockup/keeper"
	lockuptypes "github.com/osmosis-labs/osmosis/v32/x/lockup/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v32/x/poolmanager/types"
	smartaccounttypes "github.com/osmosis-labs/osmosis/v32/x/smart-account/types"
	tftypes "github.com/osmosis-labs/osmosis/v32/x/tokenfactory/types"
)

const (
	act        = int64(1_000_000_000)
	expiryUnix = int64(4_102_444_800)
)

func key(seed string) *secp256k1.PrivKey {
	h := sha256.Sum256([]byte(seed))
	return &secp256k1.PrivKey{Key: h[:]}
}

func addrOf(k *secp256k1.PrivKey) sdk.AccAddress { return sdk.AccAddress(k.PubKey().Address()) }

type LockSuite struct {
	apptesting.KeeperTestHelper
	kL, kL2, kP, kP7 *secp256k1.PrivKey
	L, L2, R, P, P7  sdk.AccAddress
	X                sdk.AccAddress
	lock             *sdk50.Lock
}

func TestLockSuite(t *testing.T) { suite.Run(t, new(LockSuite)) }

func (s *LockSuite) SetupTest() {
	s.Setup()
	s.kL, s.kL2 = key("recoverylock-test-listed-0"), key("recoverylock-test-listed-1")
	s.kP, s.kP7 = key("recoverylock-test-fee-payer-0"), key("recoverylock-test-fee-payer-7")
	s.L, s.L2, s.P, s.P7 = addrOf(s.kL), addrOf(s.kL2), addrOf(s.kP), addrOf(s.kP7)
	s.R = addrOf(key("recoverylock-test-recovery"))
	s.X = s.TestAccs[1]
	s.lock = s.App.AppKeepers.RecoveryLock
	s.Require().True(s.lock.Cfg.IsListed(s.L) && s.lock.Cfg.IsListed(s.L2))
	s.Require().True(s.lock.Cfg.IsFeePayer(s.P) && s.lock.Cfg.IsFeePayer(s.P7))
	s.Require().True(s.lock.Cfg.IsRecovery(s.R))
	big := sdk.NewCoins(sdk.NewCoin("uosmo", math.NewInt(100_000_000_000)), sdk.NewCoin("stake", math.NewInt(100_000_000_000)),
		sdk.NewCoin("foo", math.NewInt(100_000_000_000)), sdk.NewCoin("bar", math.NewInt(100_000_000_000)), sdk.NewCoin("baz", math.NewInt(100_000_000_000)))
	for _, a := range []sdk.AccAddress{s.L, s.L2, s.P, s.P7, s.X} {
		s.FundAcc(a, big)
	}
	s.Ctx = s.Ctx.WithBlockHeight(act - 10).WithBlockTime(time.Unix(expiryUnix-30*24*3600, 0))
	base, err := s.App.TxFeesKeeper.GetBaseDenom(s.Ctx)
	s.Require().NoError(err)
	fee = coins(base, 200_000)
	s.FundAcc(s.L, coins(base, 100_000_000_000))
}

func (s *LockSuite) activate() { s.Ctx = s.Ctx.WithBlockHeight(act) }

func coins(d string, a int64) sdk.Coins { return sdk.NewCoins(sdk.NewCoin(d, math.NewInt(a))) }

func isLocked(err error) bool { return err != nil && errorsmod.IsOf(err, sdk50.ErrLocked) }

// handle runs a message through the app router (no ante), as a module path would.
func (s *LockSuite) handle(m sdk.Msg) error {
	c, write := s.Ctx.CacheContext()
	_, err := s.App.MsgServiceRouter().Handler(m)(c, m)
	if err == nil {
		write()
	}
	return err
}

// dispatch runs a message through the wrapped router, as authz, wasm or ICA would.
func (s *LockSuite) dispatch(m sdk.Msg) error {
	c, write := s.Ctx.CacheContext()
	_, err := sdk50.Router{Inner: s.App.MsgServiceRouter(), L: s.lock}.Handler(m)(c, m)
	if err == nil {
		write()
	}
	return err
}

// signedTx builds a signed tx. signers are the message signers in order;
// payer (if not nil) is set as fee payer and signs last.
func (s *LockSuite) signedTx(fee sdk.Coins, payer *secp256k1.PrivKey, smartAccount bool, signers []*secp256k1.PrivKey, msgs ...sdk.Msg) sdk.Tx {
	cfg := s.App.GetTxConfig()
	b := cfg.NewTxBuilder()
	s.Require().NoError(b.SetMsgs(msgs...))
	b.SetFeeAmount(fee)
	b.SetGasLimit(2_000_000)
	all := append([]*secp256k1.PrivKey{}, signers...)
	if payer != nil {
		b.SetFeePayer(addrOf(payer))
		all = append(all, payer)
	}
	if smartAccount {
		sel := make([]uint64, len(msgs))
		for i := range sel {
			sel[i] = 1
		}
		any, err := codectypes.NewAnyWithValue(&smartaccounttypes.TxExtension{SelectedAuthenticators: sel})
		s.Require().NoError(err)
		b.(interface{ SetNonCriticalExtensionOptions(...*codectypes.Any) }).SetNonCriticalExtensionOptions(any)
	}
	mode := signing.SignMode_SIGN_MODE_DIRECT
	var sigs []signing.SignatureV2
	type sd struct {
		k   *secp256k1.PrivKey
		num uint64
		seq uint64
	}
	var data []sd
	for _, k := range all {
		acc := s.App.AccountKeeper.GetAccount(s.Ctx, addrOf(k))
		s.Require().NotNil(acc)
		data = append(data, sd{k, acc.GetAccountNumber(), acc.GetSequence()})
		sigs = append(sigs, signing.SignatureV2{PubKey: k.PubKey(), Data: &signing.SingleSignatureData{SignMode: mode}, Sequence: acc.GetSequence()})
	}
	s.Require().NoError(b.SetSignatures(sigs...))
	sigs = sigs[:0]
	for _, d := range data {
		sig, err := clienttx.SignWithPrivKey(s.Ctx, mode, authsigning.SignerData{
			ChainID: s.Ctx.ChainID(), AccountNumber: d.num, Sequence: d.seq, PubKey: d.k.PubKey(), Address: addrOf(d.k).String(),
		}, b, d.k, cfg, d.seq)
		s.Require().NoError(err)
		sigs = append(sigs, sig)
	}
	s.Require().NoError(b.SetSignatures(sigs...))
	return b.GetTx()
}

// deliver runs the real ante chain and then the messages, like block
// execution, and keeps the state only if everything succeeds.
func (s *LockSuite) deliver(tx sdk.Tx) error {
	c, write := s.Ctx.WithGasMeter(storetypes.NewInfiniteGasMeter()).CacheContext()
	c, err := s.App.GetBaseApp().AnteHandler()(c, tx, false)
	if err != nil {
		return err
	}
	for _, m := range tx.GetMsgs() {
		if _, err := s.App.MsgServiceRouter().Handler(m)(c, m); err != nil {
			return err
		}
	}
	write()
	return nil
}

var fee sdk.Coins

func (s *LockSuite) send(from, to sdk.AccAddress, c sdk.Coins) *banktypes.MsgSend {
	return &banktypes.MsgSend{FromAddress: from.String(), ToAddress: to.String(), Amount: c}
}

// smartTx is a transaction that selects authenticators (smart account path).
// The smart account ante requires the fee payer to be the signer of the first
// message, so the attacker signs the first message (and pays) and the listed
// wallet signs the second one: the listed signer is not the fee payer.
func (s *LockSuite) smartTx(attacker *secp256k1.PrivKey) sdk.Tx {
	return s.signedTx(fee, nil, true, []*secp256k1.PrivKey{attacker, s.kL},
		s.send(addrOf(attacker), s.X, coins("uosmo", 1)), s.send(s.L, s.R, coins("uosmo", 1)))
}

// ---------------------------------------------------------------------------

func (s *LockSuite) TestNoopBeforeActivation() {
	s.Ctx = s.Ctx.WithBlockHeight(act - 1)
	// coin rule
	r := s.lock.SendRestriction()
	g := storetypes.NewGasMeter(1_000_000)
	to, err := r(s.Ctx.WithGasMeter(g), s.L, s.X, coins("uosmo", 1))
	s.Require().NoError(err)
	s.Require().Equal(s.X, to)
	s.Require().Zero(g.GasConsumed())
	s.Require().NoError(s.handle(s.send(s.L, s.X, coins("uosmo", 1))))
	// staking guard via the real staking keeper
	val := s.SetupValidator(stakingtypes.Bonded)
	s.Require().NoError(s.handle(&stakingtypes.MsgDelegate{DelegatorAddress: s.L.String(), ValidatorAddress: val.String(), Amount: sdk.NewCoin("stake", math.NewInt(10))}))
	// wrapped router: same gas as the plain router
	m := s.send(s.L, s.X, coins("uosmo", 1))
	c1, _ := s.Ctx.WithGasMeter(storetypes.NewGasMeter(10_000_000)).CacheContext()
	c2, _ := s.Ctx.WithGasMeter(storetypes.NewGasMeter(10_000_000)).CacheContext()
	_, err = s.App.MsgServiceRouter().Handler(m)(c1, m)
	s.Require().NoError(err)
	_, err = sdk50.Router{Inner: s.App.MsgServiceRouter(), L: s.lock}.Handler(m)(c2, m)
	s.Require().NoError(err)
	s.Require().Equal(c1.GasMeter().GasConsumed(), c2.GasMeter().GasConsumed())
	// decorator: passes through without gas, even for the attacker paying
	g2 := storetypes.NewGasMeter(1_000_000)
	_, err = sdk50.Decorator{L: s.lock}.AnteHandle(s.Ctx.WithGasMeter(g2), s.signedTx(fee, nil, false, []*secp256k1.PrivKey{s.kL}, m), false,
		func(c sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) { return c, nil })
	s.Require().NoError(err)
	s.Require().Zero(g2.GasConsumed())
	// full tx: listed key pays its own fee and sends to the attacker
	s.Require().NoError(s.deliver(s.signedTx(fee, nil, false, []*secp256k1.PrivKey{s.kL}, s.send(s.L, s.X, coins("uosmo", 5)))))
}

func (s *LockSuite) TestCoinRuleAndStakingGuard() {
	poolId := s.PrepareBalancerPool()
	shares := s.App.BankKeeper.GetBalance(s.Ctx, s.TestAccs[0], gammtypes.GetPoolShareDenom(poolId))
	s.Require().NoError(s.App.BankKeeper.SendCoins(s.Ctx, s.TestAccs[0], s.L, sdk.NewCoins(sdk.NewCoin(shares.Denom, shares.Amount.QuoRaw(2)))))
	clPool := s.PrepareConcentratedPoolWithCoins("foo", "bar")
	tf := tftypes.NewMsgCreateDenom(s.X.String(), "evil")
	_, err := s.RunMsg(tf)
	s.Require().NoError(err)
	tfDenom := "factory/" + s.X.String() + "/evil"
	_, err = s.RunMsg(tftypes.NewMsgMintTo(s.X.String(), sdk.NewCoin(tfDenom, math.NewInt(1000)), s.L.String()))
	s.Require().NoError(err)
	val := s.SetupValidator(stakingtypes.Bonded)
	s.activate()

	blocked := map[string]sdk.Msg{
		"send to attacker":             s.send(s.L, s.X, coins("uosmo", 1)),
		"send gamm shares to attacker": s.send(s.L, s.X, coins(shares.Denom, 1)),
		"send to listed":               s.send(s.L, s.L2, coins("uosmo", 1)),
		"multi-send split": &banktypes.MsgMultiSend{Inputs: []banktypes.Input{{Address: s.L.String(), Coins: coins("uosmo", 2)}},
			Outputs: []banktypes.Output{{Address: s.R.String(), Coins: coins("uosmo", 1)}, {Address: s.X.String(), Coins: coins("uosmo", 1)}}},
		"tokenfactory burnFrom by attacker admin":      tftypes.NewMsgBurnFrom(s.X.String(), sdk.NewCoin(tfDenom, math.NewInt(10)), s.L.String()),
		"tokenfactory forceTransfer by attacker admin": tftypes.NewMsgForceTransfer(s.X.String(), sdk.NewCoin(tfDenom, math.NewInt(10)), s.L.String(), s.X.String()),
		"swap": &poolmanagertypes.MsgSwapExactAmountIn{Sender: s.L.String(), Routes: []poolmanagertypes.SwapAmountInRoute{{PoolId: poolId, TokenOutDenom: "bar"}},
			TokenIn: sdk.NewCoin("foo", math.NewInt(1000)), TokenOutMinAmount: math.OneInt()},
		"join pool":   &gammtypes.MsgJoinPool{Sender: s.L.String(), PoolId: poolId, ShareOutAmount: osmomath.NewInt(1_000_000_000_000), TokenInMaxs: nil},
		"exit pool":   &gammtypes.MsgExitPool{Sender: s.L.String(), PoolId: poolId, ShareInAmount: shares.Amount.QuoRaw(8), TokenOutMins: nil},
		"lock tokens": lockuptypes.NewMsgLockTokens(s.L, time.Hour, coins("uosmo", 1)),
		"create CL position": &cltypes.MsgCreatePosition{PoolId: clPool.GetId(), Sender: s.L.String(), LowerTick: -100, UpperTick: 100,
			TokensProvided: sdk.NewCoins(sdk.NewCoin("bar", math.NewInt(100)), sdk.NewCoin("foo", math.NewInt(100))), TokenMinAmount0: math.ZeroInt(), TokenMinAmount1: math.ZeroInt()},
		"delegate (staking guard)": &stakingtypes.MsgDelegate{DelegatorAddress: s.L.String(), ValidatorAddress: val.String(), Amount: sdk.NewCoin("stake", math.NewInt(1))},
	}
	names := make([]string, 0, len(blocked))
	for n := range blocked {
		names = append(names, n)
	}
	sort.Strings(names)
	// x/lockup MsgLockTokens re-wraps the keeper error as text
	// (errorsmod.Wrap(ErrInvalidRequest, err.Error()), x/lockup/keeper/
	// msg_server.go), which drops the error type; for it the lock error is
	// matched by its message.
	rewrapped := map[string]bool{"lock tokens": true}
	for _, n := range names {
		err := s.handle(blocked[n])
		if rewrapped[n] {
			s.Require().Error(err, n)
			s.Require().Contains(err.Error(), "may only go to the recovery address: "+sdk50.ErrLocked.Error(), "%s: want the lock error, got %v", n, err)
			continue
		}
		s.Require().True(isLocked(err), "%s: want the lock error, got %v", n, err)
	}
	s.Require().NoError(s.handle(s.send(s.L, s.R, coins(shares.Denom, 1))))
	s.Require().NoError(s.handle(s.send(s.L, s.R, coins(tfDenom, 1))))
	s.Require().NoError(s.handle(s.send(s.X, s.L, coins("uosmo", 1))), "incoming coins unaffected")
}

func (s *LockSuite) TestExitsEndToEnd() {
	val := s.SetupValidator(stakingtypes.Bonded)
	s.Require().NoError(s.handle(&stakingtypes.MsgDelegate{DelegatorAddress: s.L.String(), ValidatorAddress: val.String(), Amount: sdk.NewCoin("stake", math.NewInt(1_000_000))}))
	s.Require().NoError(s.App.DistrKeeper.SetDelegatorWithdrawAddr(s.Ctx, s.L, s.X))
	poolId := s.PrepareBalancerPool()
	shares := s.App.BankKeeper.GetBalance(s.Ctx, s.TestAccs[0], gammtypes.GetPoolShareDenom(poolId))
	s.Require().NoError(s.App.BankKeeper.SendCoins(s.Ctx, s.TestAccs[0], s.L, sdk.NewCoins(sdk.NewCoin(shares.Denom, shares.Amount.QuoRaw(2)))))
	lmsg := lockupkeeper.NewMsgServerImpl(s.App.LockupKeeper)
	lockAmt := sdk.NewCoin(shares.Denom, shares.Amount.QuoRaw(8))
	l1, err := lmsg.LockTokens(s.Ctx, lockuptypes.NewMsgLockTokens(s.L, 24*time.Hour, sdk.NewCoins(lockAmt)))
	s.Require().NoError(err)
	_, err = lmsg.LockTokens(s.Ctx, lockuptypes.NewMsgLockTokens(s.L, 48*time.Hour, sdk.NewCoins(lockAmt)))
	s.Require().NoError(err)
	s.Require().NoError(s.App.LockupKeeper.SetLockRewardReceiverAddress(s.Ctx, l1.ID, s.L, s.X.String()))
	clPool := s.PrepareConcentratedPoolWithCoins("foo", "bar")
	_, err = s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, clPool.GetId(), s.TestAccs[0], sdk.NewCoins(sdk.NewCoin("foo", math.NewInt(1_000_000)), sdk.NewCoin("bar", math.NewInt(1_000_000))))
	s.Require().NoError(err)
	posA, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, clPool.GetId(), s.L, sdk.NewCoins(sdk.NewCoin("foo", math.NewInt(1_000_000)), sdk.NewCoin("bar", math.NewInt(1_000_000))))
	s.Require().NoError(err)
	posB, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, clPool.GetId(), s.L, sdk.NewCoins(sdk.NewCoin("foo", math.NewInt(1_000_000)), sdk.NewCoin("bar", math.NewInt(1_000_000))))
	s.Require().NoError(err)
	posLocked, _, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionLocked(s.Ctx, clPool.GetId(), s.L, sdk.NewCoins(sdk.NewCoin("foo", math.NewInt(1_000_000)), sdk.NewCoin("bar", math.NewInt(1_000_000))), 24*time.Hour)
	s.Require().NoError(err)
	s.AllocateRewardsToValidator(val, math.NewInt(20_000))
	s.activate()

	tx := func(msgs ...sdk.Msg) error {
		return s.deliver(s.signedTx(fee, s.kP, false, []*secp256k1.PrivKey{s.kL}, msgs...))
	}
	// withdraw address is the attacker's: undelegate and reward withdrawal refused
	s.Require().True(isLocked(tx(&stakingtypes.MsgUndelegate{DelegatorAddress: s.L.String(), ValidatorAddress: val.String(), Amount: sdk.NewCoin("stake", math.NewInt(1000))})))
	s.Require().True(isLocked(tx(&distrtypes.MsgWithdrawDelegatorReward{DelegatorAddress: s.L.String(), ValidatorAddress: val.String()})))
	s.Require().True(isLocked(tx(&distrtypes.MsgSetWithdrawAddress{DelegatorAddress: s.L.String(), WithdrawAddress: s.X.String()})))
	// recovery operator fixes it, then exits work
	s.Require().NoError(tx(&distrtypes.MsgSetWithdrawAddress{DelegatorAddress: s.L.String(), WithdrawAddress: s.L.String()}))
	before := s.App.BankKeeper.GetBalance(s.Ctx, s.L, "stake")
	s.Require().NoError(tx(&distrtypes.MsgWithdrawDelegatorReward{DelegatorAddress: s.L.String(), ValidatorAddress: val.String()}))
	s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, s.L, "stake").Amount.GT(before.Amount), "rewards land in the wallet")
	s.Require().NoError(tx(&stakingtypes.MsgUndelegate{DelegatorAddress: s.L.String(), ValidatorAddress: val.String(), Amount: sdk.NewCoin("stake", math.NewInt(1_000_000))}))
	s.Require().NoError(tx(&distrtypes.MsgSetWithdrawAddress{DelegatorAddress: s.L.String(), WithdrawAddress: s.R.String()}))

	// unbonding completes into the wallet, where it stays locked
	ubd, err := s.App.StakingKeeper.GetUnbondingDelegation(s.Ctx, s.L, val)
	s.Require().NoError(err)
	s.Ctx = s.Ctx.WithBlockTime(ubd.Entries[0].CompletionTime.Add(time.Second))
	stakeBefore := s.App.BankKeeper.GetBalance(s.Ctx, s.L, "stake")
	_, err = s.App.StakingKeeper.CompleteUnbonding(s.Ctx, s.L, val)
	s.Require().NoError(err)
	s.Require().Equal(stakeBefore.Amount.Add(math.NewInt(1_000_000)), s.App.BankKeeper.GetBalance(s.Ctx, s.L, "stake").Amount)
	s.Require().True(isLocked(s.handle(s.send(s.L, s.X, coins("stake", 1_000_000)))))
	s.Require().NoError(tx(s.send(s.L, s.R, coins("stake", 1_000_000))))

	// a CL position with an active lock cannot be transferred (checked before
	// BeginUnlockingAll below, which also starts unlocking that lock)
	err = tx(&cltypes.MsgTransferPositions{PositionIds: []uint64{posLocked.ID}, Sender: s.L.String(), NewOwner: s.R.String()})
	s.Require().Error(err, "position with an active lock is refused by the CL keeper")
	s.Require().False(isLocked(err))
	// lock reward receiver
	s.Require().True(isLocked(tx(&lockuptypes.MsgSetRewardReceiverAddress{Owner: s.L.String(), LockID: l1.ID, RewardReceiver: s.X.String()})))
	s.Require().NoError(tx(&lockuptypes.MsgSetRewardReceiverAddress{Owner: s.L.String(), LockID: l1.ID, RewardReceiver: s.R.String()}))
	s.Require().NoError(tx(&lockuptypes.MsgSetRewardReceiverAddress{Owner: s.L.String(), LockID: l1.ID, RewardReceiver: s.L.String()}))
	// begin unlocking, then maturity pays the wallet
	s.Require().NoError(tx(&lockuptypes.MsgBeginUnlocking{Owner: s.L.String(), ID: l1.ID}))
	s.Require().NoError(tx(&lockuptypes.MsgBeginUnlockingAll{Owner: s.L.String()}))
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(72 * time.Hour))
	sharesBefore := s.App.BankKeeper.GetBalance(s.Ctx, s.L, shares.Denom)
	s.Require().NoError(s.App.LockupKeeper.UnlockMaturedLock(s.Ctx, l1.ID))
	s.Require().Equal(sharesBefore.Amount.Add(lockAmt.Amount), s.App.BankKeeper.GetBalance(s.Ctx, s.L, shares.Denom).Amount)
	s.Require().NoError(tx(s.send(s.L, s.R, sdk.NewCoins(lockAmt))))

	// concentrated liquidity
	s.Require().NoError(tx(&cltypes.MsgCollectSpreadRewards{PositionIds: []uint64{posA.ID}, Sender: s.L.String()}))
	s.Require().NoError(tx(&cltypes.MsgCollectIncentives{PositionIds: []uint64{posA.ID}, Sender: s.L.String()}))
	foo := s.App.BankKeeper.GetBalance(s.Ctx, s.L, "foo")
	s.Require().NoError(tx(&cltypes.MsgWithdrawPosition{PositionId: posA.ID, Sender: s.L.String(), LiquidityAmount: posA.Liquidity}))
	s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, s.L, "foo").Amount.GT(foo.Amount), "withdrawn liquidity lands in the wallet")
	s.Require().True(isLocked(tx(&cltypes.MsgTransferPositions{PositionIds: []uint64{posB.ID}, Sender: s.L.String(), NewOwner: s.X.String()})))
	s.Require().NoError(tx(&cltypes.MsgTransferPositions{PositionIds: []uint64{posB.ID}, Sender: s.L.String(), NewOwner: s.R.String()}))
	p, err := s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, posB.ID)
	s.Require().NoError(err)
	s.Require().Equal(s.R.String(), p.Address)
	// exit pool is not an allowed message
	s.Require().True(isLocked(tx(&gammtypes.MsgExitPool{Sender: s.L.String(), PoolId: poolId, ShareInAmount: shares.Amount.QuoRaw(16), TokenOutMins: nil})))
}

func (s *LockSuite) TestAnteFeesAndPaths() {
	s.activate()
	send := s.send(s.L, s.R, coins("uosmo", 1))
	kX := key("attacker-test-key")
	attackerKeyed := addrOf(kX)
	s.FundAcc(attackerKeyed, sdk.NewCoins(sdk.NewCoin("uosmo", math.NewInt(1_000_000_000)), fee[0].AddAmount(math.NewInt(1_000_000_000))))
	cases := []struct {
		name string
		tx   sdk.Tx
	}{
		{"listed key pays the whole balance as fee", s.signedTx(s.App.BankKeeper.GetAllBalances(s.Ctx, s.L), nil, false, []*secp256k1.PrivKey{s.kL}, s.send(s.L, s.R, coins("uosmo", 1)))},
		{"attacker pays fee for listed send to recovery", s.signedTx(fee, kX, false, []*secp256k1.PrivKey{s.kL}, send)},
		{"fee payer pays zero fee", s.signedTx(nil, s.kP, false, []*secp256k1.PrivKey{s.kL}, send)},
		{"listed is fee payer of an unrelated tx", s.signedTx(fee, s.kL, false, []*secp256k1.PrivKey{kX}, s.send(attackerKeyed, s.X, coins("uosmo", 1)))},
		{"listed send to attacker, fee payer pays", s.signedTx(fee, s.kP, false, []*secp256k1.PrivKey{s.kL}, s.send(s.L, s.X, coins("uosmo", 1)))},
		{"smart account path, listed signer, attacker pays", s.smartTx(kX)},
		{"authz exec by attacker", s.signedTx(fee, nil, false, []*secp256k1.PrivKey{kX}, func() sdk.Msg {
			m := authz.NewMsgExec(attackerKeyed, []sdk.Msg{s.send(s.L, s.R, coins("uosmo", 1))})
			return &m
		}())},
	}
	for _, c := range cases {
		balances := map[string]sdk.Coins{}
		for _, a := range []sdk.AccAddress{s.L, s.P, attackerKeyed} {
			balances[a.String()] = s.App.BankKeeper.GetAllBalances(s.Ctx, a)
		}
		cc, _ := s.Ctx.WithGasMeter(storetypes.NewInfiniteGasMeter()).CacheContext()
		_, err := s.App.GetBaseApp().AnteHandler()(cc, c.tx, false)
		s.Require().True(isLocked(err), "%s: want lock error, got %v", c.name, err)
		for _, a := range []sdk.AccAddress{s.L, s.P, attackerKeyed} {
			s.Require().Equal(balances[a.String()], s.App.BankKeeper.GetAllBalances(cc, a), "%s: no fee charged on reject", c.name)
		}
	}
	// allowed: any configured fee payer
	s.Require().NoError(s.deliver(s.signedTx(fee, s.kP7, false, []*secp256k1.PrivKey{s.kL}, s.send(s.L, s.R, coins("uosmo", 5)))))
	s.Require().NoError(s.deliver(s.signedTx(fee, s.kP, false, []*secp256k1.PrivKey{s.kL}, s.send(s.L, s.R, coins("uosmo", 5)))))
	// empty transaction: no panic, rejected by later decorators
	b := s.App.GetTxConfig().NewTxBuilder()
	b.SetGasLimit(100_000)
	b.SetFeeAmount(fee)
	s.Require().NotPanics(func() {
		cc, _ := s.Ctx.CacheContext()
		_, err := s.App.GetBaseApp().AnteHandler()(cc, b.GetTx(), false)
		s.Require().Error(err)
	})
}

// TestSmartAccountPathReached shows that the transaction used in
// TestAnteFeesAndPaths really goes down the smart account path: before
// activation it fails inside the authenticator flow (not with the lock
// error); after activation the lock rejects it first.
func (s *LockSuite) TestSmartAccountPathReached() {
	kX := key("attacker-test-key")
	s.FundAcc(addrOf(kX), sdk.NewCoins(fee[0].AddAmount(math.NewInt(1_000_000_000))))
	tx := s.smartTx(kX)
	cc, _ := s.Ctx.WithGasMeter(storetypes.NewInfiniteGasMeter()).CacheContext()
	_, err := s.App.GetBaseApp().AnteHandler()(cc, tx, false)
	s.Require().Error(err)
	s.Require().False(isLocked(err))
	s.T().Logf("before activation the smart account path rejects with: %v", err)
	// past the fee payer check, inside the authenticator selection
	s.Require().NotContains(err.Error(), "fee payer must be the first signer")
	s.Require().Contains(strings.ToLower(err.Error()), "authenticator", "not the smart account path: %v", err)
	s.activate()
	cc, _ = s.Ctx.WithGasMeter(storetypes.NewInfiniteGasMeter()).CacheContext()
	_, err = s.App.GetBaseApp().AnteHandler()(cc, tx, false)
	s.Require().True(isLocked(err), "after activation: %v", err)
}

// TestRedelegationSlashResetsWithdrawAddr: slashing a redelegation calls
// Unbond, which pays rewards to the withdraw address with no transaction
// involved. The lock's staking hook must reset an attacker-set withdraw
// address to the wallet first.
func (s *LockSuite) TestRedelegationSlashResetsWithdrawAddr() {
	v1 := s.SetupValidator(stakingtypes.Bonded)
	v2 := s.SetupValidator(stakingtypes.Bonded)
	s.Require().NoError(s.handle(&stakingtypes.MsgDelegate{DelegatorAddress: s.L.String(), ValidatorAddress: v1.String(), Amount: sdk.NewCoin("stake", math.NewInt(5_000_000))}))
	s.Require().NoError(s.App.DistrKeeper.SetDelegatorWithdrawAddr(s.Ctx, s.L, s.X))
	redHeight := s.Ctx.BlockHeight()
	s.Require().NoError(s.handle(&stakingtypes.MsgBeginRedelegate{DelegatorAddress: s.L.String(), ValidatorSrcAddress: v1.String(), ValidatorDstAddress: v2.String(), Amount: sdk.NewCoin("stake", math.NewInt(5_000_000))}))
	s.AllocateRewardsToValidator(v2, math.NewInt(20_000))
	s.activate()
	val1, err := s.App.StakingKeeper.GetValidator(s.Ctx, v1)
	s.Require().NoError(err)
	cons, err := val1.GetConsAddr()
	s.Require().NoError(err)
	before := s.App.BankKeeper.GetBalance(s.Ctx, s.X, "stake")
	walletBefore := s.App.BankKeeper.GetBalance(s.Ctx, s.L, "stake")
	_, err = s.App.StakingKeeper.Slash(s.Ctx, cons, redHeight, val1.ConsensusPower(sdk.DefaultPowerReduction), math.LegacyNewDecWithPrec(5, 2))
	s.Require().NoError(err)
	s.Require().Equal(before, s.App.BankKeeper.GetBalance(s.Ctx, s.X, "stake"), "attacker withdraw address must not receive rewards")
	w, err := s.App.DistrKeeper.GetDelegatorWithdrawAddr(s.Ctx, s.L)
	s.Require().NoError(err)
	s.Require().Equal(s.L, w, "withdraw address reset to the wallet")
	s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, s.L, "stake").Amount.GT(walletBefore.Amount), "rewards land in the wallet")
}

// TestRedelegationSlashBeforeActivation shows the hook is inert before
// activation: the same slash pays the attacker-set withdraw address, exactly
// as upstream does.
func (s *LockSuite) TestRedelegationSlashBeforeActivation() {
	v1 := s.SetupValidator(stakingtypes.Bonded)
	v2 := s.SetupValidator(stakingtypes.Bonded)
	s.Require().NoError(s.handle(&stakingtypes.MsgDelegate{DelegatorAddress: s.L.String(), ValidatorAddress: v1.String(), Amount: sdk.NewCoin("stake", math.NewInt(5_000_000))}))
	s.Require().NoError(s.App.DistrKeeper.SetDelegatorWithdrawAddr(s.Ctx, s.L, s.X))
	redHeight := s.Ctx.BlockHeight()
	s.Require().NoError(s.handle(&stakingtypes.MsgBeginRedelegate{DelegatorAddress: s.L.String(), ValidatorSrcAddress: v1.String(), ValidatorDstAddress: v2.String(), Amount: sdk.NewCoin("stake", math.NewInt(5_000_000))}))
	s.AllocateRewardsToValidator(v2, math.NewInt(20_000))
	val1, err := s.App.StakingKeeper.GetValidator(s.Ctx, v1)
	s.Require().NoError(err)
	cons, err := val1.GetConsAddr()
	s.Require().NoError(err)
	before := s.App.BankKeeper.GetBalance(s.Ctx, s.X, "stake")
	_, err = s.App.StakingKeeper.Slash(s.Ctx, cons, redHeight, val1.ConsensusPower(sdk.DefaultPowerReduction), math.LegacyNewDecWithPrec(5, 2))
	s.Require().NoError(err)
	s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, s.X, "stake").Amount.GT(before.Amount), "upstream behaviour before activation")
	w, err := s.App.DistrKeeper.GetDelegatorWithdrawAddr(s.Ctx, s.L)
	s.Require().NoError(err)
	s.Require().Equal(s.X, w, "withdraw address untouched before activation")
}

func (s *LockSuite) TestRoutersWired() {
	// the authz, wasm and ICA host keepers hold the wrapped router
	for name, k := range map[string]any{"authz": s.App.AuthzKeeper, "wasm": s.App.WasmKeeper, "icahost": s.App.ICAHostKeeper} {
		s.Require().True(findRouter(reflect.ValueOf(k), 0, map[uintptr]bool{}), "%s keeper does not hold sdk50.Router", name)
		s.Require().Zero(rawRouters(reflect.ValueOf(k), 0, map[uintptr]bool{}), "%s keeper reaches an unwrapped MsgServiceRouter", name)
	}
	val := s.SetupValidator(stakingtypes.Bonded)
	exp := s.Ctx.BlockTime().Add(1000 * time.Hour)
	for _, m := range []sdk.Msg{&banktypes.MsgSend{}, &distrtypes.MsgSetWithdrawAddress{}, &stakingtypes.MsgDelegate{}, &lockuptypes.MsgSetRewardReceiverAddress{}} {
		s.Require().NoError(s.App.AuthzKeeper.SaveGrant(s.Ctx, s.X, s.L, authz.NewGenericAuthorization(sdk.MsgTypeURL(m)), &exp))
	}
	s.activate()
	for _, m := range []sdk.Msg{
		s.send(s.L, s.X, coins("uosmo", 1)),
		&distrtypes.MsgSetWithdrawAddress{DelegatorAddress: s.L.String(), WithdrawAddress: s.X.String()},
		&stakingtypes.MsgDelegate{DelegatorAddress: s.L.String(), ValidatorAddress: val.String(), Amount: sdk.NewCoin("stake", math.NewInt(1))},
	} {
		c, _ := s.Ctx.CacheContext()
		_, err := s.App.AuthzKeeper.DispatchActions(c, s.X, []sdk.Msg{m})
		s.Require().True(isLocked(err), "authz dispatch %T: %v", m, err)
	}
	s.Require().True(isLocked(s.dispatch(&lockuptypes.MsgExtendLockup{Owner: s.L.String(), ID: 1, Duration: time.Hour})))
}

var (
	rawRouterType = reflect.TypeOf(&baseapp.MsgServiceRouter{})
	// Holders of the raw router that cannot dispatch a message signed by a
	// listed wallet, so the walk does not descend into them:
	//   - sdk50.Router: its Inner is the router it wraps.
	//   - baseapp.BaseApp: the transaction path, guarded by the ante
	//     decorator. The ICA host keeper reaches it as the upgrade keeper's
	//     version setter (ics4Wrapper -> channel -> client -> upgrade keeper).
	//   - gov keeper: dispatches only proposal messages whose single signer is
	//     the gov module account (x/gov/keeper/proposal.go SubmitProposal).
	//     Reached through the GRPC query router.
	//   - ICA controller keeper: dispatches only MsgChannelOpenInit signed by
	//     the ICA module account (controller/keeper/account.go
	//     registerInterchainAccount). Reached through the GRPC query router.
	stopTypes = map[reflect.Type]bool{
		reflect.TypeOf(sdk50.Router{}):               true,
		reflect.TypeOf(baseapp.BaseApp{}):            true,
		reflect.TypeOf(govkeeper.Keeper{}):           true,
		reflect.TypeOf(icacontrollerkeeper.Keeper{}): true,
	}
)

// rawRouters counts *baseapp.MsgServiceRouter values reachable from v
// (unexported fields included) without passing through one of stopTypes.
func rawRouters(v reflect.Value, depth int, seen map[uintptr]bool) int {
	if depth > 40 || !v.IsValid() {
		return 0
	}
	if stopTypes[v.Type()] {
		return 0
	}
	if v.Type() == rawRouterType {
		return 1
	}
	n := 0
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() || seen[v.Pointer()] {
			return 0
		}
		seen[v.Pointer()] = true
		return rawRouters(v.Elem(), depth+1, seen)
	case reflect.Interface:
		if !v.IsNil() {
			n += rawRouters(v.Elem(), depth+1, seen)
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			n += rawRouters(v.Index(i), depth+1, seen)
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			n += rawRouters(v.Field(i), depth+1, seen)
		}
	}
	return n
}

// findRouter walks a keeper's fields (unexported included, read only) for a
// value whose dynamic type is sdk50.Router.
func findRouter(v reflect.Value, depth int, seen map[uintptr]bool) bool {
	if depth > 24 || !v.IsValid() {
		return false
	}
	if v.Type() == reflect.TypeOf(sdk50.Router{}) {
		return true
	}
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() || seen[v.Pointer()] {
			return false
		}
		seen[v.Pointer()] = true
		return findRouter(v.Elem(), depth+1, seen)
	case reflect.Interface:
		if v.IsNil() {
			return false
		}
		return findRouter(v.Elem(), depth+1, seen)
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if findRouter(v.Index(i), depth+1, seen) {
				return true
			}
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if findRouter(v.Field(i), depth+1, seen) {
				return true
			}
		}
	}
	return false
}

func (s *LockSuite) TestExpiryEdge() {
	s.activate()
	m := s.send(s.L, s.X, coins("uosmo", 1))
	s.Ctx = s.Ctx.WithBlockTime(time.Unix(expiryUnix-1, 0))
	s.Require().True(isLocked(s.handle(m)))
	s.Ctx = s.Ctx.WithBlockTime(time.Unix(expiryUnix, 0))
	s.Require().NoError(s.handle(m))
}

// TestBypassCallSites pins, per file of every package linked into osmosisd,
// how often the bank functions that move coins out of an account without the
// send restriction (DelegateCoins, DelegateCoinsFromAccountToModule,
// SendManyCoins) or send from a module to many accounts
// (SendCoinsFromModuleToManyAccounts) are named, including method values,
// comments and interface declarations. Any change fails the test and must be checked by
// hand: the only call with a user address as source must stay x/staking
// (guarded by StakingBank). Files are identified by import path, not checkout
// path.
func TestBypassCallSites(t *testing.T) {
	goBin := filepath.Join(runtime.GOROOT(), "bin", "go")
	out, err := exec.Command(goBin, "list", "-tags", "recoverylock_test", "-deps", "-f", "{{.ImportPath}}|{{.Dir}}|{{join .GoFiles \",\"}}", "github.com/osmosis-labs/osmosis/v32/cmd/osmosisd").Output()
	if err != nil {
		t.Fatalf("go list failed: %v", err)
	}
	re := regexp.MustCompile(`\b(DelegateCoins|DelegateCoinsFromAccountToModule|SendManyCoins|SendCoinsFromModuleToManyAccounts)\b`)
	got := map[string]int{}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		parts := strings.SplitN(line, "|", 3)
		if len(parts) != 3 || parts[2] == "" {
			continue
		}
		for _, f := range strings.Split(parts[2], ",") {
			b, err := os.ReadFile(filepath.Join(parts[1], f))
			if err != nil {
				t.Fatalf("read %s: %v", f, err)
			}
			for _, m := range re.FindAllStringSubmatch(string(b), -1) {
				got[parts[0]+"/"+f+" "+m[1]]++
			}
		}
	}
	want := map[string]int{
		"github.com/cosmos/cosmos-sdk/x/bank/keeper/keeper.go DelegateCoins":                                           4,
		"github.com/cosmos/cosmos-sdk/x/bank/keeper/keeper.go DelegateCoinsFromAccountToModule":                        3,
		"github.com/cosmos/cosmos-sdk/x/bank/keeper/keeper.go SendCoinsFromModuleToManyAccounts":                       3,
		"github.com/cosmos/cosmos-sdk/x/bank/keeper/keeper.go SendManyCoins":                                           1,
		"github.com/cosmos/cosmos-sdk/x/bank/keeper/send.go SendManyCoins":                                             3,
		"github.com/cosmos/cosmos-sdk/x/staking/keeper/delegation.go DelegateCoinsFromAccountToModule":                 1,
		"github.com/cosmos/cosmos-sdk/x/staking/types/expected_keepers.go DelegateCoinsFromAccountToModule":            1,
		"github.com/osmosis-labs/osmosis/v32/app/recoverylock/sdk50/components.go DelegateCoins":                       1,
		"github.com/osmosis-labs/osmosis/v32/app/recoverylock/sdk50/components.go DelegateCoinsFromAccountToModule":    3,
		"github.com/osmosis-labs/osmosis/v32/x/incentives/keeper/distribute.go SendCoinsFromModuleToManyAccounts":      1,
		"github.com/osmosis-labs/osmosis/v32/x/incentives/types/expected_keepers.go SendCoinsFromModuleToManyAccounts": 1,
	}
	for k, n := range got {
		if want[k] != n {
			t.Errorf("unexpected: %s appears %d times (want %d)", k, n, want[k])
		}
	}
	for k, n := range want {
		if got[k] != n {
			t.Errorf("missing: %s appears %d times (want %d)", k, got[k], n)
		}
	}
}

// storeAndInstantiate stores a wasm file (as gov, like a permissioned
// upload) and instantiates it.
func (s *LockSuite) storeAndInstantiate(path, initMsg string) sdk.AccAddress {
	ck := wasmkeeper.NewGovPermissionKeeper(s.App.WasmKeeper)
	gov := authtypes.NewModuleAddress(govtypes.ModuleName)
	code, err := os.ReadFile(path)
	s.Require().NoError(err)
	id, _, err := ck.Create(s.Ctx, gov, code, &wasmtypes.AccessConfig{Permission: wasmtypes.AccessTypeEverybody})
	s.Require().NoError(err)
	addr, _, err := ck.Instantiate(s.Ctx, id, gov, gov, []byte(initMsg), "test", nil)
	s.Require().NoError(err)
	return addr
}

func (s *LockSuite) query(contract sdk.AccAddress, q string) string {
	out, err := s.App.WasmKeeper.QuerySmart(s.Ctx, contract, []byte(q))
	s.Require().NoError(err)
	return string(out)
}

func wexec(sender, contract sdk.AccAddress, msg string, funds sdk.Coins) *wasmtypes.MsgExecuteContract {
	return &wasmtypes.MsgExecuteContract{Sender: sender.String(), Contract: contract.String(), Msg: wasmtypes.RawContractMessage(msg), Funds: funds}
}

// TestContractExits runs real cw20-base 1.0.1 (tests/ibc-hooks/bytecode/cw20_base.wasm)
// and cw721-base v0.16.0 (testdata/cw721_base_v0.16.0.wasm, sha256
// d5fa68c6dcf210fd14e556a21d2f7a4976933da33caa97348ad99e0ec4373d9f) contracts.
// Only transfer and transfer_nft to the recovery address and the rights
// removing decrease_allowance, revoke and revoke_all pass, and only in a
// signed transaction paid by a fee payer. The router wrapper and authz
// dispatch refuse every message with a listed signer, allowed ones included.
func (s *LockSuite) TestContractExits() {
	cw20 := s.storeAndInstantiate("../../tests/ibc-hooks/bytecode/cw20_base.wasm",
		`{"name":"Test","symbol":"TST","decimals":6,"initial_balances":[{"address":"`+s.L.String()+`","amount":"1000"}]}`)
	cw721 := s.storeAndInstantiate("testdata/cw721_base_v0.16.0.wasm", `{"name":"N","symbol":"NFT","minter":"`+s.X.String()+`"}`)
	ck := wasmkeeper.NewDefaultPermissionKeeper(s.App.WasmKeeper)
	for _, id := range []string{"1", "2", "3", "4"} {
		_, err := ck.Execute(s.Ctx, cw721, s.X, []byte(`{"mint":{"token_id":"`+id+`","owner":"`+s.L.String()+`","token_uri":null,"extension":null}}`), nil)
		s.Require().NoError(err)
	}
	R, X := s.R.String(), s.X.String()
	// before activation the attacker grants itself rights over the wallet's
	// tokens (contract state, not covered by the coin rule)
	for _, m := range []struct {
		c   sdk.AccAddress
		msg string
	}{
		{cw20, `{"increase_allowance":{"spender":"` + X + `","amount":"300"}}`},
		{cw721, `{"approve":{"spender":"` + X + `","token_id":"4"}}`},
		{cw721, `{"approve_all":{"operator":"` + X + `"}}`},
	} {
		_, err := ck.Execute(s.Ctx, m.c, s.L, []byte(m.msg), nil)
		s.Require().NoError(err)
	}
	s.Require().Contains(s.query(cw20, `{"allowance":{"owner":"`+s.L.String()+`","spender":"`+X+`"}}`), `"allowance":"300"`)
	s.Require().Contains(s.query(cw721, `{"owner_of":{"token_id":"4"}}`), X)
	s.Require().Contains(s.query(cw721, `{"all_operators":{"owner":"`+s.L.String()+`"}}`), X)
	exp := s.Ctx.BlockTime().Add(1000 * time.Hour)
	s.Require().NoError(s.App.AuthzKeeper.SaveGrant(s.Ctx, s.X, s.L, authz.NewGenericAuthorization(sdk.MsgTypeURL(&wasmtypes.MsgExecuteContract{})), &exp))
	s.activate()
	tx := func(m sdk.Msg) error { return s.deliver(s.signedTx(fee, s.kP, false, []*secp256k1.PrivKey{s.kL}, m)) }

	refused := map[string]*wasmtypes.MsgExecuteContract{
		"cw20 transfer to attacker":          wexec(s.L, cw20, `{"transfer":{"recipient":"`+X+`","amount":"1"}}`, nil),
		"cw20 send":                          wexec(s.L, cw20, `{"send":{"contract":"`+R+`","amount":"1","msg":""}}`, nil),
		"cw20 increase allowance":            wexec(s.L, cw20, `{"increase_allowance":{"spender":"`+R+`","amount":"1"}}`, nil),
		"cw20 burn":                          wexec(s.L, cw20, `{"burn":{"amount":"1"}}`, nil),
		"cw20 transfer extra field":          wexec(s.L, cw20, `{"transfer":{"recipient":"`+R+`","amount":"1","memo":"x"}}`, nil),
		"cw20 transfer with funds":           wexec(s.L, cw20, `{"transfer":{"recipient":"`+R+`","amount":"1"}}`, coins("uosmo", 1)),
		"cw20 duplicate recipient key":       wexec(s.L, cw20, `{"transfer":{"recipient":"`+X+`","recipient":"`+R+`","amount":"1"}}`, nil),
		"cw20 duplicate top-level key":       wexec(s.L, cw20, `{"transfer":{"recipient":"`+R+`","amount":"1"},"transfer":{"recipient":"`+X+`","amount":"1"}}`, nil),
		"cw20 decrease_allowance expires":    wexec(s.L, cw20, `{"decrease_allowance":{"spender":"`+X+`","amount":"1","expires":{"never":{}}}}`, nil),
		"cw20 decrease_allowance bad addr":   wexec(s.L, cw20, `{"decrease_allowance":{"spender":"osmo1notanaddress","amount":"1"}}`, nil),
		"cw20 decrease_allowance with funds": wexec(s.L, cw20, `{"decrease_allowance":{"spender":"`+X+`","amount":"1"}}`, coins("uosmo", 1)),
		"cw721 transfer_nft to attacker":     wexec(s.L, cw721, `{"transfer_nft":{"recipient":"`+X+`","token_id":"1"}}`, nil),
		"cw721 send_nft":                     wexec(s.L, cw721, `{"send_nft":{"contract":"`+R+`","token_id":"1","msg":""}}`, nil),
		"cw721 approve":                      wexec(s.L, cw721, `{"approve":{"spender":"`+X+`","token_id":"1"}}`, nil),
		"cw721 approve_all":                  wexec(s.L, cw721, `{"approve_all":{"operator":"`+X+`"}}`, nil),
		"cw721 two keys":                     wexec(s.L, cw721, `{"transfer_nft":{"recipient":"`+R+`","token_id":"1"},"approve_all":{"operator":"`+X+`"}}`, nil),
		"cw721 extra field":                  wexec(s.L, cw721, `{"transfer_nft":{"recipient":"`+R+`","token_id":"1","x":1}}`, nil),
		"cw721 revoke extra field":           wexec(s.L, cw721, `{"revoke":{"spender":"`+X+`","token_id":"4","x":1}}`, nil),
		"cw721 revoke_all wrong field":       wexec(s.L, cw721, `{"revoke_all":{"spender":"`+X+`"}}`, nil),
	}
	// allowed in a fee payer transaction, refused through the router wrapper
	// and authz dispatch (no fee payer there)
	allowed := []*wasmtypes.MsgExecuteContract{
		wexec(s.L, cw20, `{"transfer":{"recipient":"`+R+`","amount":"400"}}`, nil),
		wexec(s.L, cw721, `{"transfer_nft":{"recipient":"`+R+`","token_id":"1"}}`, nil),
		wexec(s.L, cw20, `{"decrease_allowance":{"spender":"`+X+`","amount":"300"}}`, nil),
		wexec(s.L, cw721, `{"revoke":{"spender":"`+X+`","token_id":"4"}}`, nil),
		wexec(s.L, cw721, `{"revoke_all":{"operator":"`+X+`"}}`, nil),
	}
	names := make([]string, 0, len(refused))
	for n := range refused {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		m := refused[n]
		s.Require().True(isLocked(tx(m)), "tx %s", n)
	}
	for i, m := range append(append([]*wasmtypes.MsgExecuteContract{}, allowed...), refused["cw20 transfer to attacker"]) {
		s.Require().True(isLocked(s.dispatch(m)), "router %d", i)
		c, _ := s.Ctx.CacheContext()
		_, err := s.App.AuthzKeeper.DispatchActions(c, s.X, []sdk.Msg{m})
		s.Require().True(isLocked(err), "authz %d: %v", i, err)
	}
	s.Require().Contains(s.query(cw721, `{"owner_of":{"token_id":"1"}}`), s.L.String())
	s.Require().Equal(`{"balance":"1000"}`, s.query(cw20, `{"balance":{"address":"`+s.L.String()+`"}}`))
	s.Require().Contains(s.query(cw20, `{"allowance":{"owner":"`+s.L.String()+`","spender":"`+X+`"}}`), `"allowance":"300"`)

	for i, m := range allowed {
		s.Require().NoError(tx(m), "allowed %d", i)
	}
	s.Require().Equal(`{"balance":"400"}`, s.query(cw20, `{"balance":{"address":"`+R+`"}}`))
	s.Require().Equal(`{"balance":"600"}`, s.query(cw20, `{"balance":{"address":"`+s.L.String()+`"}}`))
	s.Require().Contains(s.query(cw721, `{"owner_of":{"token_id":"1"}}`), R)
	s.Require().Contains(s.query(cw20, `{"allowance":{"owner":"`+s.L.String()+`","spender":"`+X+`"}}`), `"allowance":"0"`)
	s.Require().NotContains(s.query(cw721, `{"owner_of":{"token_id":"4"}}`), X)
	s.Require().NotContains(s.query(cw721, `{"all_operators":{"owner":"`+s.L.String()+`"}}`), X)
	// the attacker's rights are gone: transfer_from and transfer_nft by X fail
	_, err := ck.Execute(s.Ctx, cw20, s.X, []byte(`{"transfer_from":{"owner":"`+s.L.String()+`","recipient":"`+X+`","amount":"1"}}`), nil)
	s.Require().Error(err)
	_, err = ck.Execute(s.Ctx, cw721, s.X, []byte(`{"transfer_nft":{"recipient":"`+X+`","token_id":"2"}}`), nil)
	s.Require().Error(err)
	s.Require().Contains(s.query(cw721, `{"owner_of":{"token_id":"2"}}`), s.L.String())
}
