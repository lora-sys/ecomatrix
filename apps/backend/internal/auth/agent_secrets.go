package auth

import (
	"os"
	"strings"
	"sync"
)

// AgentSecretStore is the contract for HMAC secret lookup. Both the
// env-backed and DB-backed implementations satisfy it.
type AgentSecretStore interface {
	SecretFor(agentID string) ([]byte, bool)
	// IsConfigured reports whether the store has at least one secret known.
	// In dev mode (no env, empty DB), this returns false and the middleware
	// becomes a no-op so the dashboard still works without HMAC.
	IsConfigured() bool
}

// EnvAgentSecretStore keeps secrets in process memory, sourced from env.
type EnvAgentSecretStore struct {
	mu      sync.RWMutex
	byAgent map[string][]byte
}

func NewAgentSecretStoreFromEnv() *EnvAgentSecretStore {
	raw := os.Getenv("ECOMATRIX_AGENT_SECRETS")
	s := &EnvAgentSecretStore{byAgent: map[string][]byte{}}
	if raw == "" {
		return s
	}
	for _, pair := range strings.Split(raw, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		eq := strings.IndexByte(pair, '=')
		if eq <= 0 {
			continue
		}
		id := strings.TrimSpace(pair[:eq])
		secret := strings.TrimSpace(pair[eq+1:])
		if id != "" && secret != "" {
			s.byAgent[id] = []byte(secret)
		}
	}
	return s
}

func (s *EnvAgentSecretStore) SecretFor(agentID string) ([]byte, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	secret, ok := s.byAgent[agentID]
	return secret, ok
}

func (s *EnvAgentSecretStore) IsConfigured() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.byAgent) > 0
}

func (s *EnvAgentSecretStore) SetSecret(agentID string, secret []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byAgent[agentID] = secret
}

// Compile-time check.
var _ AgentSecretStore = (*EnvAgentSecretStore)(nil)
var _ AgentSecretStore = (*AgentSecretStoreDB)(nil)
