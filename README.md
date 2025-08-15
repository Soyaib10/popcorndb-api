

# 1. Panic vs Return - When to do what? page 87

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

# Migration Error and Solution

## Dirty Database Recovery Workflow (golang-migrate)

### 1. Dirty DB State Example

```sql
SELECT * FROM schema_migrations;

 version | dirty
---------+------
       3 | t
```

* `version = 3` → Last migration partially applied
* `dirty = true` → DB inconsistent

---

### 2. Step-by-Step Recovery

#### Step 1: Identify failed migration

Check which migration failed and which statements were applied.

```bash
cat ./migrations/000003_failed_migration.up.sql
```

#### Step 2: Manually revert partially applied changes

Undo changes that were created or modified in the DB.

```sql
DROP TABLE directors;           -- If table was created
ALTER TABLE movies DROP COLUMN age; -- If column was added
```

#### Step 3: Force version in schema_migrations

To signal a clean state in the database:

```bash
migrate -path=./migrations -database="$POPCORN_DB_DSN" force 2
```

* Here, `2` → last successfully applied migration

#### Step 4: Re-run migration

```bash
migrate -path=./migrations -database="$POPCORN_DB_DSN" up
```

* Migration will now apply

```sql
SELECT * FROM schema_migrations;

 version | dirty
---------+------
       3 | f
```

✅ Dirty state gone, migration successful

---

### 3. Summary Diagram

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

### Notes

* Always keep `.up.sql` and `.down.sql` for every migration
* Cleanup and force are mandatory before running migrations in a dirty state
* Ensure proper privileges when running migrations in production

---

### Summary

1. If a migration has a syntax error, it may be partially applied → DB dirty.
2. `schema_migrations` table shows version + dirty=true.
3. Fix the error → rollback DB → force clean version.
4. Then you can re-run the migration.
5. Remote migration support is available (S3, GitHub).

---

# Go Struct Embedding / Nesting

* Struct nesting allows embedding one struct inside another.
* Makes “has-a” relationships explicit.
* Improves readability, maintainability, and reusability of code.
* Useful in real-world models: Users, Orders, Cars, Products, etc.

```go
package main

import "fmt"

type Address struct {
    Thana string
    Jila string
}

type Person struct {
    Name string
    Age int
    Address Address
}

func NewPerson(name string, age int, thana string, jila string) Person {
    return Person {
        Name: name,
        Age: age,
        Address: Address{
            Thana: thana,
            Jila: jila,
        },
    }
}

func main() {
	person := NewPerson("abul", 23, "nilfamari", "kanchonongha")
	fmt.Print(person.Name)
	fmt.Print(person.Address.Thana)
}
```

---

# Connection Pooling: Explicit vs Implicit

```
┌─────────────────────┐
│   Application Code  │
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│  Connection Pool    │
│                     │
│ Explicit (pgxpool)  │
│ ────────────────── │
│ - MaxConns = 25     │
│ - MinConns = 5      │
│ - IdleTimeout = 15m │
│ - Fully controlled  │
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│  PostgreSQL Server  │
└─────────────────────┘


┌─────────────────────┐
│   Application Code  │
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│  Connection Pool    │
│                     │
│ Implicit (sql.DB)   │
│ ─────────────────  │
│ - MaxOpenConns = 25 │ │
│ - MaxIdleConns = 25 │
│ - Defaults handled  │
│   automatically     │
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│  PostgreSQL Server  │
└─────────────────────┘
```

### Key Points:

* **Explicit (pgxpool)**:

  * You configure the pool yourself.
  * Max/Min connections, idle timeout controllable.
  * Full Postgres features supported.

* **Implicit (sql.DB)**:

  * Mostly automatic pooling.
  * You can optionally set limits.
  * Generic SQL, slow Postgres-specific features.

---
