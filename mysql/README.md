### Mysql implementation

This package provides a MySQL implementation of the `Locker` interface.

`Lock` and `Release` methods strongly relies on provided Context. Is used to cancel the lock acquisition and lock releasing.
If the context is canceled before the lock is acquired, an error is returned. 
The same happens if the context is canceled before the lock is released, so be careful with the context you provide, may be the context is canceled before you release the lock.

According to [Mysql documentation](https://dev.mysql.com/doc/refman/8.0/en/locking-functions.html#function_get-lock):

    - Keys locked by Lock are not released when transactions commit or roll back.
    - Locks are automatically released when the session is terminated (either normally or abnormally).

is your responsibility to release the lock when you are done with it.


### Compatibility

This library has been teste against MySQL 5.7 and 8.0. It should work with any MySQL version >= 5.7.

