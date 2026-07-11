package auth

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

// RequireAgentSignature returns a Fiber middleware that verifies an HMAC
// signature on every state-mutating A2A request. GETs are not gated (they
// are read-only and the dashboard polls them).
//
// Behavior:
//   - If the secret store is empty (no ECOMATRIX_AGENT_SECRETS), the
//     middleware is a no-op (dev mode).
//   - If the store has at least one secret and the request lacks valid
//     headers, it returns 401.
//   - If the request is signed but the body was tampered with, it returns
//     401.
//   - If everything checks out, the parsed X-Agent-Id is stored in c.Locals
//     for downstream handlers.
func RequireAgentSignature(store *AgentSecretStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Short-circuit when nothing is configured (dev mode).
		if store == nil || len(store.byAgent) == 0 {
			return c.Next()
		}
		agentID := c.Get(HeaderAgentId)
		ts := c.Get(HeaderAgentTimestamp)
		sig := c.Get(HeaderAgentSignature)

		// Body must be cached so handlers can still read it after we consume it.
		var body []byte
		if len(c.Body()) > 0 {
			body = append(body, c.Body()...)
		} else {
			body = []byte{}
		}

		secret, ok := store.SecretFor(agentID)
		if !ok {
			// Sender isn't a configured agent; reject.
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "agent not configured for HMAC signing",
			})
		}

		if err := Verify(secret, agentID, ts, sig, c.Method(), c.Path(), body); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
				"code":  "HMAC_INVALID",
			})
		}

		c.Locals("agent_id", agentID)
		// Restore body for downstream readers.
		c.Request().SetBody(body)
		return c.Next()
	}
}

// ErrNoSecretConfigured is returned when the store is queried for an agent
// without a secret configured.
var ErrNoSecretConfigured = errors.New("auth: no secret configured for agent")
