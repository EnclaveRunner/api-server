package main

import (
	"api-server/client"
	"context"
	"net/http"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

const (
	adminPassword = "admin"
	adminUsername = "admin"
)

var c *client.ClientWithResponses

type RequestType string

const (
	RequestTypeJSON RequestType = "application/json"
)

func TestMain(m *testing.M) {
	viper.Set("human_readable_output", true)
	viper.Set("log_level", "debug")
	viper.Set("production_environment", false)
	viper.Set("port", 8080)
	viper.Set("database.host", "localhost")
	viper.Set("admin.username", adminUsername)
	viper.Set("admin.password", adminPassword)

	go main()

	cTmp, err := client.NewClientWithResponses("http://localhost:8080",
		func(c *client.Client) error {
			c.RequestEditors = []client.RequestEditorFn{
				func(ctx context.Context, req *http.Request) error {
					_, _, ok := req.BasicAuth()
					if !ok {
						req.SetBasicAuth(adminUsername, adminPassword)
					}

					return nil
				},
			}

			return nil
		},
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create API client")
	}

	c = cTmp

	log.Info().Msg("Waiting for the server to start...")
	time.Sleep(3 * time.Second)

	code := m.Run()

	os.Exit(code)
}

func TestAdminUserExists(t *testing.T) {
	t.Parallel()
	resp, err := c.GetUsersListWithResponse(t.Context())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())

	adminExists := slices.ContainsFunc(
		*resp.JSON200,
		func(element client.UserResponse) bool {
			return element.Name == adminUsername
		},
	)

	assert.True(t, adminExists, "Admin user should exist")
}

func TestUserCRUD(t *testing.T) {
	t.Parallel()
	username := "testUserCRUD"
	password := "test"
	// Create User
	createResp, err := c.PostUsersUserWithResponse(
		t.Context(),
		client.PostUsersUserJSONRequestBody{
			Name:     username,
			Password: password,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, createResp.StatusCode())
	createduserid := createResp.JSON201.Id
	assert.Equal(t, username, createResp.JSON201.Name)

	// Check User Exists
	headResp, err := c.HeadUsersUserWithResponse(
		t.Context(),
		client.HeadUsersUserJSONRequestBody{
			Id: createduserid,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, headResp.StatusCode())

	// Get User
	getResp, err := c.GetUsersUserWithResponse(
		t.Context(),
		client.GetUsersUserJSONRequestBody{
			Id: createduserid,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, getResp.StatusCode())
	assert.Equal(t, username, getResp.JSON200.Name)

	// Update User
	newUsername := "updatedTestUserCRUD"
	updateResp, err := c.PatchUsersUserWithResponse(
		t.Context(),
		client.PatchUsersUserJSONRequestBody{
			Id: createduserid,
			NewName: &newUsername,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, updateResp.StatusCode())
	assert.Equal(t, newUsername, updateResp.JSON200.Name)

	// Delete User
	deleteResp, err := c.DeleteUsersUserWithResponse(
		t.Context(),
		client.DeleteUsersUserJSONRequestBody{
			Id: createduserid,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, deleteResp.StatusCode())

	// Verify User Deletion
	headRespAfterDelete, err := c.HeadUsersUserWithResponse(
		t.Context(),
		client.HeadUsersUserJSONRequestBody{
			Id: createduserid,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, headRespAfterDelete.StatusCode())
}
