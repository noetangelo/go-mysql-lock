package mysql

import (
	"context"
	"database/sql"
	"dmutex"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
	"github.com/tangelo-labs/go-dotenv"
	"log"
	"sync/atomic"
	"testing"
	"time"
)

type environment struct {
	MysqlDsn string `env:"MYSQL_DSN"`
}

func TestNewMysqlLocker(t *testing.T) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()

	env := environment{}

	require.NoError(t, dotenv.LoadAndParse(&env))

	db, err := sql.Open("mysql", env.MysqlDsn)
	require.NoError(t, err)

	db.SetMaxOpenConns(90)
	db.SetMaxIdleConns(45)
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(10 * time.Minute)

	t.Run("GIVEN several processes willing for acquire the same lock", func(t *testing.T) {
		readyChan := make(chan struct{})

		var counterOK uint32
		var counterErr uint32

		numProcesses := 50

		lockers := make([]dmutex.Locker, numProcesses)

		lockerIdChan := make(chan int, numProcesses)

		for i := 0; i < numProcesses; i++ {
			go func(i int) {
				conn, cErr := db.Conn(ctx)
				require.NoError(t, cErr)

				mLocker := NewLocker(conn)

				lockers[i] = mLocker

				<-readyChan

				lErr := mLocker.Lock(ctx, "test")
				if lErr != nil {
					log.Print(lErr)

					atomic.AddUint32(&counterErr, 1)

					return
				}

				atomic.AddUint32(&counterOK, 1)

				lockerIdChan <- i
			}(i)
		}

		t.Run("WHEN acquiring the lock at the same time", func(t *testing.T) {
			close(readyChan)

			t.Run("THEN only one process should have acquired the lock", func(t *testing.T) {
				require.Eventually(t, func() bool {
					return atomic.LoadUint32(&counterOK) == 1 && atomic.LoadUint32(&counterErr) == 0
				}, 500*time.Second, 100*time.Millisecond, "all processes should have finished")

				t.Run("AND releasing the lock consecutively makes all processes to acquire the lock", func(t *testing.T) {
					for i := 0; i < numProcesses; i++ {
						lockIdx := <-lockerIdChan

						rErr := lockers[lockIdx].Release("test")
						require.NoError(t, rErr)
					}

					require.Eventually(t, func() bool {
						return atomic.LoadUint32(&counterOK) == uint32(numProcesses) && atomic.LoadUint32(&counterErr) == 0
					}, 500*time.Second, 100*time.Millisecond, "all processes should have finished")
				})
			})
		})
	})
}
