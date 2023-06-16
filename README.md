# dmutex

This package provides a interface which defines a distributed mutex. Several implementations are provided.
It is based on the [gomysqllock]

Currently, the following implementations are provided:
- [MySQL](#mysql)

#### Installation
```$bash
go get github.com/tamgelo-labs/dmutex
```

### Example:

```go
package main

import (
	"context"
	"database/sql"
	"dmutex/mysql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/tamgelo-labs/dmutex"
)

func main() {
	db, _ := sql.Open("mysql", "root@tcp(localhost:3306)/dyno_test")
	
    conn, err := db.Conn(context.Background())
	if err != nil {
        panic(err)
    }
	
	locker := mysql.NewLocker(conn)

	lock, _ := locker.Lock(ctx, "foo")
	lock.Release()
}
```
