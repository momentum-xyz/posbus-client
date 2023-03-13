//go:build js && wasm

package main

import (
	"context"
	"encoding/json"
	"reflect"
	"syscall/js"

	"github.com/google/uuid"
	"github.com/momentum-xyz/posbus-client/pbc"
	"github.com/momentum-xyz/ubercontroller/logger"
	"github.com/momentum-xyz/ubercontroller/pkg/cmath"
	"github.com/momentum-xyz/ubercontroller/pkg/posbus"
	"github.com/momentum-xyz/ubercontroller/utils"
	"github.com/pkg/errors"
)

var client *pbc.Client
var done chan bool
var msgPort js.Value
var jsPromise js.Value // javascript Promise constructor

func main() {
	ctx := context.Background()
	client = pbc.NewClient(ctx)
	client.SetCallback(onMessage)
	done = make(chan bool)
	jsPromise = js.Global().Get("Promise")
	// Export a global variable to javascript
	js.Global().Set("PBC", make(map[string]interface{}))
	namespace := js.Global().Get("PBC")
	namespace.Set("send", js.FuncOf(Send))
	namespace.Set("setURL", js.FuncOf(SetURL))
	namespace.Set("setToken", js.FuncOf(SetToken))
	namespace.Set("setPort", js.FuncOf(SetPort))
	namespace.Set("connect", js.FuncOf(Connect))
	namespace.Set("disconnect", js.FuncOf(Disconnect))
	namespace.Set("teleport", js.FuncOf(Teleport))
	namespace.Set("close", js.FuncOf(Close))
	<-done
}

func SetURL(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		logger.L().Debugf("%+v\n", "PB SetURL: too few arguments")
		return nil
	}
	client.SetURL(args[0].String())
	return nil
}

func SetToken(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		logger.L().Debugf("%+v\n", "PB SetURL: too few arguments")
		return nil
	}
	client.SetToken(args[0].String())
	return nil
}

func Connect(this js.Value, args []js.Value) any {
	if len(args) < 3 {
		logger.L().Debugf("%+v\n", "PB Connect: too few arguments")
		return promiseReject(errors.New("to few arguments"))
	}
	url := args[0].String()
	token := args[1].String()
	userId, err := uuid.Parse(args[2].String())
	if err != nil {
		logger.L().Debugf("invalid userId %s", args[2].String())
		return promiseReject(errors.Wrap(err, "userId"))
	}
	handler := promiseExecutor(
		func() error {
			return client.Connect(url, token, userId)
		},
	)
	return jsPromise.New(handler)
}

func Teleport(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		logger.L().Debugf("%+v\n", "PB Teleport: too few arguments")
		return nil
	}
	world, err := uuid.Parse(args[0].String())
	if err != nil {
		logger.L().Error("invalid world ID %s", err)
		return nil
	}
	go client.Send(
		posbus.NewMessageFromData(
			posbus.TypeTeleportRequest,
			posbus.TeleportRequest{Target: world},
		).Buf(),
	)
	return nil
}

func Disconnect(this js.Value, args []js.Value) interface{} {
	client.Close()
	logger.L().Debug("Disconnected")
	return nil
}

func Close(this js.Value, args []js.Value) interface{} {
	client.Close()
	done <- true
	return nil
}

func Send(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		logger.L().Debugf("%+v\n", "PB Send: too few arguments")
		return nil
	}
	msgId := posbus.MessageIdByName(args[0].String())
	if msgId == 0 {
		logger.L().Debugf("%+v\n", "PB Send: unknown message name")
		return nil
	}
	dataString := []byte(args[1].String())
	//fmt.Println(dataString)
	var msg *posbus.Message
	switch msgId {
	case posbus.TypeSetUsersTransforms:
		return nil
	case posbus.TypeSendTransform:
		p := cmath.NewUserTransform()
		json.Unmarshal(dataString, &p)
		msg = posbus.NewMessageFromBuffer(
			posbus.TypeSendTransform,
			p.Bytes(),
		)
	case posbus.TypeGenericMessage:
		msg = posbus.NewMessageFromBuffer(
			posbus.TypeSendTransform,
			[]byte(dataString),
		)
	default:
		v := reflect.New(posbus.MessageDataTypeById(msgId)).Interface()
		err := json.Unmarshal(dataString, v)
		if err != nil {
			logger.L().Debugf("PB Send: cant unmarshal JSON : %+v\n", string(dataString))
			return nil
		}
		msg = posbus.NewMessageFromData(msgId, v)
	}
	client.Send(msg.Buf())

	return nil
}

// TODO: replace with below method using mapdecode
func msgGoToJs_json(data any) (map[string]any, error) {
	r, err := json.Marshal(data)
	if err != nil {
		return nil, errors.Wrap(err, "marshal")
	}
	r2 := make(map[string]any)
	if err := json.Unmarshal(r, &r2); err != nil {
		return nil, errors.Wrap(err, "unmarshal")
	}
	return r2, nil
}

// TODO: add converters to MapDecode for this
func msgGoToJs(data any) (map[string]any, error) {
	r := make(map[string]any)
	if err := utils.MapDecode(data, &r); err != nil {
		return nil, errors.Wrap(err, "decode")
	}
	return r, nil
}

func td(x any) {
	logger.L().Debugf("td %T", x)
	switch x := x.(type) {
	case uint32:
		logger.L().Debug("uint32", x)
	case posbus.Signal:
		logger.L().Debug("posbus.Signal", x)
	case map[string]any:
		logger.L().Debug("map[string]any", x)
	default:
		logger.L().Debug("unknown", x)

	}
}

func onMessage(msgType posbus.MsgType, data interface{}) error {
	// workaround: process in goroutine to avoid locking event thread
	go func() {
		r, err := msgGoToJs_json(data)
		if err != nil {
			logger.L().Error("to map", err)
			return
		}
		typeName := posbus.MessageNameById(msgType)
		logger.L().Debugf("Incoming message: %+v %+v\n", typeName, r)
		if msgPort.IsUndefined() {
			logger.L().Error("No port to post message to")
			return
		}
		msgPort.Call("postMessage", []any{typeName, r})
	}()

	return nil
}

func SetPort(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		logger.L().Debugf("%+v\n", "PB SetPort: too few arguments")
		return nil
	}
	msgPort = args[0]
	logger.L().Debug("communication port set")
	return nil
}

// Helper to run a goroutine as a javascript Promise executor.
func promiseExecutor(f func() error) js.Func {
	var jsHandler js.Func
	jsHandler = js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) != 2 {
			logger.L().Error("Promise executor: only 2 args allowed")
			return nil
		}
		resolve := args[0]
		reject := args[1]
		go func() {
			defer jsHandler.Release()
			if err := f(); err != nil {
				reject.Invoke(err) // hmm, are errors transferable?
				return
			}
			resolve.Invoke()
		}()
		return nil
	})
	return jsHandler
}

// Helper to return rejected javascript Promise.
func promiseReject(err error) js.Value {
	return jsPromise.Call("reject", err.Error())
}
