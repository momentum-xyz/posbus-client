package fixtures

import (
	"encoding/json"
	"io"
	"net/url"
	"testing"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/momentum-xyz/ubercontroller/types/entry"
	"github.com/momentum-xyz/ubercontroller/universe/logic/api/dto"
	"github.com/momentum-xyz/ubercontroller/universe/logic/common"
	"github.com/momentum-xyz/ubercontroller/utils/umid"
)

// Create a guest account in the backend through the API.
// Returns the guest user ID and authentication token (JWT).
// For use in integration tests.
func GuestAccount(t *testing.T, ctURL *url.URL) (umid.UMID, string) {
	url := ctURL.JoinPath("/api/v4/auth/guest-token").String()
	// retryable is current workaround for slow starting controller service
	resp, err := retryablehttp.Post(url, "", nil)
	if err != nil {
		t.Fatalf("Error getting guest token %+v\n", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("guest token: %s %s\n%s", url, resp.Status, string(body))
	}

	var u dto.User
	d := json.NewDecoder(resp.Body)
	err = d.Decode(&u)
	if err != nil {
		t.Fatalf("Error %+v\n", err)
	}

	userID, err := umid.Parse(u.ID)
	if err != nil {
		t.Fatalf("Error %+v\n", err)
	}
	token := *u.JWTToken

	return userID, token
}

// fixtory blueprint for users
func UserBlueprint(i int, last entry.User) entry.User {
	userId := umid.New()
	return entry.User{
		UserID:     userId,
		UserTypeID: getNormalUserTypeID(),
		Profile:    entry.UserProfile{},
	}
}

func getNormalUserTypeID() umid.UMID {
	typeId, err := common.GetNormalUserTypeID()
	if err != nil {
		panic(err)
	}
	return typeId
}
