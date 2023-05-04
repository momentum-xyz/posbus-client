package fixtures

import (
	"encoding/json"
	"testing"
	"time"

	_ "embed"

	"github.com/momentum-xyz/ubercontroller/types/entry"
	"github.com/momentum-xyz/ubercontroller/universe"
	"github.com/momentum-xyz/ubercontroller/universe/logic/tree"
)

//go:embed data/world_template.json
var worldTemplate []byte

// Create a minimal world.
func World(t *testing.T, node universe.Node, creator *entry.User, name string) universe.World {
	var template tree.WorldTemplate
	if err := json.Unmarshal(worldTemplate, &template); err != nil {
		t.Fatalf("read world template: %s", err)
	}

	template.ObjectName = &name
	template.OwnerID = &creator.UserID

	worldID, err := tree.AddWorldFromTemplate(&template, true) // can't actually create a world without updateDB... :/
	if err != nil {
		t.Fatalf("create world: %s", err)
	}
	world, ok := node.GetWorlds().GetWorld(worldID)
	if !ok {
		t.Fatalf("create world: %s", err)
	}
	time.Sleep(500 * time.Millisecond) // hack! give controller goroutines a chance to run
	return world
}
