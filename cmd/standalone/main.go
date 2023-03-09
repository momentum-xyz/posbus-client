package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/momentum-xyz/posbus-client/pbc"
	"github.com/momentum-xyz/ubercontroller/pkg/posbus"
	"github.com/momentum-xyz/ubercontroller/universe/logic/api/dto"
	"io"
	"net/http"
	"time"
)

var URL = "https://dev2.odyssey.ninja/posbus"

func main() {
	//ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	//defer stop()
	ctx := context.Background()
	resp, err := http.Post("https://dev2.odyssey.ninja/api/v4/auth/guest-token", "", nil)
	if err != nil {
		fmt.Printf("Error of getting guest token %+v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	var u dto.User
	err = json.Unmarshal(body, &u)
	if err != nil {
		fmt.Printf("Error %+v\n", err)
		return

	}

	//fmt.Printf("%+v\n", u)

	client := pbc.NewClient(ctx)
	client.SetCallback(onMessage)
	client.Connect(URL, *u.JWTToken, uuid.MustParse(u.ID))

	client.Send(
		posbus.NewMessageFromData(
			posbus.TypeTeleportRequest, uuid.MustParse("975cb9ca-4dfa-4d35-adc2-198ed1f12555"),
		).Buf(),
	)

	time.Sleep(time.Second * 20000)

}

func onMessage(msgType posbus.MsgType, data interface{}) error {
	r, _ := json.Marshal(data)
	fmt.Printf("Incoming message: %+v %+v\n", posbus.MessageNameById(msgType), string(r))
	return nil
}
