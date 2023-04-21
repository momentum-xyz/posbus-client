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
	"sync"
	"syscall"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/momentum-xyz/posbus-client/pbc"
	"github.com/momentum-xyz/posbus-client/test/scenarios"
	"github.com/momentum-xyz/ubercontroller/logger"
	"github.com/momentum-xyz/ubercontroller/pkg/posbus"
	"github.com/momentum-xyz/ubercontroller/universe/logic/api/dto"
	"github.com/momentum-xyz/ubercontroller/utils"
	"github.com/momentum-xyz/ubercontroller/utils/umid"
	"go.uber.org/zap/zapcore"
)

func main() {
	backendArg := flag.String("backend", "http://localhost:4000", "The URL to the controller backend")
	worldArg := flag.String("world", "975cb9ca-4dfa-4d35-adc2-198ed1f12555", "UUID of a world")
	token := flag.String("token", "", "An authentication token")
	nrFlyers := flag.Uint64("nrFlyers", 0, "Number of fake users to create that fly around randomly")
	logLevel := flag.String("log", "warn", "Log level (warn, info, debug")
	flag.Parse()
	backend, err := url.Parse(*backendArg)
	if err != nil {
		log.Fatalf("Invalid backend URL %s", err)
	}
	world, err := umid.Parse(*worldArg)
	if err != nil {
		log.Fatalf("Invalid world %s", err)
	}

	l, err := zapcore.ParseLevel(*logLevel)
	if err != nil {
		log.Fatalf("Invalid log level %s", err)
	}
	logger.SetLevel(l)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var user *dto.User
	if *token != "" {
		user, err = userFromToken(*token)
		if err != nil {
			log.Fatalf("Invalid userId %s", err)
		}
	} else {
		user, err = guestUser(backend)
		if err != nil {
			log.Fatalf("guest user: %s", err)
		}
	}

	//fmt.Printf("%+v\n", u)

	// 'viewer' for the ouput
	client := pbc.NewClient()
	client.SetCallback(onMessage)
	pbURL := backend.JoinPath("posbus").String()
	log.Printf("Connecting to %s as %s\n", pbURL, user.Name)
	client.Connect(ctx, pbURL, *user.JWTToken, umid.MustParse(user.ID))

	log.Printf("Teleporting %v to %s\n", user.Name, world)
	client.Send(
		posbus.BinMessage(
			&posbus.TeleportRequest{Target: world},
		),
	)

	// Run some fake users.
	// "poor man's" load test, just for some quick local testing :)
	// TODO: Use a proper testing framework, to not reinvent the wheel here.
	var wg sync.WaitGroup
	const rampUp = 420 * time.Millisecond
	log.Printf("Starting %d flyers...", *nrFlyers)
	for i := uint64(0); i < *nrFlyers; i++ {
		if ctx.Err() == nil {
			wg.Add(1)
			go func(i uint64) {
				defer wg.Done()
				scenarios.GuestFlyer(ctx, i, backend, &world)
			}(i)
			time.Sleep(rampUp)
		}
	}
	log.Println("done!")

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
	//log.Printf("Incoming message: %+v %+v\n", posbus.MessageNameById(msg.GetType()), r)
	switch m := msg.(type) {
	case *posbus.Signal:
		return onSignal(m)
	case *posbus.UsersTransformList:
		//fmt.Printf("User transform for %d users\n", len(m.Value))
	case *posbus.AttributeValueChanged:
		fmt.Printf("Attribute value changed: %+v", m)
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
		log.Println("connection closed signal")
		return nil
	default:
		log.Printf("Unhandled signal %d\n", sig.Value)
	}
	return nil
}

func userFromToken(tokenString string) (*dto.User, error) {
	fmt.Println(tokenString)
	token, _ := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(""), nil
	})
	claims := token.Claims.(*jwt.StandardClaims)
	return &dto.User{ // TODO: get with API call?
		ID:       claims.Subject,
		JWTToken: &tokenString,
		Name:     claims.Subject,
	}, nil
}

func guestUser(backend *url.URL) (*dto.User, error) {
	resp, err := http.Post(backend.JoinPath("/api/v4/auth/guest-token").String(), "", nil)
	if err != nil {
		return nil, fmt.Errorf("guest token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("guest token: %w", err)
	}
	u := &dto.User{}
	err = json.Unmarshal(body, u)
	if err != nil {
		return nil, fmt.Errorf("guest token: %w", err)
	}
	return u, nil
}
