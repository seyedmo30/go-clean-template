package api_test

import (
	"context"
	"net/http"
	"testing"

	sdk "github.com/seyedmo30/go-clean-template-client"
)

func TestGetUsers(t *testing.T) {

	client, err := sdk.NewClientWithResponses("http://localhost:8009")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	resp, err := client.GetUsersWithResponse(context.Background(), &sdk.GetUsersParams{})
	if err != nil {
		t.Fatalf("GetUsers failed: %v", err)
	}

	if resp.StatusCode() != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode())
	}

	if resp.JSON200 == nil || resp.JSON200.Users == nil {
		t.Fatal("expected JSON200.Users")
	}

	users := *resp.JSON200.Users
	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}

}

// func TestPostUsers(t *testing.T) {
// 	server := mockAPI(t)
// 	defer server.Close()

// 	client, err := sdk.NewClientWithResponses(server.URL)
// 	if err != nil {
// 		t.Fatalf("failed to create client: %v", err)
// 	}

// 	createReq := sdk.CreateUserRequestDTO{
// 		Username: "bob",
// 		Email:    sdk.OpenapiTypesEmail("bob@example.com"),
// 	}

// 	resp, err := client.PostUsersWithResponse(context.Background(), createReq)
// 	if err != nil {
// 		t.Fatalf("PostUsers failed: %v", err)
// 	}

// 	if resp.StatusCode() != http.StatusOK {
// 		t.Fatalf("expected 200, got %d", resp.StatusCode())
// 	}

// 	if resp.JSON200 == nil || resp.JSON200.Users == nil {
// 		t.Fatal("expected JSON200.Users")
// 	}

// 	user := (*resp.JSON200.Users)[0]

// 	if *user.Username != "bob" {
// 		t.Fatalf("unexpected username: %v", *user.Username)
// 	}
// 	if *user.Email != "bob@example.com" {
// 		t.Fatalf("unexpected email: %v", *user.Email)
// 	}
// }
