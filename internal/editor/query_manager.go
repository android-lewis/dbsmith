package editor

import (
	"fmt"
	"time"
)

type SavedQuery struct {
	ID             string
	Name           string
	Description    string
	Query          string
	Category       string
	Tags           []string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	ExecutionCount int
	LastExecutedAt time.Time
	Hash           string
}

type QueryManager struct {
	queries      map[string]*SavedQuery
	queryHistory []ExecutionRecord
	maxHistory   int
}

type ExecutionRecord struct {
	QueryID      string
	Query        string
	ExecutedAt   time.Time
	Duration     int64
	RowsAffected int64
	Error        string
}

func NewQueryManager() *QueryManager {
	return &QueryManager{
		queries:      make(map[string]*SavedQuery),
		queryHistory: make([]ExecutionRecord, 0),
		maxHistory:   100,
	}
}

func (qm *QueryManager) SaveQuery(name, description, query, category string, tags []string) *SavedQuery {
	id := fmt.Sprintf("query_%d", len(qm.queries))
	sq := &SavedQuery{
		ID:          id,
		Name:        name,
		Description: description,
		Query:       query,
		Category:    category,
		Tags:        tags,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	qm.queries[id] = sq
	return sq
}

func (qm *QueryManager) UpdateQuery(id, name, description, query, category string, tags []string) *SavedQuery {
	sq, exists := qm.queries[id]
	if !exists {
		return nil
	}

	sq.Name = name
	sq.Description = description
	sq.Query = query
	sq.Category = category
	sq.Tags = tags
	sq.UpdatedAt = time.Now()

	return sq
}

func (qm *QueryManager) GetQuery(id string) *SavedQuery {
	return qm.queries[id]
}

func (qm *QueryManager) DeleteQuery(id string) bool {
	_, exists := qm.queries[id]
	if !exists {
		return false
	}
	delete(qm.queries, id)
	return true
}

func (qm *QueryManager) ListQueries() []*SavedQuery {
	queries := make([]*SavedQuery, 0, len(qm.queries))
	for _, q := range qm.queries {
		queries = append(queries, q)
	}
	return queries
}
