package orm

import (
	"errors"
	"strings"
	"testing"
)

func TestQueryBuilder(t *testing.T) {
	sql, args := Table("users").
		Select("id", "email").
		Where("status", "=", "active").
		Limit(10).
		ToSQL()

	if !strings.Contains(sql, "SELECT id, email FROM users WHERE status = $1 LIMIT 10") {
		t.Errorf("unexpected SQL: %s", sql)
	}
	if len(args) != 1 || args[0] != "active" {
		t.Errorf("unexpected args: %v", args)
	}
}

func TestWhereRejectsInjection(t *testing.T) {
	for _, q := range []*QueryBuilder{
		Table("users").Where("id; DROP TABLE users", "=", 1),
		Table("users").Where("id", "= 1; DROP TABLE users --", 1),
		Table("users; DROP TABLE users"),
		Table("users").Select("id, (SELECT password FROM users)"),
	} {
		if q.Err() == nil {
			t.Errorf("expected validation error, got none")
		}
		if sql, _ := q.ToSQL(); sql != "" {
			t.Errorf("rejected query must produce no SQL, got %q", sql)
		}
	}
	// Valid qualified column + case-insensitive operator still work.
	q := Table("users").Where("users.status", "like", "a%")
	if q.Err() != nil {
		t.Fatalf("unexpected error: %v", q.Err())
	}
	if sql, _ := q.ToSQL(); !strings.Contains(sql, "users.status LIKE $1") {
		t.Errorf("unexpected SQL: %s", sql)
	}
}

func TestDBPoolTransactionErrorRollsBack(t *testing.T) {
	pool := NewDBPool()
	want := errors.New("fake")
	var got *Tx
	err := pool.Transaction(func(tx *Tx) error {
		got = tx
		return want
	})
	if err != want || got == nil || !got.RolledBack() || got.Committed() {
		t.Errorf("expected rollback on error, err=%v tx=%+v", err, got)
	}
}

func TestDBPool(t *testing.T) {
	pool := NewDBPool()
	err := pool.Transaction(func(tx *Tx) error {
		return nil
	})
	if err != nil {
		t.Errorf("transaction failed: %v", err)
	}
}
