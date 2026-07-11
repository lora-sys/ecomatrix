# Security Review (cold-start)

Date: 2026-07-11T08:17:36Z
Reviewer: security-reviewer (focused lens)

## Scope
- apps/backend/pkg/a2a (codec + errors)
- apps/backend/internal/auth (HMAC scheme)
- apps/backend/internal/transport/http (router, CORS, middleware)
- apps/backend/internal/config (env loading)
- apps/backend/cmd/server, cmd/seed
- apps/agent/ecomatrix/a2a.py (HMAC client)
- apps/frontend/lib/api.ts, hooks/, components/ (any secrets in bundle?)

## Findings

### 1. Hardcoded secrets

```
apps/backend/internal/config/config.go:73:		AdminToken:         must("ECOMATRIX_ADMIN_TOKEN", "dev-admin-token"),
```

### 2. SQL injection surface

```
apps/backend/internal/repo/db.go:150:		err := db.Raw(`SELECT version FROM schema_migrations WHERE version = ?`, v).Scan(&existing).Error
apps/backend/internal/repo/db.go:190:	if err := db.Raw(`SELECT COALESCE(MAX(version), 0) FROM schema_migrations`).Scan(&latest).Error; err != nil {
```

### 3. Path traversal / arbitrary file read

apps/backend/internal/transport/http/router.go:170:	limit, _ := strconv.Atoi(c.Query("limit", "50"))
apps/backend/internal/transport/http/router.go:171:	offset, _ := strconv.Atoi(c.Query("offset", "0"))
apps/backend/internal/transport/http/router.go:180:	sid := c.Params("sid")
apps/backend/internal/transport/http/router.go:195:	sid := c.Params("sid")
apps/backend/internal/transport/http/router.go:214:	sid := c.Params("sid")
apps/backend/internal/transport/http/router.go:222:	if err := c.BodyParser(&body); err != nil {
apps/backend/internal/transport/http/router.go:246:	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
apps/backend/internal/transport/http/router.go:268:	if err := c.BodyParser(&body); err != nil {
apps/backend/internal/transport/http/router.go:295:	if err := c.BodyParser(&env); err != nil {
apps/backend/internal/transport/http/router.go:338:	limit, _ := strconv.Atoi(c.Query("limit", "50"))

### 4. CORS middleware reflection

func corsMiddleware(cfg corsConfig, log *slog.Logger) fiber.Handler {
	allowAll := false
	allowed := map[string]struct{}{}
	for _, o := range cfg.AllowedOrigins {
		if o == "*" {
			allowAll = true
			break
		}
		allowed[o] = struct{}{}
	}
	return func(c *fiber.Ctx) error {
		origin := c.Get("Origin")
		if origin != "" {
			ok := allowAll
			if !ok {
				_, ok = allowed[origin]
			}
			if !ok {
				if c.Method() == fiber.MethodOptions {
					return c.SendStatus(fiber.StatusForbidden)
				}
				return c.Next()
			}
			c.Set("Access-Control-Allow-Origin", origin)
			c.Set("Vary", "Origin")
			c.Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
			c.Set("Access-Control-Allow-Headers", "Content-Type,X-Admin-Token,X-Request-Id")
			c.Set("Access-Control-Allow-Credentials", "true")
		}
		if c.Method() == fiber.MethodOptions {
			return c.SendStatus(fiber.StatusNoContent)
		}
		return c.Next()
	}
}

### 5. HMAC scheme review

Canonical form (Go side):
func ComputeSignature(secret []byte, method, path string, ts int64, body []byte) string {
	h := hmac.New(sha256.New, secret)
	bodyHash := sha256.Sum256(body)
	canonical := strings.Join([]string{
		method,
		path,
		strconv.FormatInt(ts, 10),
		hex.EncodeToString(bodyHash[:]),
	}, "\n")
	h.Write([]byte(canonical))
	return hex.EncodeToString(h.Sum(nil))
}

Canonical form (Python side):
def _sign(secret: bytes, method: str, path: str, ts: int, body: bytes) -> str:
    body_hash = hashlib.sha256(body).hexdigest()
    canonical = "\n".join([method.upper(), path, str(ts), body_hash]).encode("utf-8")
    return hmac.new(secret, canonical, hashlib.sha256).hexdigest()


Concerns:
- If body hash uses different encoding (Go sha256-hex vs Python hashlib.sha256().hexdigest()): check both produce same output.
- Timestamp window: MaxClockSkew = 5 min — reasonable.
- Replay window within skew: an attacker who captures a request can replay it for 5 min. Mitigation: nonce + seen-set (Phase 4).

### 6. Auth secret storage

type AgentSecretStore struct {
	mu      sync.RWMutex
	byAgent map[string][]byte
}

Concerns:
- Secrets loaded from env var; visible in /proc/<pid>/environ if process is inspectable.
- No rotation strategy.
- Secrets not bound to specific actions — once leaked, attacker can sign any action.

### 7. Rate limiting / DoS

apps/backend/cmd/server/main.go:33:	if err := repo.Migrate(db); err != nil {
apps/backend/cmd/server/main.go:34:		log.Error("migrate", "err", err)
apps/backend/cmd/seed/main.go:36:	if err := repo.Migrate(db); err != nil {
apps/backend/cmd/seed/main.go:37:		log.Error("migrate", "err", err)
apps/backend/internal/service/trade.go:23:// TradeService orchestrates A2A EXECUTE_TRADE requests.
apps/backend/internal/service/metrics.go:40:	GeneratedAt   string         `json:"generated_at"`
apps/backend/internal/service/metrics.go:122:		GeneratedAt:   time.Now().UTC().Format(time.RFC3339Nano),
apps/backend/internal/repo/agent_repo.go:22:	if err := r.db.WithContext(ctx).Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
apps/backend/internal/repo/tx_repo_list.go:12:	if err := r.db.WithContext(ctx).Order("id DESC").Limit(limit).Find(&rows).Error; err != nil {
apps/backend/internal/repo/feed_repo.go:56:	if err := r.db.WithContext(ctx).Order("id DESC").Limit(limit).Find(&rows).Error; err != nil {

Concerns:
- No rate limiting on /v1/trades or /v1/feeds. An attacker with a valid HMAC could DOS or pollute the ledger.

### 8. Logging hygiene


Concerns: verify reasoning and balances above threshold are NOT logged.

### 9. WS hub: any client can subscribe

52:	s.App.Get("/v1/stream", websocket.New(s.streamHandler))
412:func (s *Server) streamHandler(c *websocket.Conn) {

Concerns: WS endpoint is open to all origins within CORS allowlist. Could be used to scrape live state.

### 10. CORS dev-mode default is permissive

Production default (no ECOMATRIX_AGENT_SECRETS, no ECOMATRIX_DEV): no CORS headers — locked.
Dev default (ECOMATRIX_DEV unset, defaults to true): allowlist = ["*"] — permissive.
Concern: the default for ECOMATRIX_DEV is true. A misconfigured prod deploy gets permissive CORS by default.

## Verdict
Phase 3.7 HMAC + Phase 3.6 CORS allowlist close the two biggest holes.
Remaining issues are forward-looking (rate limiting, secret rotation) — Phase 4 work.
