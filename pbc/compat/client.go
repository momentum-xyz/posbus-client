// Package compat provide function to workaround autogenerated binding libraries.
//
//   - umid.UMID type, which is aliased uuid.UUD, which in turn is alias for [16]byte.
//     Replace with direct/primative byte slice as input.
//
//   - bytes encode/decode functions, workaround for serveral issues:
//     posbus.Message is interface, after which reflection needs to be used to determine actual message type.
//     And there is a nested map to any/interface{}, which needs to be handled.
//     If external application wants to use their own eventloop (and websocket), instead of the callback mechanism.
package compat

import (
	"encoding/json"
	"fmt"

	"github.com/momentum-xyz/ubercontroller/pkg/cmath"
	"github.com/momentum-xyz/ubercontroller/pkg/posbus"
	"github.com/momentum-xyz/ubercontroller/utils/umid"
)

func Encode(msg posbus.Message) ([]byte, error) {
	b := posbus.BinMessage(msg)
	return b, nil
}

/* would in theory be nicer, but we need map encoder to deal with special types (e.g. uuids, vec3 etc etc)
func EncodeMap(msgType posbus.MsgType, msg map[string]any) ([]byte, error) {
	m, ok := reflect.New(posbus.MessageDataTypeById(msgType)).Interface().(posbus.Message)
	if !ok {
		return nil, fmt.Errorf("failed to create message type %d.", msgType)

	}
	utils.MapEncode(msg, m)
	b := posbus.BinMessage(m)
	return b, nil
}
*/

type DecodeResult struct {
	Type posbus.MsgType
	Data []byte
}

func Decode(b []byte) (*DecodeResult, error) {
	msg, err := posbus.Decode(b)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %w", err)
	}
	// TODO: remove JSON workaround
	bs, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to JSON encode: %w", err)
	}
	// gopy limit: can only return max 2 values :/
	// so need to wrap it
	return &DecodeResult{msg.GetType(), bs}, nil
}

func NewHandshake(
	token string,
	userID []byte, // uuid
	sessionID []byte, // uuid
) posbus.HandShake {
	return posbus.HandShake{
		HandshakeVersion: 1,
		ProtocolVersion:  1,
		Token:            token,
		UserId:           umid.UMID(userID),
		SessionId:        umid.UMID(sessionID),
	}
}

func NewTeleportRequest(uuid []byte) posbus.TeleportRequest {
	id := umid.UMID(uuid)
	return posbus.TeleportRequest{
		Target: id,
	}
}

func NewHighFive(from []byte, to []byte, msg string) posbus.HighFive {
	return posbus.HighFive{
		SenderID:   umid.UMID(from),
		ReceiverID: umid.UMID(to),
		Message:    msg,
	}
}

func NewObjectTransform(
	id []byte,
	position cmath.Vec3,
	rotation cmath.Vec3,
	scale cmath.Vec3,
) posbus.ObjectTransform {
	return posbus.ObjectTransform{
		ID: umid.UMID(id),
		Transform: cmath.Transform{
			Position: position,
			Rotation: rotation,
			Scale:    scale,
		},
	}
}

func NewLock(uuid []byte) posbus.LockObject {
	id := umid.UMID(uuid)
	return posbus.LockObject{
		ID: id,
	}
}

func NewUnlock(uuid []byte) posbus.UnlockObject {
	id := umid.UMID(uuid)
	return posbus.UnlockObject{
		ID: id,
	}
}
