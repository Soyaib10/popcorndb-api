1. panic vs return- when to do what? page 87

# 4.3- Restricting inputs
- add error handling for checking unknown fileds, multiple json and large files more than one mb

# 4.4- Custom JSON Decoding

# 5.3- Connection Pool, golang contenxt package

# go mod and go sum

# database connection setup- key things to remember and why 

---

# PostgreSQL Connection Setup in Go with pgxpool

This is a reference for setting up a PostgreSQL connection pool in Go using `pgxpool`, with support for configurable connection limits via command-line flags or environment variables.

---

## 1. DSN

The DSN (Data Source Name) is the PostgreSQL connection string. Example:

```
postgres://popcorndb_api:paSSWORD@localhost:5432/popcorndb_api?sslmode=disable
```

> Note: Keeping the DSN in flags or environment variables only stores values in Go variables. It does not automatically affect the connection pool behavior.

---

## 2. Command-line Flags

You can use Go `flag` package to configure the connection:

```go
var cfg struct {
    db struct {
        dsn          string
        maxOpenConns int
        maxIdleConns int
        maxIdleTime  string
    }
}

flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres://user:pass@localhost/db", "PostgreSQL DSN")
flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 5, "Max open connections")
flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 2, "Max idle connections")
flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "Max idle time")
flag.Parse()
```

> These flags only hold values in the struct. They do not change the pool behavior by themselves.

---

## 3. ParseConfig

Use `pgxpool.ParseConfig(dsn)` to convert the DSN string into a structured `pgxpool.Config`:

```go
poolConfig, err := pgxpool.ParseConfig(cfg.db.dsn)
if err != nil {
    log.Fatal(err)
}

fmt.Println(poolConfig.ConnConfig.User)  // popcorndb_api
```

---

## 4. Configure Pool Limits

Set connection limits on the `pgxpool.Config` struct:

```go
poolConfig.MaxConns = int32(cfg.db.maxOpenConns)         // maximum open connections
poolConfig.MinConns = int32(cfg.db.maxIdleConns)        // minimum idle connections
duration, _ := time.ParseDuration(cfg.db.maxIdleTime)
poolConfig.MaxConnIdleTime = duration
```

---

## 5. Create Pool

Use `NewWithConfig` to apply the custom pool limits:

```go
pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
if err != nil {
    log.Fatal(err)
}
```

---

## 6. Ping Database

Check if the connection is valid:

```go
if err := pool.Ping(context.Background()); err != nil {
    pool.Close()
    log.Fatal(err)
}
```

---

## 7. Ideal Flow

1. Create a context with timeout
2. Parse the DSN → `pgxpool.Config`
3. Set pool limits (`MaxConns`, `MinConns`, `MaxConnIdleTime`)
4. Create the pool with `NewWithConfig`
5. Ping the database
6. Use the pool for queries

---

## 8. Why Flags Alone Are Not Enough

Simply setting flags does **not** enforce pool limits:

```go
pool, err := pgxpool.New(context.Background(), cfg.db.dsn)
```

* MaxConns, MinConns, MaxConnIdleTime are ignored
* The pool uses default values (MaxConns=0, MinConns=0)

Correct approach:

```go
poolConfig, _ := pgxpool.ParseConfig(cfg.db.dsn)
poolConfig.MaxConns = int32(cfg.db.maxOpenConns)
poolConfig.MinConns = int32(cfg.db.maxIdleConns)
poolConfig.MaxConnIdleTime, _ = time.ParseDuration(cfg.db.maxIdleTime)

pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
```

Now the flags actually control the pool limits.

---

## 9. Summary

* DSN string → ParseConfig → structured config
* Flags → values stored in struct
* Pool limits → set in `pgxpool.Config`
* Create pool → `NewWithConfig`
* Ping → verify connection

---

নিশ্চয়ই! আমি তোমার দেওয়া কনটেন্টকে **README.md ফরম্যাট**-এ সুন্দরভাবে সাজিয়ে দিচ্ছি, যাতে তুমি সরাসরি copy-paste করতে পারো।

````markdown
# Migration Error and Solution

## Dirty Database Recovery Workflow (golang-migrate)

### ১. Dirty DB State উদাহরণ

```sql
SELECT * FROM schema_migrations;

 version | dirty
---------+------
       3 | t
````

* `version = 3` → শেষ migration partially applied
* `dirty = true` → DB inconsistent

---

### ২. ধাপে ধাপে Recovery

#### Step 1: Identify failed migration

কোন migration fail হয়েছে এবং কোন statements apply হয়েছে তা দেখো।

```bash
cat ./migrations/000003_failed_migration.up.sql
```

#### Step 2: Manually revert partially applied changes

যা DB-এ create হয়েছে বা modify হয়েছে, সেগুলো undo করো।

```sql
DROP TABLE directors;           -- যদি table তৈরি হয়ে থাকে
ALTER TABLE movies DROP COLUMN age; -- যদি column add হয়ে থাকে
```

#### Step 3: Force version in schema\_migrations

DB clean state signal করতে:

```bash
migrate -path=./migrations -database="$POPCORN_DB_DSN" force 2
```

* এখানে `2` → last successfully applied migration

#### Step 4: Re-run migration

```bash
migrate -path=./migrations -database="$POPCORN_DB_DSN" up
```

* এখন migration apply হবে

```sql
SELECT * FROM schema_migrations;

 version | dirty
---------+------
       3 | f
```

✅ Dirty state gone, migration successful

---

### ৩. Summary Diagram

```
DB dirty (dirty=true, version=X)
        |
        v
Investigate failed migration (check SQL)
        |
        v
Manually revert partial changes
        |
        v
Force version to last good migration
        |
        v
Re-run migration (migrate up)
        |
        v
DB clean (dirty=false, version updated)
```

---

### 💡 Notes

* প্রতিটি migration সবসময় `.up.sql` ও `.down.sql` রাখো
* Dirty state এ পরের migration চালানোর আগে **cleanup + force** করা বাধ্যতামূলক
* Production DB-তে migration চালানোর সময় proper privileges নিশ্চিত করো

---

### সারসংক্ষেপ

1. Migration এ syntax error হলে partially applied হতে পারে → DB dirty।
2. `schema_migrations` টেবিলে version + dirty=true দেখাবে।
3. Error ঠিক করে → DB rollback করো → force দিয়ে version clean করো।
4. তারপর আবার migration চালাতে পারো।
5. Remote migration support আছে (S3, GitHub)।

