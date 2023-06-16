package mysql

import (
	"context"
	"database/sql"
	"dmutex"
	"fmt"
	"time"
)

type locker struct {
	conn *sql.Conn
}

// NewLocker creates a new locker.
func NewLocker(conn *sql.Conn) dmutex.Locker {
	return locker{
		conn: conn,
	}
}

// Release unlocks the lock.
func (l locker) Release(ctx context.Context, key string) error {
	row := l.conn.QueryRowContext(ctx, "SELECT RELEASE_LOCK(?)", key)
	if row.Err() != nil {
		return fmt.Errorf("%w: error releasing lock: %w", dmutex.ErrLockReleased, row.Err())
	}

	defer func() {
		if err := l.conn.Close(); err != nil {
			fmt.Printf("error closing connection: %v", err)
		}
	}()

	var result sql.NullInt32

	if err := row.Scan(&result); err != nil {
		return fmt.Errorf("%w: error scanning lock release result: %w", dmutex.ErrLockReleased, err)
	}

	if !result.Valid {
		return fmt.Errorf("%w: lock does not exist", dmutex.ErrLockReleased)
	}

	if result.Int32 == 0 {
		return fmt.Errorf("%w: lock was not established by this thread", dmutex.ErrLockReleased)
	}

	return nil
}

// Lock tries to acquire a lock with the given key in a period of time determined by the context.
// If the context is canceled, it will return ErrGetLockContextCancelled.
// If the lock is not acquired in the given time, it will return ErrMySQLTimeout.
// keys locked by Lock are not released when transactions commit or roll back.
// locks are automatically released when the session is terminated (either normally or abnormally).
func (l locker) Lock(ctx context.Context, key string) error {
	timeout := 1

	TimeTimeout, ok := ctx.Deadline()
	if ok {
		timeout = int(time.Until(TimeTimeout).Seconds())
	}

	row := l.conn.QueryRowContext(ctx, "SELECT GET_LOCK(?, ?)", key, timeout)

	var res sql.NullInt32
	err := row.Scan(&res)

	switch {
	case err != nil:
		// mysql error does not tell if it was due to context closing, checking it manually
		select {
		case <-ctx.Done():
			return dmutex.ErrGetLockContextCancelled
		default:
			rErr := l.Release(ctx, key)
			if rErr != nil {
				return fmt.Errorf("%w: error acquiring lock,could not release lock: %w", err, rErr)
			}
			break
		}

		return fmt.Errorf("could not read mysql response: %w", err)
	case !res.Valid:
		// Internal MySQL error occurred, such as out-of-memory, thread killed or others (the doc is not clear)
		// Note: some MySQL/MariaDB versions (like MariaDB 10.1) does not support -1 as timeout parameters.
		return ErrMySQLInternalError
	case res.Int32 == 0:
		// MySQL Timeout
		return ErrMySQLTimeout
	}

	return nil
}
