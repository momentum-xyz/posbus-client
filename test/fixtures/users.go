package fixtures

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/hashicorp/go-retryablehttp"
	influxWrite "github.com/influxdata/influxdb-client-go/v2/api/write"

	"github.com/momentum-xyz/ubercontroller/pkg/cmath"
	"github.com/momentum-xyz/ubercontroller/pkg/posbus"
	"github.com/momentum-xyz/ubercontroller/types/entry"
	"github.com/momentum-xyz/ubercontroller/universe"
	"github.com/momentum-xyz/ubercontroller/universe/logic/api/dto"
	"github.com/momentum-xyz/ubercontroller/universe/logic/common"
	"github.com/momentum-xyz/ubercontroller/utils/umid"
)

// Create a guest account in the backend through the API.
// Returns the guest user ID and authentication token (JWT).
// For use in integration tests.
func GuestAccount(ctURL *url.URL) (*umid.UMID, string, error) {
	url := ctURL.JoinPath("/api/v4/auth/guest-token").String()
	// retryable is current workaround for slow starting controller service
	resp, err := retryablehttp.Post(url, "", nil)
	if err != nil {
		return nil, "", fmt.Errorf("Error getting guest token: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("guest token: %s %s\n%s", url, resp.Status, string(body))
	}

	var u dto.User
	d := json.NewDecoder(resp.Body)
	err = d.Decode(&u)
	if err != nil {
		return nil, "", fmt.Errorf("Error decoding guest user %w", err)
	}

	token := *u.JWTToken

	return &u.ID, token, nil
}

// fixtory blueprint for users
func UserBlueprint(i int, last entry.User) entry.User {
	userId := umid.New()
	return entry.User{
		UserID:     userId,
		UserTypeID: getNormalUserTypeID(),
		Profile:    entry.UserProfile{},
	}
}

func getNormalUserTypeID() umid.UMID {
	typeId, err := common.GetNormalUserTypeID()
	if err != nil {
		panic(err)
	}
	return typeId
}

func GetFakeUser() universe.User {
	return &FakeUser{
		id: umid.New(),
	}
}

// Fake user for test fixtures.
// Minimal implementation of universe.User interface
// TODO: move to 'fakes' package(?)
type FakeUser struct {
	universe.User
	id   umid.UMID
	w    universe.World
	t    cmath.TransformNoScale
	lspt int64
}

func (m *FakeUser) GetID() umid.UMID {
	return m.id
}

func (m *FakeUser) SetWorld(w universe.World) {
	m.w = w
}
func (m *FakeUser) GetWorld() universe.World {
	return m.w
}

func (m *FakeUser) SetObject(o universe.Object) {}
func (m *FakeUser) SetTransform(t cmath.TransformNoScale) {
	m.t = t
}
func (m *FakeUser) GetTransform() *cmath.TransformNoScale {
	return &m.t
}
func (m *FakeUser) GetLastPosTime() int64 {
	return m.GetLastSendPosTime()
}
func (m *FakeUser) GetLastSendPosTime() int64 {
	return m.lspt
}
func (m *FakeUser) SetLastSendPosTime(i int64) {
	m.lspt = i
}
func (m *FakeUser) ReleaseSendBuffer() {}
func (m *FakeUser) LockSendBuffer()    {}

func (m *FakeUser) Send(message *websocket.PreparedMessage) error {
	return nil
}
func (m *FakeUser) SendDirectly(message *websocket.PreparedMessage) error {
	return nil
}

func (m *FakeUser) AddInfluxTags(
	prefix string, point *influxWrite.Point) *influxWrite.Point {
	return nil
}
func (m *FakeUser) GetUserDefinition() *posbus.UserData {
	return &posbus.UserData{
		ID: m.id,
	}
}
