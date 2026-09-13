package pos_test

import (
	"os"
	"testing"

	"github.com/joysriramsarkar/alap-framework/stdlib/database"
)

func TestPOSDatabaseDirect(t *testing.T) {
	tmp := "/tmp/alap_pos_test.db"
	os.Remove(tmp)
	db, err := database.New(tmp)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	defer db.Close()

	db.CreateTable("products")
	db.Insert("products", database.Record{"id": float64(1), "name": "Apple", "price": float64(150), "qty": float64(100)})
	db.Insert("products", database.Record{"id": float64(2), "name": "Bread", "price": float64(45), "qty": float64(50)})

	if db.Count("products", map[string]interface{}{}) != 2 {
		t.Error("expected 2 products")
	}
	rec, found := db.GetRecord("products", map[string]interface{}{"name": "Apple"})
	if !found {
		t.Fatal("expected Apple")
	}
	if rec["price"] != float64(150) {
		t.Errorf("expected 150, got %v", rec["price"])
	}
	updated := db.Update("products", map[string]interface{}{"name": "Apple"}, database.Record{"qty": float64(97)})
	if updated != 1 {
		t.Error("expected 1 update")
	}
	deleted := db.Delete("products", map[string]interface{}{"name": "Bread"})
	if deleted != 1 {
		t.Error("expected 1 delete")
	}
}

func TestPOSDatabasePersistence(t *testing.T) {
	tmp := "/tmp/alap_pos_persist.db"
	os.Remove(tmp)
	db1, err := database.New(tmp)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	db1.CreateTable("items")
	db1.Insert("items", database.Record{"id": float64(1), "name": "Widget"})
	db1.Close()

	db2, err := database.New(tmp)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	defer db2.Close()
	defer os.Remove(tmp)

	if db2.Count("items", map[string]interface{}{}) != 1 {
		t.Error("expected 1 item after reload")
	}
	_, found := db2.GetRecord("items", map[string]interface{}{"name": "Widget"})
	if !found {
		t.Error("expected Widget to persist")
	}
}

func TestPOSCheckoutFlow(t *testing.T) {
	tmp := "/tmp/alap_pos_checkout.db"
	os.Remove(tmp)
	db, err := database.New(tmp)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	defer db.Close()
	defer os.Remove(tmp)

	db.CreateTable("products")
	db.CreateTable("cart")
	db.CreateTable("transactions")

	products := []map[string]interface{}{
		{"id": float64(1), "name": "Apple", "price": float64(150), "qty": float64(100)},
		{"id": float64(2), "name": "Bread", "price": float64(45), "qty": float64(50)},
	}
	for _, p := range products {
		db.Insert("products", database.Record(p))
	}
	if db.Count("products", map[string]interface{}{}) != 2 {
		t.Error("expected 2 products")
	}

	db.Insert("cart", database.Record{"product_id": float64(1), "qty": float64(3), "line_total": float64(300)})
	db.Insert("cart", database.Record{"product_id": float64(2), "qty": float64(2), "line_total": float64(100)})

	cartItems := db.Query("cart", map[string]interface{}{}, 0)
	cartTotal := 0.0
	for _, item := range cartItems {
		if total, ok := item["line_total"].(float64); ok {
			cartTotal += total
		}
	}
	if cartTotal != 400 {
		t.Errorf("expected 400, got %v", cartTotal)
	}

	db.Update("products", map[string]interface{}{"id": float64(1)}, database.Record{"qty": float64(47)})
	apple, _ := db.GetRecord("products", map[string]interface{}{"id": float64(1)})
	if apple["qty"] != float64(47) {
		t.Errorf("expected 47, got %v", apple["qty"])
	}
}
