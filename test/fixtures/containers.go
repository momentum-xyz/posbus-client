package fixtures

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/momentum-xyz/ubercontroller/config"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Create a (docker) network to put multiple containers into.
func ContainerNetwork(t *testing.T, name string) {
	t.Helper()
	ctx := context.Background()
	n, err := testcontainers.GenericNetwork(ctx, testcontainers.GenericNetworkRequest{
		NetworkRequest: testcontainers.NetworkRequest{
			Name: name,
		},
	})
	if err != nil {
		t.Fatalf("container network: %s", err)
	}
	t.Cleanup(func() {
		if err := n.Remove(ctx); err != nil {
			t.Fatalf("network cleanup: %s", err)
		}
	})
}

// Startup a postgresql database.
func ContainerPostgres(t *testing.T, networkName string) (testcontainers.Container, config.Postgres) {
	// or pass a config.Postges into this function?
	const dbHost = "db"
	const dbName = "test_momentum"
	const dbUser = "test_momentum"
	const dbPass = "test_momentum"
	ctx := context.Background()
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "postgres:15.2",
			Env: map[string]string{
				"POSTGRES_DB":       dbName,
				"POSTGRES_USER":     dbUser,
				"POSTGRES_PASSWORD": dbPass,
			},
			Networks: []string{networkName},
			NetworkAliases: map[string][]string{
				networkName: {dbHost},
			},
			ExposedPorts: []string{"5432/tcp"},
			WaitingFor:   wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(5 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("postgres container: %s", err)
	}
	t.Cleanup(func() {
		err := c.Terminate(ctx)
		if err != nil {
			t.Fatalf("terminate container: %s", err)
		}
	})
	t.Log("Postgres container started.")
	// TODO: make optional. Allow also using the internal network host+port
	host, errH := c.Host(ctx)
	port, errP := c.MappedPort(ctx, "5432/tcp")
	if err := errors.Join(errH, errP); err != nil {
		t.Fatalf("postgres: %s", err)
	}
	pgConfig := config.Postgres{
		HOST:     host,
		PORT:     uint(port.Int()),
		DATABASE: dbName,
		USERNAME: dbUser,
		PASSWORD: dbPass,
		MAXCONNS: 4,
	}
	return c, pgConfig
}

// Run the Media manager as a container.
func ContainerMediaManager(t *testing.T, networkName string) (testcontainers.Container, *url.URL) {
	ctx := context.Background()

	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:    "ghcr.io/momentum-xyz/media-manager:develop",
			Env:      map[string]string{},
			Networks: []string{networkName},
			NetworkAliases: map[string][]string{
				networkName: {"media-manager"},
			},
			ExposedPorts: []string{"4000/tcp"},
			WaitingFor:   wait.ForHTTP("/health").WithPort("4000/tcp").WithStartupTimeout(4 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("controller container: %s", err)
	}
	t.Cleanup(func() {
		t.Log("Cleanup!")
		err := c.Terminate(ctx)
		if err != nil {
			t.Fatalf("terminate nedia manager: %s", err)
		}
	})
	host, errH := c.Host(ctx)
	port, errP := c.MappedPort(ctx, "4000/tcp")
	if errors.Join(errH, errP) != nil {
		t.Fatalf("media manager url: %s", err)
	}
	mmURL := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", host, port.Int()),
	}
	//c.FollowOutput(logConsumer)
	//if err := c.StartLogProducer(ctx); err != nil {
	//	s.T().Fatalf("media manager log consumer: %s", err)
	//}
	return c, mmURL
}

// Run the controller as a container.
func ContainerController(t *testing.T, networkName string, config config.Postgres) testcontainers.Container {
	ctx := context.Background()

	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "ghcr.io/momentum-xyz/ubercontroller:develop",
			Env: map[string]string{
				"PGDB_HOST":           config.HOST,
				"DB_PORT":             strconv.FormatUint(uint64(config.PORT), 10),
				"DB_DATABASE":         config.DATABASE,
				"DB_USERNAME":         config.USERNAME,
				"DB_PASSWORD":         config.PASSWORD,
				"DB_MAX_CONNS":        "4",
				"CONTROLLER_LOGLEVEL": "0",
				"RENDER_INTERNAL_URL": "http://media-manager:4000",
			},
			Networks: []string{networkName},
			NetworkAliases: map[string][]string{
				networkName: {"controller"},
			},
			ExposedPorts: []string{"4000/tcp"},
			WaitingFor:   wait.ForHTTP("/health").WithPort("4000/tcp").WithStartupTimeout(7 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("controller container: %s", err)
	}
	t.Cleanup(func() {
		err := c.Terminate(ctx)
		if err != nil {
			t.Fatalf("terminate controller: %s", err)
		}
	})
	// c.FollowOutput(logConsumer)
	// if err := c.StartLogProducer(ctx); err != nil {
	// 	t.Fatalf("controller container log consumer: %s", err)
	//}
	return c
}
