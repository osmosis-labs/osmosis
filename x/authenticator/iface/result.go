package iface

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
