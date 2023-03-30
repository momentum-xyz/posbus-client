// Integration test for pbc.Client
package test

import (
	"context"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/k-yomo/fixtory/v2"
	"github.com/momentum-xyz/posbus-client/pbc"
	"github.com/momentum-xyz/posbus-client/test/fixtures"
	"github.com/momentum-xyz/ubercontroller/config"
	"github.com/momentum-xyz/ubercontroller/pkg/cmath"
	"github.com/momentum-xyz/ubercontroller/pkg/posbus"
	"github.com/momentum-xyz/ubercontroller/types/entry"
	"github.com/momentum-xyz/ubercontroller/universe"
	"github.com/momentum-xyz/ubercontroller/utils/umid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
)

type ClientTestSuite struct {
	suite.Suite
	logConsumer *TestLogConsumer
	mmContainer testcontainers.Container
	mmURL       *url.URL
	pgContainer testcontainers.Container
	pgConfig    config.Postgres
	ctURL       *url.URL
	node        universe.Node
	guestId     umid.UMID
	guestToken  string
	world       universe.World
	Client      *pbc.Client
}

func TestClientTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, &ClientTestSuite{})
}

func (s *ClientTestSuite) SetupSuite() {
	s.logConsumer = &TestLogConsumer{}
	networkName := "test_momentum"
	fixtures.ContainerNetwork(s.T(), networkName)
	s.mmContainer, s.mmURL = fixtures.ContainerMediaManager(s.T(), networkName)
	s.pgContainer, s.pgConfig = fixtures.ContainerPostgres(s.T(), networkName)

	s.ctURL, s.node = fixtures.Controller(s.T(), s.pgConfig, s.mmURL)

	ctx := context.Background()
	userFactory := fixtory.NewFactory(s.T(), entry.User{})
	userFactory.OnBuild = func(t *testing.T, u *entry.User) {
		if err := s.node.CreateUsers(ctx, u); err != nil {
			panic(err)
		}
	}

	creatorName := "Crea Tor"
	creator := userFactory.NewBuilder(fixtures.UserBlueprint, entry.User{Profile: entry.UserProfile{Name: &creatorName}}).Build()

	s.world = fixtures.World(s.T(), s.node, creator, "Gaia")
}

func (s *ClientTestSuite) SetupTest() {
	s.guestId, s.guestToken = fixtures.GuestAccount(s.T(), s.ctURL)
	require.NotEmpty(s.T(), s.guestId)
	require.NotEmpty(s.T(), s.guestToken)

}

func (s *ClientTestSuite) TestClient() {
	//assert := assert.New(s.T())
	require := require.New(s.T())

	ctx := context.Background()
	client := pbc.NewClient()
	require.NotEmpty(s.T(), s.guestId)
	require.NotEmpty(s.T(), s.guestToken)

	// channel to read back messages for testing
	ch := make(chan any, 1)
	s.T().Cleanup(func() {
		client.SetCallback(func(msg posbus.Message) error {
			return nil
		})
		client.Close()
		close(ch)
	})
	client.SetCallback(func(msg posbus.Message) error {
		ch <- msg
		return nil
	})

	url := s.ctURL.JoinPath("posbus").String()

	err := client.Connect(ctx, url, s.guestToken, s.guestId)
	if err != nil {
		s.T().Fatalf("connection: %s", err)
	}

	// Assertion on the first message
	assertNextMsg(s.T(), ch, &posbus.Signal{}, func(sig *posbus.Signal) {
		require.Equal(posbus.SignalConnected, sig.Value, "First message is connected signal")
	})

	// Teleport to a non existant world.
	var nonExistantWorld = umid.MustParse("3b52cc0c-0e58-48ed-b147-f6d2d14c137b")
	client.Send(
		posbus.BinMessage(
			&posbus.TeleportRequest{Target: nonExistantWorld},
		),
	)

	assertNextMsg(s.T(), ch, &posbus.Signal{}, func(sig *posbus.Signal) {
		require.Equal(posbus.SignalWorldDoesNotExist, sig.Value, "Signal world does not exist")
	})

	// Teleport to a world.
	client.Send(
		posbus.BinMessage(
			&posbus.TeleportRequest{Target: s.world.GetID()},
		),
	)

	// Teleport should respond to set the world.
	assertNextMsg(s.T(), ch, &posbus.SetWorld{}, func(w *posbus.SetWorld) {
		require.Equal(w.ID, s.world.GetID(), "World ID is the one we teleported to")
	})
	// Then position the current user in the world.
	assertNextMsg(s.T(), ch, &posbus.MyTransform{}, func(w *posbus.MyTransform) {
		expectedPosition := cmath.Vec3{X: 50, Y: 50, Z: 150}
		require.Equal(expectedPosition, w.Position, "Initial user position")
	})

	// Send the object definition of the world.
	assertNextMsg(s.T(), ch, &posbus.AddObjects{}, func(w *posbus.AddObjects) {
		require.Len(w.Objects, 1, "Get one object")
		obj := w.Objects[0]
		require.Equal(s.world.GetID(), obj.ID, "It is the world itself")
		require.Equal(s.node.GetID(), obj.ParentID, "node as its parent")
	})

	// Add users to this world.
	assertNextMsg(s.T(), ch, &posbus.AddUsers{}, func(w *posbus.AddUsers) {
		require.Len(w.Users, 1, "Get one users")
		obj := w.Users[0]
		require.Equal(s.guestId, obj.ID, "It is I! Le guest")
	})

	// And now things become non-deterministic.
	// Currently we get spammed with tranforms... of ourself
	// which is kindof a bug I would say...
	assertNextMsg(s.T(), ch, &posbus.UsersTransformList{}, func(w *posbus.UsersTransformList) {
		require.Len(w.Value, 1, "Get one tranform")
		obj := w.Value[0]
		require.Equal(s.guestId, obj.ID, "It is I! Le guest")
	})
	//assert.Equal(s.T(), "foo", "bar")
}

func assertNextMsg[T any](t *testing.T, ch <-chan any, expectedType T, f func(msg T)) {
	t.Helper()
	const msg1Timeout = 2 * time.Second
	select {
	case msg := <-ch:
		if assert.IsType(t, expectedType, msg, "Message is not expected type") {
			x, _ := msg.(T)
			f(x)
		} else {
			t.FailNow() // Means assert.IsType failed.
		}
	case <-time.After(msg1Timeout):
		t.Fatalf("not received a message within %s seconds", msg1Timeout)
	}
}

// Consumer of container log outputs.
type TestLogConsumer struct {
	Msgs []string
}

func (tc *TestLogConsumer) Accept(l testcontainers.Log) {
	tc.Msgs = append(tc.Msgs, string(l.Content))
}

// Return the last logged lines.
func (s *TestLogConsumer) LastLines(lines uint) string {
	return strings.Join(s.Msgs[:lines], "\n")
}
