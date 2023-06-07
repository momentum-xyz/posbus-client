package fixtures

import (
	"context"
	"fmt"
	"net"
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
	host := "localhost"
	port := getAvailablePort()
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
	nodeCtx, err := types.NewNodeContext(ctx, logger.Sugar(), cfg)
	if err != nil {
		t.Fatalf("failed to create context: %s", err)
	}
	pool, err := service.CreateDBConnection(nodeCtx, &cfg.Postgres)
	if err != nil {
		t.Fatalf("database connection: %s", err)
	}
	node, err := service.LoadNode(nodeCtx, cfg, pool)
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

func getAvailablePort() uint {
	// TODO: allow controller to run with ':0' and make it return the port number.
	//For now, this racy hack:
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer l.Close()
	return uint(l.Addr().(*net.TCPAddr).Port)
}
