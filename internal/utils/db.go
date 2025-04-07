package utils

import (
	"context"
	"errors"
	"gorm.io/gorm"
)

// WithTransaction executes the given function within a transaction.
// If the function returns an error, the transaction is rolled back.
// Otherwise, the transaction is committed.
func WithTransaction(db *gorm.DB, fn func(tx *gorm.DB) error) error {
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r) // re-throw panic after rollback
		}
	}()

	if err := fn(tx); err != nil {
		rollbackErr := tx.Rollback().Error
		if rollbackErr != nil {
			// Return both errors
			return errors.Join(err, rollbackErr)
		}
		return err
	}

	return tx.Commit().Error
}

// WithTransactionContext is the same as WithTransaction but accepts a context
func WithTransactionContext(ctx context.Context, db *gorm.DB, fn func(tx *gorm.DB) error) error {
	tx := db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r) // re-throw panic after rollback
		}
	}()

	if err := fn(tx); err != nil {
		rollbackErr := tx.Rollback().Error
		if rollbackErr != nil {
			// Return both errors
			return errors.Join(err, rollbackErr)
		}
		return err
	}

	return tx.Commit().Error
}
