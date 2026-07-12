// Package http wires HTTP handlers and middleware.
package http

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/ecomatrix/backend/internal/auth"
	"github.com/ecomatrix/backend/internal/domain"
	"github.com/ecomatrix/backend/internal/repo"
	"github.com/ecomatrix/backend/internal/service"
	"github.com/ecomatrix/backend/internal/transport/ws"
	"github.com/ecomatrix/backend/pkg/a2a"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

type Server struct {
	App       *fiber.App
	Agents    *repo.AgentRepo
	Txs       *repo.TxRepo
	Feed      *repo.FeedRepo
	Trade     *service.TradeService
	Metrics   *service.MetricsService
	Hub       *ws.Hub
	Log       *slog.Logger
	Admin     string
	DB            *sql.DB
	CORS          corsConfig
	AuthStore     auth.AgentSecretStore
	RateLimit     *auth.RateLimiter
	Conversations *repo.ConversationsRepo
	LLMCache      *repo.LLMCacheRepo
}

type corsConfig struct {
	AllowedOrigins []string // empty = no CORS; "*" = wildcard; otherwise exact match
	DevMode        bool
}

func (s *Server) Register() {
	s.App.Use(requestIDMiddleware())
	s.App.Use(corsMiddleware(s.CORS, s.Log))
	s.App.Use(auth.RequireAgentSignature(s.AuthStore))
	s.App.Use(loggingMiddleware(s.Log))

	s.App.Get("/healthz", s.healthz)
	s.App.Get("/readyz", s.readyz)
	s.App.Get("/v1/stream", websocket.New(s.streamHandler))

	v1 := s.App.Group("/v1")
	v1.Get("/agents", s.listAgents)
	// Order matters: more specific route first.
	v1.Get("/agents/by-string-id/:sid", s.getAgentByStringID)
	v1.Get("/agents/by-string-id/:sid/long-term-memory", s.getAgentLTM)
	v1.Get("/agents/by-string-id/:sid/conversations", s.getAgentConversations)
	v1.Put("/agents/by-string-id/:sid/long-term-memory", s.putAgentLTM)
	v1.Get("/agents/:id", s.getAgent)
	v1.Post("/agents", s.requireAdmin, s.createAgent)
	v1.Post("/trades", s.rateLimit("EXECUTE_TRADE"), s.postTrade)
	v1.Get("/transactions", s.listTransactions)
	v1.Get("/feeds", s.listFeeds)
	v1.Post("/feeds", s.rateLimit("POST_FEED"), s.postFeed)
	v1.Get("/metrics", s.getMetrics)
	v1.Get("/metrics/history", s.getMetricsHistory)
	v1.Get("/llm-cache/stats", s.getLLMCacheStats)
}

// ---------- middleware ----------

// corsMiddleware enforces an origin allowlist configured via
// ECOMATRIX_CORS_ALLOWED_ORIGINS (comma-separated).
//
//   - "*"            -> wildcard (any origin)
//   - "https://a,https://b" -> exact match
//   - unset + ECOMATRIX_DEV=true -> defaults to ["*"] for local dev
//   - unset + ECOMATRIX_DEV != true -> no CORS headers (effectively locked down)
//
// Preflight requests with disallowed origins get a 403 instead of the
// permissive silent pass the previous version emitted.
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

func requestIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		rid := c.Get("X-Request-Id")
		if rid == "" {
			rid = strconv.FormatInt(time.Now().UnixNano(), 36)
		}
		c.Locals("request_id", rid)
		c.Set("X-Request-Id", rid)
		return c.Next()
	}
}

func loggingMiddleware(log *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		log.Info("http",
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", c.Response().StatusCode()),
			slog.Int("latency_ms", int(time.Since(start).Milliseconds())),
			slog.String("request_id", c.Locals("request_id").(string)),
		)
		return err
	}
}

func (s *Server) requireAdmin(c *fiber.Ctx) error {
	if c.Get("X-Admin-Token") != s.Admin {
		return fiber.NewError(fiber.StatusUnauthorized, "admin token required")
	}
	return c.Next()
}

// ---------- handlers ----------

func (s *Server) healthz(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}

func (s *Server) readyz(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
	defer cancel()
	if s.DB == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"status": "db_unreachable"})
	}
	if err := s.DB.PingContext(ctx); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"status": "db_unreachable"})
	}
	return c.JSON(fiber.Map{"status": "ready", "ws_conns": s.Hub.ConnCount()})
}

func (s *Server) listAgents(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	rows, err := s.Agents.List(c.Context(), limit, offset)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"agents": rows})
}

func (s *Server) getAgentByStringID(c *fiber.Ctx) error {
	sid := c.Params("sid")
	if sid == "" {
		return fiber.NewError(fiber.StatusBadRequest, "sid is required")
	}
	a, err := s.Agents.ByStringID(c.Context(), sid)
	if err != nil {
		if errors.Is(err, domain.ErrAgentNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "agent not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(a)
}

func (s *Server) getAgentLTM(c *fiber.Ctx) error {
	sid := c.Params("sid")
	if sid == "" {
		return fiber.NewError(fiber.StatusBadRequest, "sid is required")
	}
	a, err := s.Agents.ByStringID(c.Context(), sid)
	if err != nil {
		if errors.Is(err, domain.ErrAgentNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "agent not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	ltm, err := s.Agents.GetLongTermMemory(c.Context(), a.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"string_id": sid, "long_term_memory": ltm})
}

func (s *Server) putAgentLTM(c *fiber.Ctx) error {
	sid := c.Params("sid")
	if sid == "" {
		return fiber.NewError(fiber.StatusBadRequest, "sid is required")
	}
	var body struct {
		Summary string   `json:"summary"`
		Facts   []string `json:"facts"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid json")
	}
	if len(body.Summary) > 500 {
		return fiber.NewError(fiber.StatusBadRequest, "summary must be <= 500 chars")
	}
	if len(body.Facts) > 50 {
		return fiber.NewError(fiber.StatusBadRequest, "facts must be <= 50 entries")
	}
	a, err := s.Agents.ByStringID(c.Context(), sid)
	if err != nil {
		if errors.Is(err, domain.ErrAgentNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "agent not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	ltm := domain.LongTermMemory{Summary: body.Summary, Facts: body.Facts}
	if err := s.Agents.SetLongTermMemory(c.Context(), a.ID, ltm); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"string_id": sid, "long_term_memory": ltm})
}

func (s *Server) getAgentConversations(c *fiber.Ctx) error {
	sid := c.Params("sid")
	if sid == "" {
		return fiber.NewError(fiber.StatusBadRequest, "sid is required")
	}
	if _, err := s.Agents.ByStringID(c.Context(), sid); err != nil {
		if errors.Is(err, domain.ErrAgentNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "agent not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	convs, err := s.Conversations.Recent(c.Context(), sid, limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"conversations": convs})
}

func (s *Server) getAgent(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "id must be integer")
	}
	a, err := s.Agents.ByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrAgentNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "agent not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(a)
}

func (s *Server) createAgent(c *fiber.Ctx) error {
	var body struct {
		StringID    string `json:"string_id"`
		JobType     string `json:"job_type"`
		Balance     int64  `json:"balance"`
		Vitality    int    `json:"vitality"`
		CreditScore int    `json:"credit_score"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid json")
	}
	if !a2a.ValidateAgentID(body.StringID) {
		return fiber.NewError(fiber.StatusBadRequest, "string_id must match agent_*")
	}
	if !isValidJobType(body.JobType) {
		return fiber.NewError(fiber.StatusBadRequest, "job_type invalid")
	}
	if body.Balance < 0 {
		return fiber.NewError(fiber.StatusBadRequest, "balance must be >= 0")
	}
	if body.Vitality < 0 || body.Vitality > 100 {
		return fiber.NewError(fiber.StatusBadRequest, "vitality must be 0..100")
	}
	if body.CreditScore < 0 || body.CreditScore > 100 {
		return fiber.NewError(fiber.StatusBadRequest, "credit_score must be 0..100")
	}
	a, err := s.Agents.Create(c.Context(), repo.NewAgentInput(body.StringID, body.JobType, body.Balance, body.Vitality, body.CreditScore))
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(a)
}

func (s *Server) postTrade(c *fiber.Ctx) error {
	var env a2a.Envelope
	if err := c.BodyParser(&env); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(a2aEnvelopeError(a2a.New(a2a.CodeInvalidEnvelope, "body is not valid JSON", false), ""))
	}
	if err := a2a.Validate(env); err != nil {
		if e, ok := a2a.As(err); ok {
			return c.Status(httpStatusForCode(e.Code)).JSON(a2aEnvelopeError(e, env.MsgID))
		}
		return c.Status(fiber.StatusBadRequest).JSON(a2aEnvelopeError(a2a.New(a2a.CodeInvalidEnvelope, err.Error(), false), env.MsgID))
	}
	if env.Action != a2a.ActionExecuteTrade {
		return c.Status(fiber.StatusBadRequest).JSON(a2aEnvelopeError(
			a2a.New(a2a.CodeUnknownAction, "only EXECUTE_TRADE is supported in Phase 1", false), env.MsgID))
	}
	payload, err := a2a.DecodeTradePayload(env.Payload)
	if err != nil {
		if e, ok := a2a.As(err); ok {
			return c.Status(httpStatusForCode(e.Code)).JSON(a2aEnvelopeError(e, env.MsgID))
		}
		return c.Status(fiber.StatusBadRequest).JSON(a2aEnvelopeError(a2a.New(a2a.CodeInvalidEnvelope, err.Error(), false), env.MsgID))
	}
	res, aerr := s.Trade.Settle(c.Context(), env, payload)
	if aerr != nil {
		return c.Status(httpStatusForCode(aerr.Code)).JSON(a2aEnvelopeError(aerr, env.MsgID))
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"protocol_v": a2a.ProtocolVersion,
		"msg_id":     env.MsgID,
		"status":     "SETTLED",
		"tx_id":      res.Receipt.TxID,
		"receipt":    res.Receipt,
		"replay":     res.Replay,
	})
}

func (s *Server) LLMCacheStats(ctx context.Context) (interface{}, error) {
	return s.LLMCache.Stats(ctx)
}

func (s *Server) rateLimit(action string) fiber.Handler {
	actionVal := a2a.Action(action)
	return func(c *fiber.Ctx) error {
		c.Locals("a2a_action", actionVal)
		if s.RateLimit == nil {
			return c.Next()
		}
		// We don't know the agent yet (HMAC runs after this in Register); we
		// use the X-Agent-Id header as a soft key. After HMAC verifies it,
		// the AuthStore can also re-check.
		sender := c.Get("X-Agent-Id")
		if sender == "" {
			sender = "anonymous"
		}
		if !s.RateLimit.Allow(sender + "|" + action) {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": fiber.Map{
					"code":     a2a.CodeRateLimited,
					"message":  "rate limit exceeded",
					"retryable": true,
				},
			})
		}
		return c.Next()
	}
}

func (s *Server) getMetrics(c *fiber.Ctx) error {
	snap, err := s.Metrics.Collect(c.Context(), s.Hub.ConnCount())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(snap)
}

func (s *Server) getMetricsHistory(c *fiber.Ctx) error {
	history := s.Metrics.History()
	return c.JSON(fiber.Map{
		"window_seconds": 60,
		"capacity":       120,
		"count":          len(history),
		"samples":        history,
	})
}

func (s *Server) getLLMCacheStats(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
	defer cancel()
	stats, err := s.LLMCacheStats(ctx)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(stats)
}

func (s *Server) listFeeds(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	rows, err := s.Feed.Recent(c.Context(), limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"feeds": rows})
}

func (s *Server) postFeed(c *fiber.Ctx) error {
	var env a2a.Envelope
	if err := c.BodyParser(&env); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(a2aEnvelopeError(a2a.New(a2a.CodeInvalidEnvelope, "body is not valid JSON", false), ""))
	}
	if err := a2a.Validate(env); err != nil {
		if e, ok := a2a.As(err); ok {
			return c.Status(httpStatusForCode(e.Code)).JSON(a2aEnvelopeError(e, env.MsgID))
		}
		return c.Status(fiber.StatusBadRequest).JSON(a2aEnvelopeError(a2a.New(a2a.CodeInvalidEnvelope, err.Error(), false), env.MsgID))
	}
	if env.Action != a2a.ActionPostFeed {
		return c.Status(fiber.StatusBadRequest).JSON(a2aEnvelopeError(
			a2a.New(a2a.CodeUnknownAction, "POST /v1/feeds requires POST_FEED action", false), env.MsgID))
	}
	payload, err := a2a.DecodeFeedPayload(env.Payload)
	if err != nil {
		if e, ok := a2a.As(err); ok {
			return c.Status(httpStatusForCode(e.Code)).JSON(a2aEnvelopeError(e, env.MsgID))
		}
		return c.Status(fiber.StatusBadRequest).JSON(a2aEnvelopeError(a2a.New(a2a.CodeInvalidEnvelope, err.Error(), false), env.MsgID))
	}
	sender, err := s.Agents.ByStringID(c.Context(), env.Sender)
	if err != nil {
		if errors.Is(err, domain.ErrAgentNotFound) {
			return c.Status(httpStatusForCode(a2a.CodeUnknownAgent)).JSON(a2aEnvelopeError(a2a.New(a2a.CodeUnknownAgent, "sender unknown", false), env.MsgID))
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	post, err := s.Feed.Insert(c.Context(), domain.FeedPost{
		AgentID:    sender.ID,
		Content:    payload.Content,
		IntentType: payload.IntentType,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	s.Hub.Publish(map[string]any{
		"type":        "feed.posted",
		"post_id":     post.ID,
		"agent_id":    sender.StringID,
		"intent_type": post.IntentType,
		"content":     post.Content,
	})
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"protocol_v": a2a.ProtocolVersion,
		"msg_id":     env.MsgID,
		"status":     "POSTED",
		"post_id":    post.ID,
	})
}

func (s *Server) listTransactions(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := s.Txs.Recent(c.Context(), limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"transactions": rows})
}

// ---------- ws ----------

func (s *Server) streamHandler(c *websocket.Conn) {
	release := s.Hub.Add(c)
	defer release()
	for {
		if _, _, err := c.ReadMessage(); err != nil {
			return
		}
	}
}

// ---------- helpers ----------

func a2aEnvelopeError(e *a2a.Error, msgID string) fiber.Map {
	return fiber.Map{
		"protocol_v": a2a.ProtocolVersion,
		"msg_id":     msgID,
		"status":     "REJECTED",
		"error": fiber.Map{
			"code":      string(e.Code),
			"message":   e.Message,
			"retryable": e.Retryable,
		},
	}
}

func httpStatusForCode(c a2a.Code) int {
	switch c {
	case a2a.CodeProtocolMismatch, a2a.CodeInvalidEnvelope, a2a.CodeUnknownAction:
		return fiber.StatusBadRequest
	case a2a.CodeUnknownAgent:
		return fiber.StatusNotFound
	case a2a.CodeInsufficientFunds, a2a.CodeSelfTrade:
		return fiber.StatusUnprocessableEntity
	case a2a.CodeRateLimited:
		return fiber.StatusTooManyRequests
	default:
		return fiber.StatusInternalServerError
	}
}

// loadCORSOrigins reads the allowlist from the environment. Returns nil
// (no CORS) when ECOMATRIX_DEV != true and no origins are configured.
func loadCORSOrigins() []string {
	raw := os.Getenv("ECOMATRIX_CORS_ALLOWED_ORIGINS")
	dev := os.Getenv("ECOMATRIX_DEV") == "true"
	if raw == "" {
		if dev {
			return []string{"*"}
		}
		return nil
	}
	var out []string
	for _, o := range strings.Split(raw, ",") {
		o = strings.TrimSpace(o)
		if o != "" {
			out = append(out, o)
		}
	}
	return out
}

func isValidJobType(j string) bool {
	switch j {
	case "miner", "merchant", "hacker", "mediator":
		return true
	}
	return false
}

// rateLimitMiddleware caps the rate of state-mutating A2A requests per
// (sender, action). It is applied AFTER HMAC verification, so the bucket
// key is the legitimate agent identity.
func rateLimitMiddleware(rl *auth.RateLimiter) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if rl == nil {
			return c.Next()
		}
		action := string(c.Locals("a2a_action").(a2a.Action))
		sender := c.Locals("agent_id").(string)
		if !rl.Allow(sender + "|" + action) {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    a2a.CodeRateLimited,
					"message": "rate limit exceeded",
					"retryable": true,
				},
			})
		}
		return c.Next()
	}
}

// CORSConfigFromConfig returns a corsConfig populated from the process
// environment. Used by cmd/server/main.go.
func CORSConfigFromConfig() corsConfig {
	dev := os.Getenv("ECOMATRIX_DEV") == "true"
	return corsConfig{
		AllowedOrigins: loadCORSOrigins(),
		DevMode:        dev,
	}
}
