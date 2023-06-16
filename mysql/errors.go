package mysql

import "errors"

var (
	// ErrMySQLTimeout is returned when the MySQL server can't acquire the lock in the specified timeout
	ErrMySQLTimeout = errors.New("(mysql) timeout while acquiring the lock")

	// ErrMySQLInternalError is returned when MySQL is returning a generic internal error
	ErrMySQLInternalError = errors.New("internal mysql error acquiring the lock")
)
