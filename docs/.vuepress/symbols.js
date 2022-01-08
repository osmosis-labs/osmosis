function sdk(anchor) {
  return `/docs/data/sdk.html#${anchor}`;
}

function data(path, anchor) {
  return `/docs/data/${path}.html#${anchor}`;
}

function msg(path, anchor) {
  return `/docs/messages/${path}.html#${anchor}`;
}

function py(type) {
  return `https://docs.python.org/3/library/functions.html#${type}`;
}

const str_address = "https://docs.python.org/3/library/stdtypes.html#str";

module.exports = {
  // sdk
  Coin: sdk("coin"),
  Coins: sdk("coins"),
  Dec: sdk("decimal-numbers"),
  Timestamp: sdk("timestamp"),
  JiguBox: sdk("jigubox"),

  // account
  Account: data("account", "account"),
  LazyGradedVestingAccount: data("account", "account-with-vesting"),
  VestingScheduleEntry: data("account", "vesting-schedule-entry"),

  // delegations
  Delegation: data("delegations", "delegation"),
  UnbondingDelegation: data("delegations", "unbonding-delegation"),
  Redelegation: data("delegations", "redelegation"),

  // oracle
  ExchangeRateVote: data("oracle", "exchange-rate-vote"),
  ExchangeRatePrevote: data("oracle", "exchange-rate-prevote"),

  // validator
  Validator: data("validator", "validator"),
  Description: data("validator", "delegate-description"),
  Commission: data("validator", "commission"),
  CommissionRates: data("validator", "commission-rates"),

  // tx
  StdFee: data("transactions", "fee"),
  StdSignMsg: data("transactions", "sign-message"),
  StdSignature: data("transactions", "signature"),
  StdTx: data("transactions", "transaction"),
  TxInfo: data("transactions", "transaction-info"),
  TxBroadcastResult: data("transactions", "tx-broadcast-result"),

  // block
  Block: data("blocks", "block"),

  // proposal:
  Proposal: data("proposals", "proposal"),
  ParamChanges: data("proposals", "parameter-changes"),

  // treasury
  PolicyConstraints: data("treasury", "policy-constraints"),

  // messages
  StdMsg: "/docs/messages/#stdmsg",
  MsgInfo: "/docs/messages/#msginfo",
  MsgSend: msg("bank", "send"),
  MsgMultiSend: msg("bank", "multisend"),

  // query
  EventsQuery: "/docs/query/event",
  MsgInfosQuery: "/docs/query/msginfo",
  TxInfosQuery: "/docs/query/txinfo",

  // builtins
  int: py("int"),
  bool: py("bool"),
  str: str_address,
  float: py("float"),

  // aliases
  AccAddress: str_address,
  ValAddress: str_address,
  ValConsAddress: str_address,
  Denom: str_address
};
