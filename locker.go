package dmutex

import "context"

type Locker interface {
	Lock(ctx context.Context, key string) error
	Release(ctx context.Context, key string) error
}
