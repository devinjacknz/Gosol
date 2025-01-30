package repository

import "errors"

var (
	// ErrNotFound represents a not found error
	ErrNotFound = errors.New("record not found")

	// ErrInvalidInput represents an invalid input error
	ErrInvalidInput = errors.New("invalid input")

	// ErrDuplicateKey represents a duplicate key error
	ErrDuplicateKey = errors.New("duplicate key")

	// ErrDatabaseConnection represents a database connection error
	ErrDatabaseConnection = errors.New("database connection error")

	// ErrTransactionFailed represents a transaction failure
	ErrTransactionFailed = errors.New("transaction failed")
)
