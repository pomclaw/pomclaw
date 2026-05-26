package toolsmanager

import (
	"github.com/pomclaw/pomclaw/pkg/contracts"
	"github.com/pomclaw/pomclaw/pkg/tools"
)

// rememberAdapter adapts MemoryStore to tools.Rememberer interface.
// Works with both Oracle and PostgreSQL memory stores through the interface.
type rememberAdapter struct {
	store contracts.SqlMemoryStore
}

func (a *rememberAdapter) Remember(agentID string, text string, importance float64, category string) (string, error) {
	return a.store.Remember(agentID, text, importance, category)
}

// recallAdapter adapts MemoryStore to tools.Recaller interface.
// Works with both Oracle and PostgreSQL memory stores through the interface.
type recallAdapter struct {
	store contracts.SqlMemoryStore
}

func (a *recallAdapter) Recall(agentID string, query string, maxResults int) ([]tools.RecallResult, error) {
	memResults, err := a.store.Recall(agentID, query, maxResults)
	if err != nil {
		return nil, err
	}
	results := make([]tools.RecallResult, len(memResults))
	for i, r := range memResults {
		results[i] = tools.RecallResult{
			MemoryID:   r.MemoryID,
			Text:       r.Text,
			Importance: r.Importance,
			Category:   r.Category,
			Score:      r.Score,
		}
	}
	return results, nil
}
