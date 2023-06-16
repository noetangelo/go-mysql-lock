package dmutex

import "errors"

var (
	// ErrGetLockContextCancelled is returned when user given context is cancelled while trying to obtain the lock
	ErrGetLockContextCancelled = errors.New("context cancelled while trying to obtain lock")

	// ErrLockReleased is returned when any problem happens releasing the lock.
	ErrLockReleased = errors.New("release error")
)
