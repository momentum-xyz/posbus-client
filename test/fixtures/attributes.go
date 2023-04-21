package fixtures

import (
	"testing"

	"github.com/momentum-xyz/ubercontroller/types/entry"
	"github.com/momentum-xyz/ubercontroller/universe"
	"github.com/momentum-xyz/ubercontroller/utils/umid"
)

// Change an attribute to trigger the backend to send a update message.
func ChangePosbusAutoAttribute(t *testing.T, node universe.Node, world universe.World, userId *umid.UMID) {
	pluginID := universe.GetSystemPluginID()
	attributeID := entry.NewAttributeID(pluginID, "VoiceChatAction")
	objectUserAttributeID := entry.NewObjectUserAttributeID(attributeID, world.GetID(), *userId)

	_, err := node.GetObjectUserAttributes().Upsert(objectUserAttributeID,
		func(current *entry.AttributePayload) (*entry.AttributePayload, error) {
			value := map[string]any{"foo": map[string]any{"bar": "baz"}}

			v := entry.AttributeValue(value)
			return entry.NewAttributePayload(&v, nil), nil
		}, true)
	if err != nil {
		t.Fatalf("update object user attribute %v", err)
	}
}
