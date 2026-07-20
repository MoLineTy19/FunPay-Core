package main

import (
	"FunPay-Core/internal/engine"
	"FunPay-Core/internal/fp"
	"FunPay-Core/internal/rest"
	"context"
	"errors"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
)

func toSnapshot(a fp.Account) rest.AccountSnapshot {
	return rest.AccountSnapshot{
		UserID:   a.UserID,
		Login:    a.Login,
		Balance:  a.Balance.String(),
		LoadedAt: time.Now(),
	}
}

// fpOfferCreator адаптирует *fp.Client под rest.OfferCreator:
// конвертирует fp.OfferCreated → rest.OfferCreated (разные типы в разных пакетах).
type fpOfferCreator struct {
	c *fp.Client
}

func (f fpOfferCreator) CreateOffer(ctx context.Context, csrf, nodeID string, fields map[string]string, price decimal.Decimal, amount int, active bool) (rest.OfferCreated, error) {
	oc, err := f.c.CreateOffer(ctx, csrf, nodeID, fields, price, amount, active)
	if err != nil {
		return rest.OfferCreated{}, err
	}
	return rest.OfferCreated{NodeID: oc.NodeID, OfferID: oc.OfferID, URL: oc.URL}, nil
}

const accountRefreshInterval = 60 * time.Second

func refreshAccountLoop(ctx context.Context, client *fp.Client, srv *rest.Server, buf *engine.Buffer, prev decimal.Decimal) {
	ticker := time.NewTicker(accountRefreshInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			account, err := client.GetAccount(ctx)
			if errors.Is(err, fp.ErrAuthLost) {
				slog.Error("account refresh stopped: auth lost", "err", err)
				return
			}
			if err != nil {
				slog.Error("account refresh failed", "err", err)
				continue
			}
			if !prev.Equal(account.Balance) {
				buf.Push([]engine.Event{{
					Type: engine.AccountBalance,
					At:   time.Now(),
					Payload: engine.AccountBalancePayload{
						UserID:  account.UserID,
						Login:   account.Login,
						Balance: account.Balance.String(),
					},
				}})
				prev = account.Balance
				slog.Info("balance changed", "balance", account.Balance)
			}

			srv.SetAccount(toSnapshot(account))
			slog.Debug("account refreshed", "balance", account.Balance)
		}
	}
}

func main() {
	debug := flag.Bool("debug", false, "enable debug-level logging")
	flag.Parse()

	level := slog.LevelInfo
	if *debug {
		level = slog.LevelDebug
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})))

	err := godotenv.Load()
	if err != nil {
		slog.Error("Error loading .env file", "err", err)
		os.Exit(1)
	}

	goldenKey := os.Getenv("FP_GOLDEN_KEY")
	sessionID := os.Getenv("FP_PHPSESSID")
	goldenSeal := os.Getenv("FP_GOLDEN_SEAL")

	slog.Info("engine starting")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	client := fp.NewClient(goldenKey, sessionID, goldenSeal, 800*time.Millisecond, 600*time.Millisecond)
	account, err := client.GetAccount(ctx)
	if err != nil {
		slog.Error("get account failed", "err", err)
		return
	}
	slog.Info("account loaded", "login", account.Login, "balance", account.Balance)

	userID := os.Getenv("FP_USER_ID")
	csrfToken := os.Getenv("FP_CSRF_TOKEN")

	objectTypes := []string{"orders_counters", "chat_bookmarks"}

	runner := fp.NewRunner(client, userID, csrfToken, objectTypes)

	if err := runner.Init(ctx); err != nil {
		slog.Error("runner init failed", "err", err)
		return
	}

	buf := engine.NewBuffer()
	engineToken := os.Getenv("ENGINE_TOKEN")
	if engineToken == "" {
		slog.Error("ENGINE_TOKEN not set")
		os.Exit(1)
	}

	listenAddr := os.Getenv("ENGINE_LISTEN")
	if listenAddr == "" {
		listenAddr = "127.0.0.1:8731"
	}

	srv := rest.NewServer(buf, engineToken)
	srv.SetAccount(toSnapshot(account))
	srv.SetOfferCreator(fpOfferCreator{c: client}, csrfToken)
	slog.Info("offer creator wired")
	go refreshAccountLoop(ctx, client, srv, buf, account.Balance)
	go func() {
		if err := srv.Start(ctx, listenAddr); err != nil {
			slog.Error("rest server stopped", "err", err)
			cancel()
		}
	}()

	slog.Info("rest listening", "addr", listenAddr)
	for {
		ev, err := runner.Poll(ctx)
		if err != nil {
			if errors.Is(err, fp.ErrAuthLost) {
				slog.Error("auth lost: golden_seal expired", "err", err)
				buf.Push([]engine.Event{{
					Type: engine.EngineStatus,
					At:   time.Now(),
					Payload: engine.EngineStatusPayload{
						State: "auth_lost",
						Error: err.Error(),
					},
				}})
				srv.SetState("auth_lost")
				slog.Info("polling paused, waiting for restart or signal")
				<-ctx.Done()
				return
			}
			slog.Error("poll failed", "err", err)
			return
		}

		events := engine.WrapEvents(ev)
		buf.Push(events)
		buf.EvictExpired(time.Now())

		for _, e := range events {
			slog.Info("event", "event_id", e.EventID, "type", e.Type, "at", e.At)
		}

		slog.Debug("poll cycle", "events", len(events))
		select {
		case <-ctx.Done():
			return
		case <-time.After(2 * time.Second):
		}
	}
}
