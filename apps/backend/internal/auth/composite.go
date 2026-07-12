package auth

import "sync"

// CompositeStore tries each underlying store in order. Returns the first
// match. IsConfigured is true if any underlying store is configured.
type CompositeStore struct {
	mu     sync.RWMutex
	stores []AgentSecretStore
	any    bool // cached IsConfigured result; invalidated on rebuild
}

func NewCompositeStore(stores ...AgentSecretStore) *CompositeStore {
	cs := &CompositeStore{stores: stores}
	cs.recompute()
	return cs
}

func (c *CompositeStore) recompute() {
	any := false
	for _, s := range c.stores {
		if s.IsConfigured() {
			any = true
			break
		}
	}
	c.mu.Lock()
	c.any = any
	c.mu.Unlock()
}

func (c *CompositeStore) SecretFor(agentID string) ([]byte, bool) {
	c.mu.RLock()
	stores := c.stores
	c.mu.RUnlock()
	for _, s := range stores {
		if v, ok := s.SecretFor(agentID); ok {
			return v, true
		}
	}
	return nil, false
}

func (c *CompositeStore) IsConfigured() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.any
}

var _ AgentSecretStore = (*CompositeStore)(nil)
