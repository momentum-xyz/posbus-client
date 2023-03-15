package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/momentum-xyz/posbus-client/pbc"
	"github.com/momentum-xyz/ubercontroller/pkg/cmath"
	"github.com/momentum-xyz/ubercontroller/pkg/posbus"
	"github.com/momentum-xyz/ubercontroller/universe/logic/api/dto"
	"github.com/momentum-xyz/ubercontroller/utils"
)

var URL = "http://localhost:4000/posbus"

func main() {
	backendArg := flag.String("backend", "http://localhost:4000", "The URL to the controller backend")
	worldArg := flag.String("world", "975cb9ca-4dfa-4d35-adc2-198ed1f12555", "UUID of a world")
	flag.Parse()
	backend, err := url.Parse(*backendArg)
	if err != nil {
		log.Fatalf("Invalid backend URL %s", err)
	}
	world, err := uuid.Parse(*worldArg)
	if err != nil {
		log.Fatalf("Invalid world %s", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	resp, err := http.Post(backend.JoinPath("/api/v4/auth/guest-token").String(), "", nil)
	if err != nil {
		log.Fatalf("Error of getting guest token %+v\n", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	var u dto.User
	err = json.Unmarshal(body, &u)
	if err != nil {
		log.Fatalf("Error %+v\n", err)
	}

	//fmt.Printf("%+v\n", u)

	client := pbc.NewClient(ctx)
	client.SetCallback(onMessage)
	client.Connect(URL, *u.JWTToken, uuid.MustParse(u.ID))

	client.Send(
		posbus.NewMessageFromData(
			posbus.TypeTeleportRequest,
			posbus.TeleportRequest{Target: world},
		).Buf(),
	)
	time.Sleep(time.Second)
	t := cmath.NewUserTransform()
	client.Send(
		posbus.NewMessageFromBuffer(posbus.TypeSendTransform, t.Bytes()).Buf(),
	)

	<-ctx.Done()
	fmt.Println("Stopped.")
	os.Exit(0)
}

func onMessage(msgType posbus.MsgType, data interface{}) error {

	r := make(map[string]interface{})
	err := utils.MapDecode(data, &r)
	//r, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err, posbus.MessageNameById(msgType))
	}
	fmt.Printf("Incoming message: %+v %+v\n", posbus.MessageNameById(msgType), r)
	return nil
}
