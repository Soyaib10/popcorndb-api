while i am learning this i also want to keep a professional documentation on these topis. the documentation should be professional, to the point, expalining what is doing, why is doing, what advantages it gives. documentation should be in english, can you make it for me? while writting these docs i need you to be more concise, to be point and professional and no funcky symbols like ✅ 👉

# Title of the Topic

## Overview

Briefly explain what the concept or technique is, and in what context it is used.

## Problem

Describe the challenge or issue that exists without the solution.
Show with a concise code example or scenario.

```go
// Example demonstrating the problem
```

## Solution

Explain how the issue is addressed.
Show the modified code or approach.

```go
// Example implementing the solution
```

## Example

Provide a practical example of usage in real code. Keep it minimal but realistic.

```go
// Example code demonstrating how it works in practice
```

## Advantages

List the benefits of this approach clearly and concisely.

* Advantages ....

## Best Practices

Summarize recommended guidelines for using this approach in production code.

* Best practices .....


---




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

# 8.1. Partial Updates in Go with JSON and Pointers

## Overview

When implementing partial updates (using HTTP `PATCH`), APIs must distinguish between:

* A field not provided in the JSON.
* A field provided with a zero-value (e.g., `""`, `0`, `false`).

Go’s JSON decoding sets missing fields to zero-values, making this distinction impossible without pointers.

## Problem

```go
var input struct {
    Title   string   `json:"title"`
    Year    int32    `json:"year"`
    Runtime int32    `json:"runtime"`
    Genres  []string `json:"genres"`
}
```

* `{ "year": 2020 }` → `Title = ""`
* `{ "title": "" }` → `Title = ""`

Both cases look identical, preventing correct validation and partial updates.



## Solution: Use Pointers

Redefine struct fields as pointers:

```go
var input struct {
    Title   *string   `json:"title"`
    Year    *int32    `json:"year"`
    Runtime *int32    `json:"runtime"`
    Genres  []string  `json:"genres"`
}
```

* Missing field → `nil`
* Provided field → non-nil (may hold zero-value)

## Example

```go
if input.Title != nil {
    if *input.Title == "" {
        v.AddError("title", "must not be empty")
    } else {
        movie.Title = *input.Title
    }
}

if input.Year != nil {
    movie.Year = *input.Year
}

if input.Runtime != nil {
    movie.Runtime = *input.Runtime
}

if input.Genres != nil {
    movie.Genres = input.Genres
}
```

## Advantages

* Differentiates missing vs. provided fields.
* Enables precise validation.
* Supports partial updates without overwriting unchanged fields.
* Aligns with REST semantics (`PATCH` vs. `PUT`).

## Best Practices

* Use pointers for scalar fields (`*string`, `*int32`, `*bool`) in update inputs.
* Keep slices and maps as-is, since their zero-value is `nil`.
* Validate pointer values only if non-nil.
* Use `PATCH` for partial updates and `PUT` for full replacements.
* Keep update logic explicit: check each field before applying changes.

Got it 👍 No funky icons, only plain professional Markdown.
Here’s the **clean version**:

---



# 8.2. Preventing Data Race with Optimistic Locking

## Why the Data Race Happens

When two concurrent processes (for example, Alice and Bob) try to update the same database record at the same time, a data race occurs. Both read the same initial state, make changes, and then attempt to update. Without proper handling, one update may overwrite the other, causing inconsistent data.


## Solution: Optimistic Locking with Version Numbers

We prevent the data race by using a `version` column in the database and updating records only if the version matches. Each successful update increments the version, ensuring that only one update succeeds.


## Example SQL

```sql
UPDATE movies
SET title = $1,
    year = $2,
    runtime = $3,
    genres = $4,
    version = version + 1
WHERE id = $5
  AND version = $6
RETURNING version;
```

* The `WHERE` clause ensures updates only apply if the version matches.
* The `RETURNING version` clause gives back the new version, keeping the application state in sync with the database.

## Example Scenario

### Initial State

```
MovieID | Title      | Version
--------+-----------+---------
1       | Inception | 1
```

### With `RETURNING version`

```
   Alice reads v1              Bob reads v1
          |                          |
          v                          v
   Alice updates (v1)        Bob updates (v1)
          |                          |
   DB sets version=2        DB expects version=1
   returns version=2         but now version=2
          |                          |
   Alice syncs state         No rows match → error
```

**Result:**

```
MovieID | Title          | Version
--------+----------------+--------
1       | Inception 2020 | 2
```

* Alice’s update succeeds
* Alice’s application now knows version = 2
* Bob’s update fails with conflict

### Without `RETURNING version`

```
   Alice reads v1              Bob reads v1
          |                          |
          v                          v
   Alice updates (v1)        Bob updates (v1)
          |                          |
   DB sets version=2         DB expects version=1
   but Alice's app           but now version=2
   still thinks version=1     → fails with conflict later
```

**Problem:**

* Alice’s update succeeds, but her application still thinks version = 1.
* On her next update, she will try with version = 1 again, causing an unnecessary conflict.
* Application state and database state become inconsistent.

## Why `RETURNING version` is Important

* Keeps application state in sync with the database.
* Avoids unnecessary false conflicts.
* Ensures each successful update carries the latest version forward.

---


# 8.3. Using pg_sleep() in PostgreSQL with Go (pgxpool)

## Problem
- `pg_sleep(n)` returns `void` (`OID 2278`).
- When used directly in `SELECT`, Go/pgx cannot scan into normal types, causing:


cannot scan unknown type (OID 2278) into \*interface{}

- Need a way to add delay without breaking Go Scan.

## Options

### 1. Direct SELECT
```sql
SELECT pg_sleep(2), id, created_at, title, year, runtime, genres, version
FROM movies
WHERE id = $1;
````

Pros:

* Simple to write.

Cons:

* Must add dummy field in Go Scan (`&[]byte{}` or `&dummy`).
* Not clean for production.

### 2. Subquery Version

```sql
SELECT id, created_at, title, year, runtime, genres, version
FROM (
    SELECT pg_sleep(2), id, created_at, title, year, runtime, genres, version
    FROM movies
    WHERE id = $1
) sub;
```

Pros:

* No dummy field in Go.
* Delay works.

Cons:

* Slightly more complex query.

### 3. CTE Version (Recommended)

```sql
WITH delay AS (SELECT pg_sleep(2))
SELECT id, created_at, title, year, runtime, genres, version
FROM movies
WHERE id = $1;
```

Pros:

* Clean separation, delay isolated in CTE.
* Go Scan works normally.
* Production-friendly.

Cons:

* Requires understanding of CTE.

## Go Example (CTE version)

```go
query := `
    WITH delay AS (SELECT pg_sleep(10))
    SELECT id, created_at, title, year, runtime, genres, version
    FROM movies
    WHERE id = $1
`

err := m.DB.QueryRow(context.Background(), query, id).Scan(
    &movie.ID,
    &movie.CreatedAt,
    &movie.Title,
    &movie.Year,
    &movie.Runtime,
    &movie.Genres,
    &movie.Version,
)
```

## Verdict

* Quick test/demo → Direct SELECT (with dummy field).
* Cleaner inline delay → Subquery.
* Best practice → CTE (no dummy, clean, readable).

```

---

যদি চাও আমি এটা **ডিরেক্টলি `.md` ফাইলে তৈরি করে দিয়ে দিতে পারি**, ready to use।  
চাও আমি সেটা করি?
```
