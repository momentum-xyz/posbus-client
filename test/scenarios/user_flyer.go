package scenarios

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/url"
	"time"

	"github.com/momentum-xyz/posbus-client/pbc"
	"github.com/momentum-xyz/posbus-client/test/fixtures"
	"github.com/momentum-xyz/ubercontroller/pkg/cmath"
	"github.com/momentum-xyz/ubercontroller/pkg/posbus"
	"github.com/momentum-xyz/ubercontroller/utils/umid"
)

// Flying speed, in 'units' per second. Based on 3d engine values
const SPEED_CRUISE = float64(8)
const SPEED_BOOST = float64(16)

// Interval clients send their position.
const POS_UPDATE_TIME = 250 * time.Millisecond

// Space constrains, randomly move inside this cube
const MIN = -100
const MAX = 100

// Test scenario of a guest user flying around in a world.
func GuestFlyer(ctx context.Context, i uint64, backend *url.URL, world *umid.UMID) error {
	userID, token, err := fixtures.GuestAccount(backend)
	if err != nil {
		return fmt.Errorf("User for guest flyer scenario: %w", err)
	}

	client := pbc.NewClient()
	us := &scenario{
		index:  i,
		client: client,
		rnd:    rand.New(rand.NewSource(int64(userID.ClockSequence()))),
	}
	client.SetCallback(us.onMessage)
	url := backend.JoinPath("/posbus").String()
	client.Connect(ctx, url, token, *userID)
	client.Send(
		posbus.BinMessage(
			&posbus.TeleportRequest{Target: *world},
		),
	)

	log.Printf("Guest flyer %d running", i)
	ticker := time.NewTicker(POS_UPDATE_TIME)
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return nil
		case <-ticker.C:
			us.moveUser(POS_UPDATE_TIME)
		}
	}
}

type scenario struct {
	index    uint64
	client   *pbc.Client
	position cmath.Vec3
	rotation cmath.Vec3
	moving   bool
	target   cmath.Vec3
	rnd      *rand.Rand
}

func (s *scenario) onMessage(msg posbus.Message) {
	switch m := msg.(type) {
	case *posbus.MyTransform:
		s.position = m.Position
		s.rotation = m.Rotation
		s.moving = true
	default:
		return
	}
}

func (s *scenario) moveUser(step time.Duration) {
	distance := cmath.Distance(&s.position, &s.target)
	amount := float32(step.Seconds() * SPEED_CRUISE)
	if distance < float64(amount) { // close enough!
		s.setRandomTarget()
		return
	}
	direction := vec3Normalize(vec3Sub(s.target, s.position))
	move := cmath.MultiplyN(direction, amount)
	s.position.Plus(move)
	s.rotation = direction // TODO: calc euler angles here

	nPos := &posbus.MyTransform{
		Position: s.position,
		Rotation: s.rotation,
	}
	//fmt.Printf("Move %d: %+v\n", s.index, nPos)
	s.client.Send(
		posbus.BinMessage(nPos),
	)
}

// Set a random target location for the use to fly to.
func (s *scenario) setRandomTarget() cmath.Vec3 {
	v := cmath.Vec3{
		X: randomF(s.rnd, MIN, MAX),
		Y: randomF(s.rnd, MIN, MAX),
		Z: randomF(s.rnd, MIN, MAX),
	}
	s.target = v
	return v
}

func randomF(rnd *rand.Rand, min float32, max float32) float32 {
	return min + rnd.Float32()*(max-min)
}

// TODO: really need to extend our cmath.Vec3.. or use an existing lib?

// Normalize a vec3 into 'unit' vec3
func vec3Normalize(v cmath.Vec3) cmath.Vec3 {
	n := dotProd(v, v)
	if n == 0 {
		return cmath.Vec3{X: 0, Y: 0, Z: 0}
	}
	d := float32(1 / math.Sqrt(float64(n)))
	return cmath.MultiplyN(v, d)
}

func dotProd(v cmath.Vec3, v2 cmath.Vec3) float32 {
	return v.X*v2.X + v.Y*v2.Y + v.Z*v.Z
}

func vec3Sub(v cmath.Vec3, v2 cmath.Vec3) cmath.Vec3 {
	return cmath.Vec3{X: v.X - v2.X, Y: v.Y - v2.Y, Z: v.Z - v2.Z}
}
