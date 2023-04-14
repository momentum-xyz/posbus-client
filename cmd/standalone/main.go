package main

import (
	"context"
	"encoding/json"
	"errors"
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

	"github.com/momentum-xyz/posbus-client/pbc"
	"github.com/momentum-xyz/ubercontroller/pkg/cmath"
	"github.com/momentum-xyz/ubercontroller/pkg/posbus"
	"github.com/momentum-xyz/ubercontroller/universe/logic/api/dto"
	"github.com/momentum-xyz/ubercontroller/utils"
	"github.com/momentum-xyz/ubercontroller/utils/umid"
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
	world, err := umid.Parse(*worldArg)
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
	if err != nil {
		log.Fatalf("Error %+v\n", err)
	}

	var u dto.User
	err = json.Unmarshal(body, &u)
	if err != nil {
		log.Fatalf("Error %+v\n", err)
	}

	//fmt.Printf("%+v\n", u)

	connectionCtx, cancelConnection := context.WithCancel(ctx)
	defer cancelConnection()
	client := pbc.NewClient()
	client.SetCallback(onMessage)
	log.Printf("Connecting to %s as %s\n", URL, u.Name)
	client.Connect(connectionCtx, URL, *u.JWTToken, umid.MustParse(u.ID))

	log.Printf("Teleporting %v to %s\n", u.Name, world)
	client.Send(
		posbus.BinMessage(
			&posbus.TeleportRequest{Target: world},
		),
	)
	time.Sleep(time.Second)
	t := &cmath.TransformNoScale{}
	client.Send(posbus.BinMessage((*posbus.MyTransform)(t)))

	/* example reconnect:
	time.Sleep(time.Second * 3)
	cancelConnection()

	time.Sleep(time.Second * 3)
	client.Connect(ctx, URL, *u.JWTToken, uuid.MustParse(u.ID))
	*/

	<-ctx.Done()
	fmt.Println("Stopped.")
	os.Exit(0)
}

func onMessage(msg posbus.Message) error {

	r := make(map[string]interface{})
	err := utils.MapDecode(msg, &r)
	//r, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err, posbus.MessageNameById(msg.GetType()))
	}
	log.Printf("Incoming message: %+v %+v\n", posbus.MessageNameById(msg.GetType()), r)
	switch m := msg.(type) {
	case *posbus.Signal:
		return onSignal(m)
	}
	return nil
}

func onSignal(sig *posbus.Signal) error {
	switch sig.Value {
	case posbus.SignalNone:
		return errors.New("none signal received")
	case posbus.SignalWorldDoesNotExist:
		return errors.New("world does not exist signal received")
	case posbus.SignalDualConnection:
		return errors.New("dual connection signal received")
	case posbus.SignalConnected:
		log.Println("connected signal")
		return nil
	case posbus.SignalConnectionClosed:
		log.Println("connecting closed signal")
		return nil
	default:
		log.Printf("Unhandled signal %d\n", sig.Value)
	}
	return nil
}
