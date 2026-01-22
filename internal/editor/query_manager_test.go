package editor

import (
	"testing"
)

func TestQueryManagerSaveQuery(t *testing.T) {
	qm := NewQueryManager()
	sq := qm.SaveQuery("test", "desc", "SELECT * FROM users", "queries", []string{"select"})

	if sq.ID == "" {
		t.Errorf("Expected query ID")
	}
	if sq.Name != "test" {
		t.Errorf("Expected name 'test', got %s", sq.Name)
	}
}

func TestQueryManagerGetQuery(t *testing.T) {
	qm := NewQueryManager()
	sq := qm.SaveQuery("test", "desc", "SELECT * FROM users", "queries", []string{})

	retrieved := qm.GetQuery(sq.ID)
	if retrieved == nil {
		t.Errorf("Expected to retrieve query")
		return
	}
	if retrieved.Name != "test" {
		t.Errorf("Expected name 'test'")
	}
}

func TestQueryManagerDeleteQuery(t *testing.T) {
	qm := NewQueryManager()
	sq := qm.SaveQuery("test", "desc", "SELECT * FROM users", "queries", []string{})

	deleted := qm.DeleteQuery(sq.ID)
	if !deleted {
		t.Errorf("Expected successful deletion")
	}

	retrieved := qm.GetQuery(sq.ID)
	if retrieved != nil {
		t.Errorf("Expected query to be deleted")
	}
}

func TestQueryManagerListQueries(t *testing.T) {
	qm := NewQueryManager()
	qm.SaveQuery("test1", "desc1", "SELECT * FROM users", "queries", []string{})
	qm.SaveQuery("test2", "desc2", "SELECT * FROM orders", "queries", []string{})

	queries := qm.ListQueries()
	if len(queries) != 2 {
		t.Errorf("Expected 2 queries, got %d", len(queries))
	}
}

func TestQueryManagerUpdateQuery(t *testing.T) {
	qm := NewQueryManager()
	sq := qm.SaveQuery("test", "desc", "SELECT * FROM users", "queries", []string{})

	updated := qm.UpdateQuery(sq.ID, "updated", "new desc", "SELECT id FROM users", "category", []string{"updated"})
	if updated.Name != "updated" {
		t.Errorf("Expected name to be updated")
	}
	if updated.Query != "SELECT id FROM users" {
		t.Errorf("Expected query to be updated")
	}
}
