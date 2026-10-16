// Command singleton runs the Singleton network node: registry, verifier, console, wallet
// and landing page in one binary. `singleton verify-root` audits a published set.
package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/stbodurov/bingo/singleton/internal/app"
	"github.com/stbodurov/bingo/singleton/internal/proof"
	"github.com/stbodurov/bingo/singleton/web"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "verify-root" {
		verifyRoot(os.Args[2:])
		return
	}
	addr := flag.String("addr", envOr("ADDR", ":4201"), "listen address")
	dbPath := flag.String("db", envOr("DB_PATH", "data/singleton.db"), "SQLite database path")
	baseURL := flag.String("base-url", envOr("BASE_URL", ""), "public base URL (default http://localhost<addr>)")
	noSeed := flag.Bool("no-seed", false, "do not create the sandbox network and demo workspace")
	holders := flag.Int("holders", 1800, "synthetic holders in the sandbox network")
	epoch := flag.Duration("epoch", 90*time.Second, "minimum time between sealed epochs")
	flag.Parse()

	logger := log.New(os.Stdout, "singleton ", log.LstdFlags|log.Lmsgprefix)
	if *dbPath != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(*dbPath), 0o755); err != nil {
			logger.Fatal(err)
		}
	}
	db, err := app.OpenDB(*dbPath)
	if err != nil {
		logger.Fatal(err)
	}
	cfg := app.DefaultConfig()
	cfg.EpochInterval = *epoch
	if *baseURL != "" {
		cfg.BaseURL = *baseURL
	} else {
		host := *addr
		if host[0] == ':' {
			host = "localhost" + host
		}
		cfg.BaseURL = "http://" + host
	}
	if t := os.Getenv("ADMIN_TOKEN"); t != "" {
		cfg.AdminToken = t
	}
	svc, err := app.NewService(db, cfg, logger)
	if err != nil {
		logger.Fatal(err)
	}
	if !*noSeed && !svc.Seeded() {
		logger.Printf("seeding sandbox network with %d synthetic holders (first run only)…", *holders)
		start := time.Now()
		if err := svc.Seed(*holders); err != nil {
			logger.Fatal(err)
		}
		logger.Printf("seeded in %s", time.Since(start).Round(time.Millisecond))
	}
	if _, err := svc.CurrentEpoch(); err != nil {
		logger.Fatal(err)
	}
	srv, err := web.New(svc, logger)
	if err != nil {
		logger.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		t := time.NewTicker(15 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				svc.MaybeSeal(time.Now().UTC(), false)
				app.SweepPow()
			}
		}
	}()

	hs := &http.Server{Addr: *addr, Handler: srv.Handler(), ReadHeaderTimeout: 10 * time.Second, WriteTimeout: 60 * time.Second}
	go func() {
		logger.Printf("listening on %s", cfg.BaseURL)
		logger.Printf("  landing      %s/", cfg.BaseURL)
		logger.Printf("  pass         %s/pass", cfg.BaseURL)
		logger.Printf("  console      %s/console   (demo: %s / %s)", cfg.BaseURL, app.DemoEmail, app.DemoPassword)
		logger.Printf("  demos        %s/demo/driftwood", cfg.BaseURL)
		if err := hs.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal(err)
		}
	}()
	<-ctx.Done()
	shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = hs.Shutdown(shutdown)
	_ = db.Close()
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// verifyRoot downloads a set snapshot and recomputes its Merkle root, so anyone can check
// that the set verifiers use is the one anchored on-chain.
func verifyRoot(args []string) {
	fs := flag.NewFlagSet("verify-root", flag.ExitOnError)
	base := fs.String("url", "http://localhost:4201", "Singleton node URL")
	epoch := fs.Int64("epoch", 0, "epoch (default: current)")
	set := fs.String("set", "human", "set name")
	_ = fs.Parse(args)
	if *epoch == 0 {
		var e struct {
			Epoch struct {
				ID int64 `json:"epoch"`
			} `json:"epoch"`
		}
		must(getJSON(*base+"/api/epoch", &e))
		*epoch = e.Epoch.ID
	}
	var snap struct {
		Root     string   `json:"root"`
		AnchorTx string   `json:"anchorTx"`
		Members  []string `json:"members"`
	}
	must(getJSON(fmt.Sprintf("%s/api/sets/%d/%s", *base, *epoch, *set), &snap))
	raw := make([][]byte, len(snap.Members))
	for i, m := range snap.Members {
		b, err := hex.DecodeString(m)
		must(err)
		raw[i] = b
	}
	got := proof.BuildMerkle(raw).RootHex()
	fmt.Printf("epoch %d · set %s · %d members\n", *epoch, *set, len(raw))
	fmt.Printf("published root  %s\nrecomputed root %s\nanchor tx       %s\n", snap.Root, got, snap.AnchorTx)
	if got != snap.Root {
		fmt.Println("MISMATCH — the served set does not match the committed root")
		os.Exit(1)
	}
	fmt.Println("OK — the set matches its commitment")
}

func getJSON(url string, v any) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("%s: HTTP %d", url, resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(v)
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
