// Command server is the EcoMatrix backend entrypoint.
package main

import (
	"errors"
	"log/slog"
	"time"
	"os"

	"github.com/ecomatrix/backend/internal/config"
	"github.com/ecomatrix/backend/internal/observability"
	"github.com/ecomatrix/backend/internal/repo"
	"github.com/ecomatrix/backend/internal/service"
	httpx "github.com/ecomatrix/backend/internal/transport/http"
	"github.com/ecomatrix/backend/internal/transport/ws"
	"github.com/gofiber/fiber/v2"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}
	log := observability.NewLogger(cfg.LogLevel)

	db, err := repo.Open(cfg.DBDSN)
	if err != nil {
		log.Error("open db", "err", err)
		os.Exit(1)
	}
	if err := repo.Migrate(db); err != nil {
		log.Error("migrate", "err", err)
		os.Exit(1)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Error("get sql.DB", "err", err)
		os.Exit(1)
	}

	hub := ws.NewHub(cfg.WSHubBuffer, time.Duration(cfg.WSHeartbeatS)*time.Second)
	agents := repo.NewAgentRepo(db)
	txs := repo.NewTxRepo(db)
	trade := service.NewTradeService(db, agents, txs, hub)

	app := fiber.New(fiber.Config{
		AppName:               "ecomatrix-backend",
		DisableStartupMessage: true,
		ErrorHandler:          fiberErrorHandler(log),
	})
	srv := &httpx.Server{
		App: app, Agents: agents, Txs: txs, Trade: trade, Hub: hub,
		Log: log, Admin: cfg.AdminToken, DB: sqlDB,
	}
	srv.Register()

	log.Info("listening", "addr", cfg.HTTPAddr)
	if err := app.Listen(cfg.HTTPAddr); err != nil {
		log.Error("listen", "err", err)
		os.Exit(1)
	}
}

func fiberErrorHandler(log *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		msg := "internal error"
		var fe *fiber.Error
		if errors.As(err, &fe) {
			code = fe.Code
			msg = fe.Message
		}
		if code >= 500 {
			log.Error("http.error", "path", c.Path(), "err", err.Error())
		}
		return c.Status(code).JSON(fiber.Map{"error": msg})
	}
}
