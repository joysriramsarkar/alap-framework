// Package vm implements the nilrt virtual machine that executes NABC bytecode.
package vm

import (
	"encoding/json"
	"fmt"

	"github.com/joysriramsarkar/alap-framework/stdlib/database"
)

// dbHandles stores live database instances keyed by their handle string.
var dbHandles = make(map[string]*database.Database)

// registerDBBuiltins registers all database native functions.
func (vm *VM) registerDBBuiltins() {
	vm.RegisterNative("db_open", func(args []Value) (Value, error) {
		if len(args) == 0 {
			return Nil, fmt.Errorf("db_open: path required")
		}
		db, err := database.New(args[0].StrVal)
		if err != nil {
			return Nil, err
		}
		handle := fmt.Sprintf("db_%p", db)
		dbHandles[handle] = db
		return StrVal(handle), nil
	})

	vm.RegisterNative("db_close", func(args []Value) (Value, error) {
		if len(args) == 0 {
			return Nil, fmt.Errorf("db_close: handle required")
		}
		handle := args[0].StrVal
		delete(dbHandles, handle)
		return True, nil
	})

	vm.RegisterNative("db_insert", func(args []Value) (Value, error) {
		if len(args) < 3 {
			return Nil, fmt.Errorf("db_insert: handle, table, json required")
		}
		db := dbHandles[args[0].StrVal]
		if db == nil {
			return Nil, fmt.Errorf("db_insert: invalid handle")
		}
		table := args[1].StrVal
		var rec map[string]interface{}
		if err := json.Unmarshal([]byte(args[2].StrVal), &rec); err != nil {
			return Nil, fmt.Errorf("db_insert: invalid json: %w", err)
		}
		result := db.Insert(table, database.Record(rec))
		data, _ := json.Marshal(result)
		return StrVal(string(data)), nil
	})

	vm.RegisterNative("db_select", func(args []Value) (Value, error) {
		if len(args) < 3 {
			return Nil, fmt.Errorf("db_select: handle, table, json_where required")
		}
		db := dbHandles[args[0].StrVal]
		if db == nil {
			return Nil, fmt.Errorf("db_select: invalid handle")
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
		return StrVal(string(data)), nil
	})

	vm.RegisterNative("db_update", func(args []Value) (Value, error) {
		if len(args) < 4 {
			return Nil, fmt.Errorf("db_update: handle, table, where_json, updates_json required")
		}
		db := dbHandles[args[0].StrVal]
		if db == nil {
			return Nil, fmt.Errorf("db_update: invalid handle")
		}
		table := args[1].StrVal
		var where, updates map[string]interface{}
		_ = json.Unmarshal([]byte(args[2].StrVal), &where)
		_ = json.Unmarshal([]byte(args[3].StrVal), &updates)
		count := db.Update(table, where, database.Record(updates))
		return IntVal(int64(count)), nil
	})

	vm.RegisterNative("db_delete", func(args []Value) (Value, error) {
		if len(args) < 3 {
			return Nil, fmt.Errorf("db_delete: handle, table, json_where required")
		}
		db := dbHandles[args[0].StrVal]
		if db == nil {
			return Nil, fmt.Errorf("db_delete: invalid handle")
		}
		table := args[1].StrVal
		var where map[string]interface{}
		_ = json.Unmarshal([]byte(args[2].StrVal), &where)
		count := db.Delete(table, where)
		return IntVal(int64(count)), nil
	})

	vm.RegisterNative("db_count", func(args []Value) (Value, error) {
		if len(args) < 3 {
			return Nil, fmt.Errorf("db_count: handle, table, json_where required")
		}
		db := dbHandles[args[0].StrVal]
		if db == nil {
			return Nil, fmt.Errorf("db_count: invalid handle")
		}
		table := args[1].StrVal
		var where map[string]interface{}
		_ = json.Unmarshal([]byte(args[2].StrVal), &where)
		count := db.Count(table, where)
		return IntVal(int64(count)), nil
	})

	vm.RegisterNative("db_get", func(args []Value) (Value, error) {
		if len(args) < 3 {
			return Nil, fmt.Errorf("db_get: handle, table, json_where required")
		}
		db := dbHandles[args[0].StrVal]
		if db == nil {
			return Nil, fmt.Errorf("db_get: invalid handle")
		}
		table := args[1].StrVal
		var where map[string]interface{}
		_ = json.Unmarshal([]byte(args[2].StrVal), &where)
		rec, found := db.GetRecord(table, where)
		if !found {
			return Nil, nil
		}
		data, _ := json.Marshal(rec)
		return StrVal(string(data)), nil
	})

	vm.RegisterNative("db_tables", func(args []Value) (Value, error) {
		if len(args) < 1 {
			return Nil, fmt.Errorf("db_tables: handle required")
		}
		db := dbHandles[args[0].StrVal]
		if db == nil {
			return Nil, fmt.Errorf("db_tables: invalid handle")
		}
		names := db.TableNames()
		if names == nil {
			names = make([]string, 0)
		}
		data, _ := json.Marshal(names)
		return StrVal(string(data)), nil
	})

	vm.RegisterNative("db_create_table", func(args []Value) (Value, error) {
		if len(args) < 2 {
			return Nil, fmt.Errorf("db_create_table: handle, table required")
		}
		db := dbHandles[args[0].StrVal]
		if db == nil {
			return Nil, fmt.Errorf("db_create_table: invalid handle")
		}
		db.CreateTable(args[1].StrVal)
		return True, nil
	})
}
