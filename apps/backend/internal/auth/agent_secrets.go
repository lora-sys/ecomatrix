package auth

import (
	"os"
	"strings"
	"sync"
)

// AgentSecretStore maps agent string_id -> shared HMAC secret.
//
// In MVP we read from the environment so the secrets stay out of the
// database (the agent table doesn't carry a secret column). The format is:
//
//   ECOMATRIX_AGENT_SECRETS="agent_miner_01=s3cret-a,agent_merchant_01=s3cret-b"
//
// If an agent doesn't have a secret configured, RequireAgentSignature
// returns ErrMissingHeaders (callers fall back to admin-token auth if
// configured to do so).
type AgentSecretStore struct {
	mu      sync.RWMutex
	byAgent map[string][]byte
}

func NewAgentSecretStoreFromEnv() *AgentSecretStore {
	raw := os.Getenv("ECOMATRIX_AGENT_SECRETS")
	s := &AgentSecretStore{byAgent: map[string][]byte{}}
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

func (s *AgentSecretStore) SecretFor(agentID string) ([]byte, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	secret, ok := s.byAgent[agentID]
	return secret, ok
}

// SetSecret is a test/seed helper.
func (s *AgentSecretStore) SetSecret(agentID string, secret []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byAgent[agentID] = secret
}
