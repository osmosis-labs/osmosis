// Note:
// the contents of this file are only kept due to circular dependencies.
// In a future upgrade, we should remove this.
package repository

import (
	"context"
)

// Tx defines an interface for atomic transaction.
type Tx interface {
	// Exec executes the transaction.
	// Returns an error if transaction is not in progress.
	Exec(context.Context) error

	// IsActive returns true if transaction is in progress.
	IsActive() bool

	// ClearAll clears all data. Returns an error if any.
	ClearAll(ctx context.Context) error
}

// TxManager defines an interface for atomic transaction manager.
type TxManager interface {
	// StartTx starts a new atomic transaction.
	StartTx() Tx
}
