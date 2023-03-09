package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/momentum-xyz/posbus-client/pbc"
	"github.com/momentum-xyz/ubercontroller/logger"
	"github.com/momentum-xyz/ubercontroller/pkg/cmath"
	"github.com/momentum-xyz/ubercontroller/pkg/posbus"
	"reflect"
	"syscall/js"
)

var client *pbc.Client
var done chan bool

func main() {
	ctx := context.Background()
	client = pbc.NewClient(ctx)
	client.SetCallback(onMessage)
	done = make(chan bool)
	js.Global().Set("PB_Send", js.FuncOf(Send))
	js.Global().Set("PB_SetURL", js.FuncOf(SetURL))
	js.Global().Set("PB_SetToken", js.FuncOf(SetToken))
	js.Global().Set("PB_Connect", js.FuncOf(Connect))
	js.Global().Set("Close", js.FuncOf(Close))
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

func Connect(this js.Value, args []js.Value) interface{} {
	if len(args) < 3 {
		logger.L().Debugf("%+v\n", "PB Connect: too few arguments")
		return nil
	}
	client.Connect(args[0].String(), args[1].String(), uuid.MustParse(args[2].String()))
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

func onMessage(msgType posbus.MsgType, data interface{}) error {
	r, _ := json.Marshal(data)
	fmt.Printf("Incoming message: %+v %+v\n", posbus.MessageNameById(msgType), string(r))
	return nil
}
