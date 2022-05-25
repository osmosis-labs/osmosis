# Token Factory

The tokenfactory module allows any account to create a new token with
the name `factory/{creator address}/{nonce}`. Because tokens are
namespaced by creator address, this allows token minting to be
permissionless, due to not needing to resolve name collisions. A single
account can create multiple denoms, by providing a unique nonce for each
created denom. Once a denom is created, the original creator is given
"admin" privileges over the asset. This allows them to:

- Mint their denom to any account
- Burn their denom from any account
- Create a transfer of their denom between any two accounts
- Change the admin In the future, more admin capabilities may be
    added. Admins can choose to share admin privileges with other
    accounts using the authz module. The `ChangeAdmin` functionality,
    allows changing the master admin account, or even setting it to
    `""`, meaning no account has admin privileges of the asset.


## Messages

### CreateDenom
- Creates a denom of `factory/{creator address}/{nonce}` given the denom creator address and the denom nonce. The case a denom has a slash in its nonce is handled within the module.
``` {.go}
message MsgCreateDenom {
  string sender = 1 [ (gogoproto.moretags) = "yaml:\"sender\"" ];
  string nonce = 2 [ (gogoproto.moretags) = "yaml:\"nonce\"" ];
}
```

**State Modifications:**
- Fund community pool with the denom creation fee from the creator address, set in `Params`
- Set `DenomMetaData` via bank keeper
- Set `AuthorityMetadata` for the given denom to store the admin for the created denom `factory/{creator address}/{nonce}`. Admin is automatically set as the Msg sender
- Add denom to the `CreatorPrefixStore`, where a state of denoms created per creator is kept

### Mint
- Minting of a specific denom is only allowed for the creator of the denom registered during `CreateDenom`
``` {.go}
message MsgMint {
  string sender = 1 [ (gogoproto.moretags) = "yaml:\"sender\"" ];
  cosmos.base.v1beta1.Coin amount = 2 [
    (gogoproto.moretags) = "yaml:\"amount\"",
    (gogoproto.nullable) = false
  ];
}
```

**State Modifications:**
- Saftey check the following
  - Check that the denom minting is created via `tokenfactory` module
  - Check that the sender of the message is the admin of the denom
- Mint designated amount of tokens for the denom via `bank` module



### Burn
- Burning of a specific denom is only allowed for the creator of the denom registered during `CreateDenom`
``` {.go}
message MsgBurn {
  string sender = 1 [ (gogoproto.moretags) = "yaml:\"sender\"" ];
  cosmos.base.v1beta1.Coin amount = 2 [
    (gogoproto.moretags) = "yaml:\"amount\"",
    (gogoproto.nullable) = false
  ];
}
```

**State Modifications:**
- Saftey check the following
  - Check that the denom minting is created via `tokenfactory` module
  - Check that the sender of the message is the admin of the denom
- Burn designated amount of tokens for the denom via `bank` module


### ChangeAdmin
- Burning of a specific denom is only allowed for the creator of the denom registered during `CreateDenom`
``` {.go}
message MsgChangeAdmin {
  string sender = 1 [ (gogoproto.moretags) = "yaml:\"sender\"" ];
  string denom = 2 [ (gogoproto.moretags) = "yaml:\"denom\"" ];
  string newAdmin = 3 [ (gogoproto.moretags) = "yaml:\"new_admin\"" ];
}
```

**State Modifications:**
- Check that sender of the message is the admin of denom
- Modify `AuthorityMetadata` state entry to change the admin of the denom