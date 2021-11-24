# AuthZ

The authz (message authorization) module allows users to authorize another account to send messages on their behalf. Certain authorizations such as the spending of another account's tokens, can be parameterized to constrain the permissions of the grantee (like setting a spending limit).

## Message Types

### MsgGrantAuthorization

```go
// MsgGrantAuthorization grants the provided authorization to the grantee on the granter's
// account during the provided period time
type MsgGrantAuthorization struct {
	Granter       sdk.AccAddress `json:"granter"`
	Grantee       sdk.AccAddress `json:"grantee"`
	Authorization Authorization  `json:"authorization"`
	Period        time.Duration  `json:"period"`
}
```

### MsgRevokeAuthorization

```go
// MsgRevokeAuthorization revokes any authorization with the provided sdk.Msg type on the
// granter's account with that has been granted to the grantee
type MsgRevokeAuthorization struct {
	Granter sdk.AccAddress `json:"granter"`
	Grantee sdk.AccAddress `json:"grantee"`
	// AuthorizationMsgType is the type of sdk.Msg that the revoked Authorization refers to.
	// i.e. this is what `Authorization.MsgType()` returns
	AuthorizationMsgType string `json:"authorization_msg_type"`
}
```

### MsgExecAuthorized

```go
// MsgExecAuthorized attempts to execute the provided messages using
// authorizations granted to the grantee. Each message should have only
// one signer corresponding to the granter of the authorization.
type MsgExecAuthorized struct {
	Grantee sdk.AccAddress `json:"grantee"`
	Msgs    []sdk.Msg      `json:"msgs"`
}
```
