// Package nil provides database native functions registration for NilLang / Alap.
// Provides db_open, db_close, db_insert, db_select, db_update, db_delete, db_count
package nil

import (
	"encoding/json"
	"fmt"

	"github.com/joysriramsarkar/alap-framework/runtime/vm"
	"github.com/joysriramsarkar/alap-framework/stdlib/database"
)

// RegisterDBNativeFunctions registers database operations as VM native functions.
func RegisterDBNativeFunctions(v *vm.VM) {
	v.RegisterNative("db_open", func(args []vm.Value) (vm.Value, error) {
		if len(args) == 0 {
			return vm.Nil, fmt.Errorf("db_open: path required")
		}
		db, err := database.New(args[0].StrVal)
		if err != nil {
			return vm.Nil, err
		}
		handle := registerDBHandle(args[0].StrVal, db)
		return vm.StrVal(handle), nil
	})

	v.RegisterNative("db_close", func(args []vm.Value) (vm.Value, error) {
		if len(args) == 0 {
			return vm.Nil, fmt.Errorf("db_close: handle required")
		}
		handle := args[0].StrVal
		if len(handle) >= 3 && handle[:3] == "db:" {
			delete(dbInstances, handle[3:])
		}
		return vm.True, nil
	})

	v.RegisterNative("db_insert", func(args []vm.Value) (vm.Value, error) {
		if len(args) < 3 {
			return vm.Nil, fmt.Errorf("db_insert: handle, table, json required")
		}
		db := lookupDB(args[0])
		if db == nil {
			return vm.Nil, fmt.Errorf("db_insert: invalid handle")
		}
		table := args[1].StrVal
		var rec map[string]interface{}
		if err := json.Unmarshal([]byte(args[2].StrVal), &rec); err != nil {
			return vm.Nil, fmt.Errorf("db_insert: invalid json: %w", err)
		}
		result := db.Insert(table, database.Record(rec))
		data, _ := json.Marshal(result)
		return vm.StrVal(string(data)), nil
	})

	v.RegisterNative("db_select", func(args []vm.Value) (vm.Value, error) {
		if len(args) < 3 {
			return vm.Nil, fmt.Errorf("db_select: handle, table, json_where required")
		}
		db := lookupDB(args[0])
		if db == nil {
			return vm.Nil, fmt.Errorf("db_select: invalid handle")
		}
		table := args[1].StrVal
		var where map[string]interface{}
		if args[2].StrVal != "" {
			_ = json.Unmarshal([]byte(args[2].StrVal), &where)
		}
		limit := 0
		if len(args) > 3 {
			limit = int(args[3].IntVal)
		}
		records := db.Query(table, where, limit)
		if records == nil {
			records = make([]database.Record, 0)
		}
		data, _ := json.Marshal(records)
		return vm.StrVal(string(data)), nil
	})

	v.RegisterNative("db_update", func(args []vm.Value) (vm.Value, error) {
		if len(args) < 4 {
			return vm.Nil, fmt.Errorf("db_update: handle, table, where_json, updates_json required")
		}
		db := lookupDB(args[0])
		if db == nil {
			return vm.Nil, fmt.Errorf("db_update: invalid handle")
		}
		table := args[1].StrVal
		var where, updates map[string]interface{}
		_ = json.Unmarshal([]byte(args[2].StrVal), &where)
		_ = json.Unmarshal([]byte(args[3].StrVal), &updates)
		count := db.Update(table, where, database.Record(updates))
		return vm.IntVal(int64(count)), nil
	})

	v.RegisterNative("db_delete", func(args []vm.Value) (vm.Value, error) {
		if len(args) < 3 {
			return vm.Nil, fmt.Errorf("db_delete: handle, table, json_where required")
		}
		db := lookupDB(args[0])
		if db == nil {
			return vm.Nil, fmt.Errorf("db_delete: invalid handle")
		}
		table := args[1].StrVal
		var where map[string]interface{}
		_ = json.Unmarshal([]byte(args[2].StrVal), &where)
		count := db.Delete(table, where)
		return vm.IntVal(int64(count)), nil
	})

	v.RegisterNative("db_count", func(args []vm.Value) (vm.Value, error) {
		if len(args) < 3 {
			return vm.Nil, fmt.Errorf("db_count: handle, table, json_where required")
		}
		db := lookupDB(args[0])
		if db == nil {
			return vm.Nil, fmt.Errorf("db_count: invalid handle")
		}
		table := args[1].StrVal
		var where map[string]interface{}
		_ = json.Unmarshal([]byte(args[2].StrVal), &where)
		count := db.Count(table, where)
		return vm.IntVal(int64(count)), nil
	})
}

var dbInstances = make(map[string]*database.Database)

func lookupDB(handle vm.Value) *database.Database {
	s := handle.StrVal
	if len(s) < 4 || s[:3] != "db:" {
		return nil
	}
	ptrStr := s[3:]
	db, ok := dbInstances[ptrStr]
	if !ok {
		return nil
	}
	return db
}

func registerDBHandle(path string, db *database.Database) string {
	key := fmt.Sprintf("%p", db)
	dbInstances[key] = db
	return "db:" + key
}
