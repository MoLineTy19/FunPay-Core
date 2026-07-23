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
	"strconv"
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

type fpOfferCreator struct {
	c *fp.Client
}

func (f fpOfferCreator) CreateOffer(ctx context.Context, nodeID, serverID string, fields map[string]map[string]string, price decimal.Decimal, amount int, active bool) (rest.OfferCreated, error) {
	oc, err := f.c.CreateOffer(ctx, nodeID, serverID, fields, price, amount, active)
	if err != nil {
		return rest.OfferCreated{}, err
	}
	return rest.OfferCreated{NodeID: oc.NodeID, OfferID: oc.OfferID, URL: oc.URL}, nil
}

type fpOfferEditor struct {
	c *fp.Client
}

func (f fpOfferEditor) EditOffer(ctx context.Context, nodeID, offerID string, fields map[string]map[string]string, price *decimal.Decimal, amount *int, active *bool) (rest.OfferUpdated, error) {
	ou, err := f.c.EditOffer(ctx, nodeID, offerID, fields, price, amount, active)
	if err != nil {
		return rest.OfferUpdated{}, err
	}
	return rest.OfferUpdated{NodeID: ou.NodeID, OfferID: ou.OfferID, URL: ou.URL}, nil
}

type fpOfferDeleter struct {
	c *fp.Client
}

func (f fpOfferDeleter) DeleteOffer(ctx context.Context, nodeID, offerID string) (rest.OfferDeleted, error) {
	od, err := f.c.DeleteOffer(ctx, nodeID, offerID)
	if err != nil {
		return rest.OfferDeleted{}, err
	}
	return rest.OfferDeleted{NodeID: od.NodeID, OfferID: od.OfferID}, nil
}

type fpOfferLister struct {
	c *fp.Client
}

func (f fpOfferLister) ListOffers(ctx context.Context, nodeID string) ([]rest.OfferListItem, error) {
	offers, err := f.c.GetMyOffers(ctx, nodeID)
	if err != nil {
		return nil, err
	}
	items := make([]rest.OfferListItem, 0, len(offers))
	for _, o := range offers {
		items = append(items, rest.OfferListItem{
			OfferID: o.OfferID,
			Summary: o.Summary,
			Server:  o.Server,
			Amount:  o.Amount,
			Price:   o.Price,
		})
	}
	return items, nil
}

type fpOfferFormGetter struct {
	c *fp.Client
}

func (f fpOfferFormGetter) GetOfferForm(ctx context.Context, nodeID string) (rest.OfferForm, error) {
	schema, err := f.c.GetOfferForm(ctx, nodeID)
	if err != nil {
		return rest.OfferForm{}, err
	}
	fields := make([]rest.OfferFormField, 0, len(schema.Fields))
	for _, fld := range schema.Fields {
		fields = append(fields, rest.OfferFormField{ID: fld.ID, Type: int(fld.Type)})
	}
	servers := make([]rest.OfferServer, 0, len(schema.Servers))
	for _, sv := range schema.Servers {
		servers = append(servers, rest.OfferServer{ID: sv.Value, Name: sv.Label})
	}
	return rest.OfferForm{
		NodeID:   schema.NodeID,
		ServerID: schema.ServerID,
		Fields:   fields,
		Servers:  servers,
	}, nil
}

type fpOrderLister struct {
	c *fp.Client
}

func (f fpOrderLister) ListOrders(ctx context.Context) ([]rest.OrderListItem, error) {
	sales, err := f.c.GetSales(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]rest.OrderListItem, 0, len(sales))
	for _, s := range sales {
		items = append(items, rest.OrderListItem{
			ID:        s.ID,
			Status:    string(s.Status),
			BuyerName: s.BuyerName,
			Summary:   s.Summary,
			Price:     s.Price,
		})
	}
	return items, nil
}

type fpOrderGetter struct {
	c *fp.Client
}

func (f fpOrderGetter) GetOrder(ctx context.Context, orderID string) (rest.OrderDetail, error) {
	o, err := f.c.GetOrder(ctx, orderID)
	if err != nil {
		return rest.OrderDetail{}, err
	}
	return rest.OrderDetail{
		ID:        o.ID,
		NodeID:    o.NodeID,
		BuyerID:   o.BuyerID,
		BuyerName: o.BuyerName,
		Amount:    o.Amount,
		Currency:  o.Currency,
		Status:    string(o.Status),
		ChatID:    o.ChatID,
	}, nil
}

type fpChatMessager struct {
	c      *fp.Client
	runner *fp.Runner
	userID string
}

func (f fpChatMessager) SendChatMessage(ctx context.Context, node, text string) (rest.MessageSentResult, error) {
	sent, err := f.runner.SendChatMessage(ctx, node, text)
	if err != nil {
		return rest.MessageSentResult{}, err
	}
	return rest.MessageSentResult{MessageID: strconv.FormatInt(sent.MessageID, 10)}, nil
}

type fpOrderRefunder struct {
	c *fp.Client
}

func (f fpOrderRefunder) RefundOrder(ctx context.Context, orderID string) (rest.RefundedResult, error) {
	res, err := f.c.RefundOrder(ctx, orderID)
	if err != nil {
		return rest.RefundedResult{}, err
	}
	return rest.RefundedResult{OrderID: res.OrderID}, nil
}

// Версия встраивается через -ldflags при релизной сборке (см. .goreleaser.yml).
// При обычном `go build` остаются значения по умолчанию.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func printVersion() {
	slog.Info("build info", "version", version, "commit", commit, "date", date)
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
	showVersion := flag.Bool("version", false, "print build version and exit")
	flag.Parse()

	level := slog.LevelInfo
	if *debug {
		level = slog.LevelDebug
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})))

	printVersion()

	if *showVersion {
		return
	}

	err := godotenv.Load()
	if err != nil {
		slog.Error("Error loading .env file", "err", err)
		os.Exit(1)
	}

	goldenKey := os.Getenv("FP_GOLDEN_KEY")
	sessionID := os.Getenv("FP_PHPSESSID")
	goldenSeal := os.Getenv("FP_GOLDEN_SEAL")
	csrfToken := os.Getenv("FP_CSRF_TOKEN")

	slog.Info("engine starting")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	client := fp.NewClient(goldenKey, sessionID, goldenSeal, csrfToken, 800*time.Millisecond, 600*time.Millisecond)
	account, err := client.GetAccount(ctx)
	if err != nil {
		slog.Error("get account failed", "err", err)
		return
	}
	slog.Info("account loaded", "login", account.Login, "balance", account.Balance)

	userID := os.Getenv("FP_USER_ID")

	objectTypes := []string{"orders_counters", "chat_counter", "chat_bookmarks"}

	runner := fp.NewRunner(client, userID, objectTypes)

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
	srv.SetOfferCreator(fpOfferCreator{c: client})
	srv.SetOfferEditor(fpOfferEditor{c: client})
	srv.SetOfferDeleter(fpOfferDeleter{c: client})
	srv.SetOfferLister(fpOfferLister{c: client})
	srv.SetOfferFormGetter(fpOfferFormGetter{c: client})
	srv.SetOrderLister(fpOrderLister{c: client})
	srv.SetOrderGetter(fpOrderGetter{c: client})
	srv.SetChatMessager(fpChatMessager{c: client, runner: runner, userID: userID})
	srv.SetOrderRefunder(fpOrderRefunder{c: client})
	slog.Info("orders endpoints wired")
	resumeCh := make(chan struct{}, 1)
	srv.SetResumeCh(resumeCh)
	slog.Info("offer CRUD wired")
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
				if !awaitResume(ctx, resumeCh, client, runner) {
					// ctx отменён (SIGINT) — выходим.
					return
				}
				srv.SetState("healthy")
				buf.Push([]engine.Event{{
					Type: engine.EngineStatus,
					At:   time.Now(),
					Payload: engine.EngineStatusPayload{
						State: "healthy",
					},
				}})
				slog.Info("auth restored, polling resumed")
				continue
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

func awaitResume(ctx context.Context, resumeCh <-chan struct{}, client *fp.Client, runner *fp.Runner) bool {
	for {
		slog.Info("polling paused, waiting for POST /control/resume or signal")
		select {
		case <-ctx.Done():
			return false
		case <-resumeCh:
		}

		// Перечитываем .env (оператор обновил seal там).
		envMap, err := godotenv.Read()
		if err != nil {
			slog.Error("resume: re-read .env failed", "err", err)
			continue
		}
		newSeal := envMap["FP_GOLDEN_SEAL"]
		_, _, currentSeal, _ := client.SnapshotAuth()
		if newSeal == "" || newSeal == currentSeal {
			slog.Error("resume: .env FP_GOLDEN_SEAL not updated (still same as in-memory); staying paused")
			continue
		}

		newKey := envMap["FP_GOLDEN_KEY"]
		newSession := envMap["FP_PHPSESSID"]
		newCSRF := envMap["FP_CSRF_TOKEN"]
		client.UpdateAuth(newKey, newSession, newSeal, newCSRF)
		slog.Info("resume: auth updated from .env, re-init runner")

		if err := runner.Init(ctx); err != nil {
			slog.Error("resume: runner.Init failed", "err", err)
			continue
		}
		return true
	}
}
