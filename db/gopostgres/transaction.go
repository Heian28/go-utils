package gopostgres

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

// GoPostgresTransactionInterface is an interface that wraps the WithTransaction method.
// It is used to wrap the WithTransaction method in the trx struct.
type GoPostgresTransactionInterface interface {
	WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error
}

// trx is a struct that wraps the gorm.DB and implements the GoPostgresTransactionInterface interface.
// It is used to wrap the WithTransaction method in the trx struct.
type trx struct {
	db *gorm.DB
}

// begin is a method that returns a new gorm.DB with the context.
func (t *trx) begin(ctx context.Context) *gorm.DB {
	return t.db.WithContext(ctx).Begin()
}

// commit is a method that commits the transaction.
func (t *trx) commit(tx *gorm.DB) error {
	return tx.Commit().Error
}

// rollback is a method that rolls back the transaction.
func (t *trx) rollback(tx *gorm.DB) error {
	return tx.Rollback().Error
}

// WithTransaction is a method that wraps the WithTransaction method in the trx struct.
// It is used to wrap the WithTransaction method in the trx struct.
func (t *trx) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	tx := t.begin(ctx)

	defer func() {
		if r := recover(); r != nil {
			t.rollback(tx)
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		t.rollback(tx)
		return err
	}

	if err := t.commit(tx); err != nil {
		return err
	}

	return nil
}

var GoPostgresTransaction = &trx{}

// NewGoPostgresTransaction is a function that returns a new GoPostgresTransactionInterface.
// It is used to wrap the WithTransaction method in the trx struct.
func NewGoPostgresTransaction(db *gorm.DB) GoPostgresTransactionInterface {
	GoPostgresTransaction.db = db
	return GoPostgresTransaction
}

var mGoPostgresTransaction map[string]*trx = make(map[string]*trx)

// SetupPostgreTransaction is a function that sets up the transaction for a connection.
// It is used to set up the transaction for a connection.
func SetupPostgreTransaction(connections map[string]*gorm.DB) {
	for connection, db := range connections {
		mGoPostgresTransaction[connection] = &trx{db: db}
	}
}

// GetPostgreTransaction is a function that returns the transaction for a connection.
// It is used to get the transaction for a connection.
func GetPostgreTransaction(connectionName string) (*trx, error) {
	if conn, ok := mGoPostgresTransaction[connectionName]; ok {
		return conn, nil
	} else {
		return nil, errors.New("connection not found")
	}
}

// GoTransaction is a function that returns the transaction for a connection.
// It is used to get the transaction for a connection.
func GoTransaction(connectionName string) GoPostgresTransactionInterface {
	if len(connectionName) == 0 {
		return GoPostgresTransaction
	} else {
		if con, ok := mGoPostgresTransaction[connectionName]; ok {
			return con
		}
	}
	return nil
}
