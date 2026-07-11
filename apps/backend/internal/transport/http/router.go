// Package http wires HTTP handlers and middleware.
package http

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/ecomatrix/backend/internal/domain"
	"github.com/ecomatrix/backend/internal/repo"
	"github.com/ecomatrix/backend/internal/service"
	"github.com/ecomatrix/backend/internal/transport/ws"
	"github.com/ecomatrix/backend/pkg/a2a"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

type Server struct {
	App    *fiber.App
	Agents *repo.AgentRepo
	Txs    *repo.TxRepo
	Trade  *service.TradeService
	Hub    *ws.Hub
	Log    *slog.Logger
	Admin  string
	DB     *sql.DB
}

func (s *Server) Register() {
	s.App.Use(requestIDMiddleware())
	s.App.Use(loggingMiddleware(s.Log))

	s.App.Get("/healthz", s.healthz)
	s.App.Get("/readyz", s.readyz)
	s.App.Get("/v1/stream", websocket.New(s.streamHandler))

	v1 := s.App.Group("/v1")
	v1.Get("/agents", s.listAgents)
	// Order matters: more specific route first.
	v1.Get("/agents/by-string-id/:sid", s.getAgentByStringID)
	v1.Get("/agents/:id", s.getAgent)
	v1.Post("/agents", s.requireAdmin, s.createAgent)
	v1.Post("/trades", s.postTrade)
	v1.Get("/transactions", s.listTransactions)
}

// ---------- middleware ----------

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

func isValidJobType(j string) bool {
	switch j {
	case "miner", "merchant", "hacker", "mediator":
		return true
	}
	return false
}
