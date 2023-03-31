package fixtures

import (
	"testing"

	"github.com/momentum-xyz/ubercontroller/pkg/cmath"
	"github.com/momentum-xyz/ubercontroller/universe"
	"github.com/momentum-xyz/ubercontroller/utils/umid"
)

// Fixtures for 3D objects.

func Object(t *testing.T, node universe.Node, w universe.World, name string) universe.Object {
	// Objects are... weird.
	// Can't just create a struct and pass that in...
	// So TODO: creating an easy to use factory for these.
	id := umid.New()
	obj, err := w.CreateObject(id)
	if err != nil {
		t.Fatalf("create object: %s", err)
	}
	obj.SetOwnerID(w.GetOwnerID(), false)
	oTypeId := umid.MustParse("4ed3a5bb-53f8-4511-941b-07902982c31c") // Custom object type, preseeded value.
	oType, ok := node.GetObjectTypes().GetObjectType(oTypeId)
	if !ok {
		t.Fatalf("get object type: %s", err)
	}
	obj.SetObjectType(oType, false)
	obj.SetName(name, false)
	obj.SetEnabled(true)

	obj.SetTransform(&cmath.Transform{
		Position: cmath.Vec3{
			X: 53.2194, Y: 6.5665, Z: 18,
		},
	}, false)

	return obj
}
