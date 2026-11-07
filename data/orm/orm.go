// Package orm provides a prototype SQL query builder.
//
// Safety model: table/column names and operators cannot be bound as SQL
// parameters, so they are validated against an identifier pattern / an
// operator allowlist before being interpolated. Values are always bound
// as parameters ($1, $2, ...), never interpolated. An invalid identifier
// or operator is not interpolated: it is recorded (see QueryBuilder.Err)
// and ToSQL then returns no SQL, so a rejected query cannot be executed
// by accident through this builder.
//
// Transactions: DBPool/Tx are an in-memory prototype only. There is no
// real database connection, driver, atomicity, or rollback behind them;
// Transaction merely serialises the callback under the pool mutex and
// tracks commit/rollback state. Do not rely on it for real transactional
// guarantees.
package orm

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

// identRe matches a (possibly qualified) SQL identifier: name or table.name.
var identRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z_][A-Za-z0-9_]*)?$`)

// allowedOperators is the Where operator allowlist (compared upper-cased).
var allowedOperators = map[string]bool{
	"=": true, "!=": true, "<>": true,
	"<": true, "<=": true, ">": true, ">=": true,
	"LIKE": true, "ILIKE": true, "IN": true,
	"IS": true, "IS NOT": true,
}

func validIdent(s string) bool { return identRe.MatchString(s) }

type QueryBuilder struct {
	table    string
	selected []string
	wheres   []string
	args     []interface{}
	limit    int
	err      error // first validation error, if any
}

func Table(name string) *QueryBuilder {
	q := &QueryBuilder{
		table:    name,
		selected: []string{"*"},
		wheres:   make([]string, 0),
		args:     make([]interface{}, 0),
	}
	if !validIdent(name) {
		q.err = fmt.Errorf("orm: invalid table name %q", name)
	}
	return q
}

func (q *QueryBuilder) Select(cols ...string) *QueryBuilder {
	for _, c := range cols {
		if c != "*" && !validIdent(c) {
			if q.err == nil {
				q.err = fmt.Errorf("orm: invalid column name %q", c)
			}
			return q
		}
	}
	q.selected = cols
	return q
}

// Where adds a predicate. col must be a plain (optionally qualified)
// identifier and op must be on the allowlist; otherwise the predicate is
// dropped and the error is recorded in Err (and ToSQL returns no SQL).
func (q *QueryBuilder) Where(col, op string, val interface{}) *QueryBuilder {
	opNorm := strings.ToUpper(strings.Join(strings.Fields(op), " "))
	if !validIdent(col) {
		if q.err == nil {
			q.err = fmt.Errorf("orm: invalid column name %q in Where", col)
		}
		return q
	}
	if !allowedOperators[opNorm] {
		if q.err == nil {
			q.err = fmt.Errorf("orm: operator %q is not allowed in Where", op)
		}
		return q
	}
	argIdx := len(q.args) + 1
	q.wheres = append(q.wheres, fmt.Sprintf("%s %s $%d", col, opNorm, argIdx))
	q.args = append(q.args, val)
	return q
}

func (q *QueryBuilder) Limit(limit int) *QueryBuilder {
	q.limit = limit
	return q
}

// Err returns the first validation error recorded while building, if any.
func (q *QueryBuilder) Err() error { return q.err }

func (q *QueryBuilder) ToSQL() (string, []interface{}) {
	if q.err != nil {
		return "", nil
	}
	sql := fmt.Sprintf("SELECT %s FROM %s", strings.Join(q.selected, ", "), q.table)
	if len(q.wheres) > 0 {
		sql += " WHERE " + strings.Join(q.wheres, " AND ")
	}
	if q.limit > 0 {
		sql += fmt.Sprintf(" LIMIT %d", q.limit)
	}
	return sql, q.args
}

// Tx tracks the outcome of a prototype DBPool.Transaction. It holds no
// database connection and buffers no operations: Commit/Rollback only
// record state, they do not make any real work atomic or undo it.
type Tx struct {
	committed  bool
	rolledBack bool
}

func (tx *Tx) Commit() error {
	tx.committed = true
	return nil
}

// Rollback marks the transaction rolled back (state only; see package doc).
func (tx *Tx) Rollback() error {
	tx.rolledBack = true
	return nil
}

func (tx *Tx) Committed() bool  { return tx.committed }
func (tx *Tx) RolledBack() bool { return tx.rolledBack }

// DBPool is a prototype transaction boundary: it serialises Transaction
// callbacks with a mutex. It is not a connection pool.
type DBPool struct {
	mu sync.Mutex
}

func NewDBPool() *DBPool {
	return &DBPool{}
}

// Transaction runs fn while holding the pool lock. If fn returns an
// error, the Tx is marked rolled back and the error is returned;
// otherwise the Tx is committed (unless fn already committed it).
func (p *DBPool) Transaction(fn func(tx *Tx) error) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	tx := &Tx{}
	if err := fn(tx); err != nil {
		tx.rolledBack = true
		return err
	}
	if !tx.committed {
		return tx.Commit()
	}
	return nil
}
