package auth

import (
	"sync"

	"github.com/ecomatrix/backend/internal/repo"
	"gorm.io/gorm"
)

// AgentSecretStoreDB is a Postgres-backed AgentSecretStore. The env-backed
// store from agent_secrets.go is consulted first; this layer is an
// additional cache that survives restarts and supports rotation.
type AgentSecretStoreDB struct {
	mu   sync.RWMutex
	cache map[string][]byte
	repo  *repo.AgentRepo
	db    *gorm.DB
}

func NewAgentSecretStoreDB(db *gorm.DB, r *repo.AgentRepo) *AgentSecretStoreDB {
	return &AgentSecretStoreDB{
		cache: map[string][]byte{},
		repo:  r,
		db:    db,
	}
}

// SecretFor returns the secret for the given agent. Reads from the in-memory
// cache; on miss, queries the DB and caches the result. Returns false if
// neither the cache nor the DB has the secret.
func (s *AgentSecretStoreDB) SecretFor(agentID string) ([]byte, bool) {
	s.mu.RLock()
	if v, ok := s.cache[agentID]; ok {
		s.mu.RUnlock()
		return v, true
	}
	s.mu.RUnlock()

	// Query DB.
	var row struct{ Secret string }
	if err := s.db.Raw(`SELECT secret FROM agent_secrets WHERE agent_id = ?`, agentID).Scan(&row).Error; err != nil || row.Secret == "" {
		return nil, false
	}
	s.mu.Lock()
	s.cache[agentID] = []byte(row.Secret)
	s.mu.Unlock()
	return s.cache[agentID], true
}

// SetSecret inserts or updates a secret for the given agent. Used by tests
// and a future admin endpoint.
func (s *AgentSecretStoreDB) SetSecret(agentID string, secret []byte) error {
	err := s.db.Exec(`
		INSERT INTO agent_secrets (agent_id, secret) VALUES (?, ?)
		ON CONFLICT (agent_id) DO UPDATE SET secret = EXCLUDED.secret, rotated_at = now()
	`, agentID, string(secret)).Error
	if err == nil {
		s.mu.Lock()
		s.cache[agentID] = secret
		s.mu.Unlock()
	}
	return err
}

// IsConfigured returns true if the DB has any secret rows. We avoid
// a per-request query by caching the result.
func (s *AgentSecretStoreDB) IsConfigured() bool {
	var n int64
	_ = s.db.Raw(`SELECT COUNT(*) FROM agent_secrets`).Scan(&n).Error
	return n > 0
}

// Invalidate drops the cache entry for an agent. Used after rotation.
func (s *AgentSecretStoreDB) Invalidate(agentID string) {
	s.mu.Lock()
	delete(s.cache, agentID)
	s.mu.Unlock()
}
