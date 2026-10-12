# Alap Framework — Data Layer

Alap provides three data facilities:

| Package | Purpose |
|---------|---------|
| `data/orm` | SQL query builder, connection pool, transactions. |
| `data/migration` | Versioned, ordered schema migrations. |
| `stdlib/database` | A lightweight in-process JSON-backed database. |

---

## Query builder (`data/orm`)

`Table(name)` starts a fluent builder. Values are parameterised (never string
interpolated), so the generated SQL is injection-safe.

```go
sql, params := orm.Table("orders").
    Select("id", "customer_id", "total", "status").
    Where("status", "=", "completed").
    Where("total", ">=", 500).
    Limit(20).
    ToSQL()
```

Generated shape:

```sql
SELECT id, customer_id, total, status FROM orders WHERE status = $1 AND total >= $2 LIMIT 20
```

| Method | Description |
|--------|-------------|
| `Table(name)` | Begin a query against a table. |
| `Select(cols...)` | Projection (defaults to `*`). |
| `Where(col, op, val)` | Add a predicate; `op` is `=`, `>=`, `<`, … |
| `Limit(n)` | Row cap. |
| `ToSQL()` | Return `(sql, params)`. |

> `Where` stores the operator and value; `ToSQL()` is the canonical place to
> inspect the exact emitted SQL. The exact placeholder dialect is fixed by the
> builder — treat `ToSQL()` output as the contract.

---

## Transactions (`data/orm`)

```go
pool := orm.NewDBPool()
err := pool.Transaction(func(tx *orm.Tx) error {
    // perform work via tx
    return tx.Commit()
})
```

`DBPool.Transaction` runs the callback and is the intended boundary for atomic
units of work.

---

## Migrations (`data/migration`)

```go
runner := migration.NewRunner().
    Register(migration.Migration{Version: 1, Name: "create_users", Up: createUsers}).
    Register(migration.Migration{Version: 2, Name: "add_index",    Up: addIndex})

applied := runner.Up()   // returns the versions applied, in order
```

`Up()` applies registered migrations in version order and returns the list of
applied versions — suitable for asserting in tests.

---

## Embedded database (`stdlib/database`)

`stdlib/database` is a file-backed, in-process database useful for examples,
tests, and small tools. It stores tables as records and persists to a JSON file.

```go
db, err := database.New("pos.db")
if err != nil { /* ... */ }
defer db.Close()

db.CreateTable("orders")
db.Insert("orders", database.Record{"id": 1, "total": 99})

rows := db.Query("orders", map[string]interface{}{"total": 99}, 10)
count := db.Count("orders", nil)

db.Update("orders", map[string]interface{}{"id": 1}, database.Record{"total": 120})
db.Delete("orders", map[string]interface{}{"id": 1})

names := db.TableNames()
```

| Method | Description |
|--------|-------------|
| `New(path)` | Open/create a database file. |
| `CreateTable(name)` | Create a table. |
| `Insert(table, rec)` | Insert a record. |
| `Update(table, where, updates)` | Update matching records; returns count. |
| `Delete(table, where)` | Delete matching records; returns count. |
| `Query(table, where, limit)` | Select matching records. |
| `Count(table, where)` | Count matching records. |
| `GetRecord(table, where)` | Fetch one record. |
| `TableNames()` / `Table(name)` | Introspect tables. |
| `Close()` | Persist and close. |

---

## Files (`stdlib/fs`)

```go
_ = fs.WriteTextFile("notes.txt", "hello")
content, _ := fs.ReadTextFile("notes.txt")
_ = fs.AppendTextFile("notes.txt", "\nmore")
exists := fs.Exists("notes.txt")
entries, _ := fs.ReadDir(".")
```

Available: `ReadTextFile`, `WriteTextFile`, `AppendTextFile`, `Exists`,
`Remove`, `RemoveAll`, `Mkdir`, `ReadDir`.

---

## HTTP client (`stdlib/net`)

```go
body, _ := net.Get("https://example.com")
_ = net.Post(url, "application/json", `{"ok":true}`)
resp, _ := net.Fetch("GET", url, map[string]string{"Accept": "application/json"}, "")
// resp has Body / status fields
```

---

## Status

The data layer is functional and tested at the builder/migration/DB level. It is
**not** a production database: it does not manage network drivers, pooling
against a real server, or schema introspection. See [`ROADMAP.md`](./ROADMAP.md).
