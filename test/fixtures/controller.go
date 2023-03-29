package fixtures

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"go.uber.org/zap"

	"github.com/momentum-xyz/ubercontroller/config"
	"github.com/momentum-xyz/ubercontroller/pkg/service"
	"github.com/momentum-xyz/ubercontroller/types"
	"github.com/momentum-xyz/ubercontroller/universe"
)

// Run the controller in a goroutine.
func Controller(t *testing.T, pgConfig config.Postgres, mmURL *url.URL) (*url.URL, universe.Node) {
	// TODO: move this to controller repo

	ctx, cancel := context.WithCancel(context.Background())
	host := "127.0.0.42"
	port := uint(4242) // TODO: get a random, open, port
	cfg := &config.Config{
		Postgres: pgConfig,
	}
	cfg.Common.RenderInternalURL = mmURL.String()
	cfg.Settings.Address = host
	cfg.Settings.Port = port
	cfg.Settings.SeedDataFiles = "/dev/null"

	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("zap logger: %s", err)
	}
	// TODO: refacor context bits to use proper setter&getter
	ctx = context.WithValue(ctx, types.LoggerContextKey, logger.Sugar())
	ctx = context.WithValue(ctx, types.ConfigContextKey, cfg)
	pool, err := service.CreateDBConnection(ctx, &cfg.Postgres)
	if err != nil {
		t.Fatalf("database connection: %s", err)
	}
	node, err := service.LoadNode(ctx, cfg, pool)
	if err != nil {
		t.Fatalf("load node: %s", err)
	}
	go func() {
		if err := node.Run(); err != nil {
			fmt.Printf("run node: %s\n", err)
		}
	}()
	//TODO: implement a wait-for-ready for above goroutine

	t.Cleanup(func() {
		if err := node.Stop(); err != nil {
			t.Error(err)
		}
		cancel()
	})

	return &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", host, port),
	}, node
}
