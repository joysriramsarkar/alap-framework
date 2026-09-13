// Package database provides a lightweight file-based database for NilLang
// applications. It uses JSON-lines format for persistence and is designed
// for POS, inventory, and sales applications running on Alap/Onuron.
package database

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// Record represents a single database row as key-value pairs.
type Record map[string]interface{}

// Database is a file-backed key-value store for NilLang apps.
type Database struct {
	mu     sync.RWMutex
	path   string
	tables map[string][]Record
}

// New creates or opens a database at the given path.
func New(path string) (*Database, error) {
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed creating database directory: %w", err)
		}
	}

	db := &Database{
		path:   path,
		tables: make(map[string][]Record),
	}
	if err := db.load(); err != nil {
		return nil, err
	}
	return db, nil
}

func (db *Database) load() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	data, err := os.ReadFile(db.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var entry struct {
			Table  string `json:"table"`
			Record Record `json:"record"`
		}
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		db.tables[entry.Table] = append(db.tables[entry.Table], entry.Record)
	}
	return nil
}

func (db *Database) saveLocked() error {
	tmpPath := db.path + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(f)
	for table, records := range db.tables {
		for _, rec := range records {
			entry := struct {
				Table  string `json:"table"`
				Record Record `json:"record"`
			}{Table: table, Record: rec}
			data, _ := json.Marshal(entry)
			writer.Write(data)
			writer.WriteString("\n")
		}
	}
	if err := writer.Flush(); err != nil {
		f.Close()
		return err
	}
	_ = f.Sync()
	f.Close()

	if err := os.Rename(tmpPath, db.path); err != nil {
		_ = os.Remove(db.path)
		return os.Rename(tmpPath, db.path)
	}
	return nil
}

func (db *Database) CreateTable(name string) {
	db.mu.Lock()
	defer db.mu.Unlock()
	if _, exists := db.tables[name]; !exists {
		db.tables[name] = make([]Record, 0)
	}
}

func (db *Database) Insert(table string, rec Record) Record {
	db.mu.Lock()
	defer db.mu.Unlock()
	if _, exists := db.tables[table]; !exists {
		db.tables[table] = make([]Record, 0)
	}
	if _, hasID := rec["id"]; !hasID {
		maxID := 0
		for _, r := range db.tables[table] {
			if id, ok := r["id"].(float64); ok {
				if int(id) > maxID {
					maxID = int(id)
				}
			}
		}
		rec["id"] = float64(maxID + 1)
	}
	db.tables[table] = append(db.tables[table], rec)
	_ = db.saveLocked()
	return rec
}

func (db *Database) Update(table string, where map[string]interface{}, updates Record) int {
	db.mu.Lock()
	defer db.mu.Unlock()
	records, exists := db.tables[table]
	if !exists {
		return 0
	}
	count := 0
	for i, rec := range records {
		if matchesWhere(rec, where) {
			for k, v := range updates {
				rec[k] = v
			}
			records[i] = rec
			count++
		}
	}
	_ = db.saveLocked()
	return count
}

func (db *Database) Delete(table string, where map[string]interface{}) int {
	db.mu.Lock()
	defer db.mu.Unlock()
	records, exists := db.tables[table]
	if !exists {
		return 0
	}
	remaining := make([]Record, 0, len(records))
	count := 0
	for _, rec := range records {
		if matchesWhere(rec, where) {
			count++
		} else {
			remaining = append(remaining, rec)
		}
	}
	db.tables[table] = remaining
	_ = db.saveLocked()
	return count
}

func (db *Database) Query(table string, where map[string]interface{}, limit int) []Record {
	db.mu.RLock()
	defer db.mu.RUnlock()
	records, exists := db.tables[table]
	if !exists {
		return make([]Record, 0)
	}
	results := make([]Record, 0)
	for _, rec := range records {
		if matchesWhere(rec, where) {
			results = append(results, copyRecord(rec))
		}
	}
	sort.Slice(results, func(i, j int) bool {
		return getRecordID(results[i]) < getRecordID(results[j])
	})
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}
	return results
}

func (db *Database) Count(table string, where map[string]interface{}) int {
	db.mu.RLock()
	defer db.mu.RUnlock()
	records, exists := db.tables[table]
	if !exists {
		return 0
	}
	count := 0
	for _, rec := range records {
		if matchesWhere(rec, where) {
			count++
		}
	}
	return count
}

func (db *Database) GetRecord(table string, where map[string]interface{}) (Record, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	records, exists := db.tables[table]
	if !exists {
		return nil, false
	}
	for _, rec := range records {
		if matchesWhere(rec, where) {
			return copyRecord(rec), true
		}
	}
	return nil, false
}

func (db *Database) TableNames() []string {
	db.mu.RLock()
	defer db.mu.RUnlock()
	names := make([]string, 0, len(db.tables))
	for name := range db.tables {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (db *Database) Close() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.saveLocked()
}

func (db *Database) Table(name string) []Record {
	db.mu.RLock()
	defer db.mu.RUnlock()
	records, exists := db.tables[name]
	if !exists {
		return make([]Record, 0)
	}
	results := make([]Record, len(records))
	for i, r := range records {
		results[i] = copyRecord(r)
	}
	return results
}

func matchesWhere(rec Record, where map[string]interface{}) bool {
	for field, expected := range where {
		actual, exists := rec[field]
		if !exists {
			return false
		}
		if fmt.Sprintf("%v", actual) != fmt.Sprintf("%v", expected) {
			return false
		}
	}
	return true
}

func copyRecord(rec Record) Record {
	c := make(Record, len(rec))
	for k, v := range rec {
		c[k] = v
	}
	return c
}

func getRecordID(rec Record) int {
	if id, ok := rec["id"]; ok {
		if f, ok := id.(float64); ok {
			return int(f)
		}
	}
	return 0
}
