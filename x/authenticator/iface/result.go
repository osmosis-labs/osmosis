package iface

import errorsmod "cosmossdk.io/errors"

type AuthenticationResult interface {
	isAuthenticationResult()
	IsAuthenticated() bool
	IsRejected() bool
	IsAuthenticationFailed() bool // Named for clarity, as opposed to IsNotAuthenticated which is ambiguous
	Error() error
}

type authenticated struct{}
type notAuthenticated struct{}
type rejected struct {
	msg string // TODO: Should we include the message in the error?
	err error
}

func (a authenticated) isAuthenticationResult()      {}
func (a authenticated) IsAuthenticated() bool        { return true }
func (a authenticated) IsRejected() bool             { return false }
func (a authenticated) IsAuthenticationFailed() bool { return false }
func (a authenticated) Error() error                 { return nil }

func (n notAuthenticated) isAuthenticationResult()      {}
func (n notAuthenticated) IsAuthenticated() bool        { return false }
func (n notAuthenticated) IsRejected() bool             { return false }
func (n notAuthenticated) IsAuthenticationFailed() bool { return true } // Represents cases where authentication wasn't attempted or was bypassed
func (n notAuthenticated) Error() error                 { return nil }

func (r rejected) isAuthenticationResult()      {}
func (r rejected) IsAuthenticated() bool        { return false }
func (r rejected) IsRejected() bool             { return true }
func (r rejected) IsAuthenticationFailed() bool { return false }
func (r rejected) Error() error                 { return errorsmod.Wrap(r.err, r.msg) }

func Authenticated() AuthenticationResult {
	return authenticated{}
}

func NotAuthenticated() AuthenticationResult {
	return notAuthenticated{}
}

func Rejected(msg string, err error) AuthenticationResult {
	return rejected{msg: msg, err: err}
}

type ConfirmationResult interface {
	isConfirmationResult()
	IsConfirm() bool
	IsBlock() bool
	Error() error
}

type confirm struct{}
type block struct {
	err error
}

func (c confirm) isConfirmationResult() {}
func (c confirm) IsConfirm() bool       { return true }
func (c confirm) IsBlock() bool         { return false }
func (c confirm) Error() error          { return nil }

func (b block) isConfirmationResult() {}
func (b block) IsConfirm() bool       { return false }
func (b block) IsBlock() bool         { return true }
func (b block) Error() error          { return b.err }

func Confirm() ConfirmationResult {
	return confirm{}
}

func Block(err error) ConfirmationResult {
	return block{err}
}
