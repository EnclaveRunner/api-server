package main

import (
	"api-server/client"
	"context"
	"net/http"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
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

// ============================================================================
// Users Tests
// ============================================================================

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
			Id:      createduserid,
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

func TestCreateUserEmptyFields(t *testing.T) {
	t.Parallel()

	// Empty username
	resp1, err := c.PostUsersUserWithResponse(
		t.Context(),
		client.PostUsersUserJSONRequestBody{
			Name:     "",
			Password: "password",
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp1.StatusCode())

	// Empty password
	resp2, err := c.PostUsersUserWithResponse(
		t.Context(),
		client.PostUsersUserJSONRequestBody{
			Name:     "testuser",
			Password: "",
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp2.StatusCode())

	// Both empty
	resp3, err := c.PostUsersUserWithResponse(
		t.Context(),
		client.PostUsersUserJSONRequestBody{
			Name:     "",
			Password: "",
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp3.StatusCode())
}

func TestCreateUserDuplicateName(t *testing.T) {
	t.Parallel()
	username := "testUserDuplicate"
	password := "test"

	// Create first user
	createResp1, err := c.PostUsersUserWithResponse(
		t.Context(),
		client.PostUsersUserJSONRequestBody{
			Name:     username,
			Password: password,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, createResp1.StatusCode())
	userId := createResp1.JSON201.Id

	// Try to create user with same name
	createResp2, err := c.PostUsersUserWithResponse(
		t.Context(),
		client.PostUsersUserJSONRequestBody{
			Name:     username,
			Password: "different",
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusConflict, createResp2.StatusCode())

	// Cleanup
	_, _ = c.DeleteUsersUserWithResponse(
		t.Context(),
		client.DeleteUsersUserJSONRequestBody{
			Id: userId,
		},
	)
}

func TestGetUserInvalidUUID(t *testing.T) {
	t.Parallel()

	resp, err := c.GetUsersUserWithResponse(
		t.Context(),
		client.GetUsersUserJSONRequestBody{
			Id: "not-a-valid-uuid",
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode())
}

func TestGetUserNotFound(t *testing.T) {
	t.Parallel()

	// Use a valid UUID that doesn't exist
	uuidRandom, _ := uuid.NewRandom()
	resp, err := c.GetUsersUserWithResponse(
		t.Context(),
		client.GetUsersUserJSONRequestBody{
			Id: uuidRandom.String(),
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode())
}

func TestHeadUserInvalidUUID(t *testing.T) {
	t.Parallel()

	resp, err := c.HeadUsersUserWithResponse(
		t.Context(),
		client.HeadUsersUserJSONRequestBody{
			Id: "not-a-valid-uuid",
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode())
}

func TestPatchUserInvalidUUID(t *testing.T) {
	t.Parallel()
	newName := "newName"

	resp, err := c.PatchUsersUserWithResponse(
		t.Context(),
		client.PatchUsersUserJSONRequestBody{
			Id:      "not-a-valid-uuid",
			NewName: &newName,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode())
}

func TestPatchUserNotFound(t *testing.T) {
	t.Parallel()
	newName := "newName"

	uuidRandom, _ := uuid.NewRandom()
	resp, err := c.PatchUsersUserWithResponse(
		t.Context(),
		client.PatchUsersUserJSONRequestBody{
			Id:      uuidRandom.String(),
			NewName: &newName,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode())
}

func TestPatchUserDuplicateName(t *testing.T) {
	t.Parallel()

	// Create two users
	user1Resp, err := c.PostUsersUserWithResponse(
		t.Context(),
		client.PostUsersUserJSONRequestBody{
			Name:     "testPatchUser1",
			Password: "test",
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, user1Resp.StatusCode())
	user1Id := user1Resp.JSON201.Id

	user2Resp, err := c.PostUsersUserWithResponse(
		t.Context(),
		client.PostUsersUserJSONRequestBody{
			Name:     "testPatchUser2",
			Password: "test",
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, user2Resp.StatusCode())
	user2Id := user2Resp.JSON201.Id

	// Try to rename user2 to user1's name
	existingName := "testPatchUser1"
	patchResp, err := c.PatchUsersUserWithResponse(
		t.Context(),
		client.PatchUsersUserJSONRequestBody{
			Id:      user2Id,
			NewName: &existingName,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusConflict, patchResp.StatusCode())

	// Cleanup
	_, _ = c.DeleteUsersUserWithResponse(
		t.Context(),
		client.DeleteUsersUserJSONRequestBody{Id: user1Id},
	)
	_, _ = c.DeleteUsersUserWithResponse(
		t.Context(),
		client.DeleteUsersUserJSONRequestBody{Id: user2Id},
	)
}

func TestPatchUserOnlyPassword(t *testing.T) {
	t.Parallel()

	// Create user
	createResp, err := c.PostUsersUserWithResponse(
		t.Context(),
		client.PostUsersUserJSONRequestBody{
			Name:     "testPatchPasswordOnly",
			Password: "oldpassword",
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, createResp.StatusCode())
	userId := createResp.JSON201.Id
	originalName := createResp.JSON201.Name

	// Update only password
	newPassword := "newpassword"
	patchResp, err := c.PatchUsersUserWithResponse(
		t.Context(),
		client.PatchUsersUserJSONRequestBody{
			Id:          userId,
			NewPassword: &newPassword,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, patchResp.StatusCode())
	assert.Equal(
		t,
		originalName,
		patchResp.JSON200.Name,
		"Name should remain unchanged",
	)

	// Cleanup
	_, _ = c.DeleteUsersUserWithResponse(
		t.Context(),
		client.DeleteUsersUserJSONRequestBody{Id: userId},
	)
}

func TestPatchUserOnlyName(t *testing.T) {
	t.Parallel()

	// Create user
	createResp, err := c.PostUsersUserWithResponse(
		t.Context(),
		client.PostUsersUserJSONRequestBody{
			Name:     "testPatchNameOnly",
			Password: "password",
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, createResp.StatusCode())
	userId := createResp.JSON201.Id

	// Update only name
	newName := "testPatchNameOnlyUpdated"
	patchResp, err := c.PatchUsersUserWithResponse(
		t.Context(),
		client.PatchUsersUserJSONRequestBody{
			Id:      userId,
			NewName: &newName,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, patchResp.StatusCode())
	assert.Equal(t, newName, patchResp.JSON200.Name)

	// Cleanup
	_, _ = c.DeleteUsersUserWithResponse(
		t.Context(),
		client.DeleteUsersUserJSONRequestBody{Id: userId},
	)
}

func TestDeleteUserInvalidUUID(t *testing.T) {
	t.Parallel()

	resp, err := c.DeleteUsersUserWithResponse(
		t.Context(),
		client.DeleteUsersUserJSONRequestBody{
			Id: "not-a-valid-uuid",
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode())
}

func TestDeleteUserNotFound(t *testing.T) {
	t.Parallel()

	uuidRandom, _ := uuid.NewRandom()
	resp, err := c.DeleteUsersUserWithResponse(
		t.Context(),
		client.DeleteUsersUserJSONRequestBody{
			Id: uuidRandom.String(),
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode())
}

func TestGetUsersMe(t *testing.T) {
	t.Parallel()

	// Get current user (admin)
	resp, err := c.GetUsersMeWithResponse(t.Context())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.NotNil(t, resp.JSON200)
	assert.Equal(t, adminUsername, resp.JSON200.Name)
	assert.NotEmpty(t, resp.JSON200.Id)
}

func TestGetUsersMeUnauthenticated(t *testing.T) {
	t.Skip("Needs default RBAC config implementation")
	t.Parallel()

	// Create client without auth
	noAuthClient, err := client.NewClientWithResponses("http://localhost:8080")
	assert.NoError(t, err)

	resp, err := noAuthClient.GetUsersMeWithResponse(t.Context())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode())
}

func TestPatchUsersMe(t *testing.T) {
	t.Skip("Needs default RBAC config implementation")
	t.Parallel()

	// Create a test user to avoid modifying the admin user
	createResp, err := c.PostUsersUserWithResponse(
		t.Context(),
		client.PostUsersUserJSONRequestBody{
			Name:     "testPatchMe",
			Password: "password",
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, createResp.StatusCode())
	userId := createResp.JSON201.Id
	username := createResp.JSON201.Name

	// Create client with the new user's credentials
	userClient, err := client.NewClientWithResponses("http://localhost:8080",
		func(c *client.Client) error {
			c.RequestEditors = []client.RequestEditorFn{
				func(ctx context.Context, req *http.Request) error {
					_, _, ok := req.BasicAuth()
					if !ok {
						req.SetBasicAuth(username, "password")
					}

					return nil
				},
			}

			return nil
		},
	)
	assert.NoError(t, err)

	// Update name
	newName := "testPatchMeUpdated"
	patchResp, err := userClient.PatchUsersMeWithResponse(
		t.Context(),
		client.PatchUsersMeJSONRequestBody{
			NewName: &newName,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, patchResp.StatusCode())
	assert.Equal(t, newName, patchResp.JSON200.Name)

	// Verify the update
	getResp, err := userClient.GetUsersMeWithResponse(t.Context())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, getResp.StatusCode())
	assert.Equal(t, newName, getResp.JSON200.Name)

	// Cleanup (using admin client)
	_, _ = c.DeleteUsersUserWithResponse(
		t.Context(),
		client.DeleteUsersUserJSONRequestBody{Id: userId},
	)
}

func TestPatchUsersMePassword(t *testing.T) {
	t.Skip("Needs default RBAC configuration implementation")
	t.Parallel()

	// Create a test user
	createResp, err := c.PostUsersUserWithResponse(
		t.Context(),
		client.PostUsersUserJSONRequestBody{
			Name:     "testPatchMePassword",
			Password: "oldpassword",
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, createResp.StatusCode())
	userId := createResp.JSON201.Id
	username := createResp.JSON201.Name

	// Create client with the user's credentials
	userClient, err := client.NewClientWithResponses("http://localhost:8080",
		func(c *client.Client) error {
			c.RequestEditors = []client.RequestEditorFn{
				func(ctx context.Context, req *http.Request) error {
					_, _, ok := req.BasicAuth()
					if !ok {
						req.SetBasicAuth(username, "oldpassword")
					}

					return nil
				},
			}

			return nil
		},
	)
	assert.NoError(t, err)

	// Update password
	newPassword := "newpassword"
	patchResp, err := userClient.PatchUsersMeWithResponse(
		t.Context(),
		client.PatchUsersMeJSONRequestBody{
			NewPassword: &newPassword,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, patchResp.StatusCode())

	// Create new client with new password to verify it works
	newClient, err := client.NewClientWithResponses("http://localhost:8080",
		func(c *client.Client) error {
			c.RequestEditors = []client.RequestEditorFn{
				func(ctx context.Context, req *http.Request) error {
					_, _, ok := req.BasicAuth()
					if !ok {
						req.SetBasicAuth(username, newPassword)
					}

					return nil
				},
			}

			return nil
		},
	)
	assert.NoError(t, err)

	// Verify we can authenticate with new password
	getResp, err := newClient.GetUsersMeWithResponse(t.Context())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, getResp.StatusCode())

	// Cleanup (using admin client)
	_, _ = c.DeleteUsersUserWithResponse(
		t.Context(),
		client.DeleteUsersUserJSONRequestBody{Id: userId},
	)
}

func TestPatchUsersMeUnauthenticated(t *testing.T) {
	t.Skip("Needs default RBAC configuration implementation")
	t.Parallel()

	// Create client without auth
	noAuthClient, err := client.NewClientWithResponses("http://localhost:8080")
	assert.NoError(t, err)

	newName := "shouldNotWork"
	resp, err := noAuthClient.PatchUsersMeWithResponse(
		t.Context(),
		client.PatchUsersMeJSONRequestBody{
			NewName: &newName,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode())
}

func TestPatchUsersMeDuplicateName(t *testing.T) {
	t.Skip("Needs default RBAC configuration implementation")
	t.Parallel()

	// Create first user
	user1Resp, err := c.PostUsersUserWithResponse(
		t.Context(),
		client.PostUsersUserJSONRequestBody{
			Name:     "testPatchMeUser1",
			Password: "password",
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, user1Resp.StatusCode())
	user1Id := user1Resp.JSON201.Id

	// Create second user
	user2Resp, err := c.PostUsersUserWithResponse(
		t.Context(),
		client.PostUsersUserJSONRequestBody{
			Name:     "testPatchMeUser2",
			Password: "password",
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, user2Resp.StatusCode())
	user2Id := user2Resp.JSON201.Id
	user2Name := user2Resp.JSON201.Name

	// Create client with user2's credentials
	user2Client, err := client.NewClientWithResponses("http://localhost:8080",
		func(c *client.Client) error {
			c.RequestEditors = []client.RequestEditorFn{
				func(ctx context.Context, req *http.Request) error {
					_, _, ok := req.BasicAuth()
					if !ok {
						req.SetBasicAuth(user2Name, "password")
					}

					return nil
				},
			}

			return nil
		},
	)
	assert.NoError(t, err)

	// Try to update user2's name to user1's name
	conflictName := "testPatchMeUser1"
	patchResp, err := user2Client.PatchUsersMeWithResponse(
		t.Context(),
		client.PatchUsersMeJSONRequestBody{
			NewName: &conflictName,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusConflict, patchResp.StatusCode())

	// Cleanup
	_, _ = c.DeleteUsersUserWithResponse(
		t.Context(),
		client.DeleteUsersUserJSONRequestBody{Id: user1Id},
	)
	_, _ = c.DeleteUsersUserWithResponse(
		t.Context(),
		client.DeleteUsersUserJSONRequestBody{Id: user2Id},
	)
}

// ============================================================================
// RBAC Tests
// ============================================================================

// Test Role CRUD operations
func TestRoleCRUD(t *testing.T) {
	t.Parallel()
	roleName := "testRoleCRUD"

	// Create Role
	createResp, err := c.PostRbacRoleWithResponse(
		t.Context(),
		client.PostRbacRoleJSONRequestBody{
			Role: roleName,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, createResp.StatusCode())

	// Check Role Exists
	headResp, err := c.HeadRbacRoleWithResponse(
		t.Context(),
		client.HeadRbacRoleJSONRequestBody{
			Role: roleName,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, headResp.StatusCode())

	// Get Role (users in role)
	getResp, err := c.GetRbacRoleWithResponse(
		t.Context(),
		client.GetRbacRoleJSONRequestBody{
			Role: roleName,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, getResp.StatusCode())
	assert.NotNil(t, getResp.JSON200)
	assert.Equal(t, 0, len(*getResp.JSON200), "New role should have no users")

	// List all roles
	listResp, err := c.GetRbacListRolesWithResponse(t.Context())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, listResp.StatusCode())
	assert.True(t, slices.Contains(*listResp.JSON200, roleName))

	// Delete Role
	deleteResp, err := c.DeleteRbacRoleWithResponse(
		t.Context(),
		client.DeleteRbacRoleJSONRequestBody{
			Role: roleName,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, deleteResp.StatusCode())

	// Verify Role Deletion
	headRespAfterDelete, err := c.HeadRbacRoleWithResponse(
		t.Context(),
		client.HeadRbacRoleJSONRequestBody{
			Role: roleName,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, headRespAfterDelete.StatusCode())
}

func TestGetRoleNotFound(t *testing.T) {
	t.Parallel()

	resp, err := c.GetRbacRoleWithResponse(
		t.Context(),
		client.GetRbacRoleJSONRequestBody{
			Role: "nonExistentRole",
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode())
}

func TestDeleteRoleNotFound(t *testing.T) {
	t.Parallel()

	resp, err := c.DeleteRbacRoleWithResponse(
		t.Context(),
		client.DeleteRbacRoleJSONRequestBody{
			Role: "nonExistentRole",
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode())
}

// Test Resource Group CRUD operations
func TestResourceGroupCRUD(t *testing.T) {
	t.Parallel()
	rgName := "testResourceGroupCRUD"

	// Create Resource Group
	createResp, err := c.PostRbacResourceGroupWithResponse(
		t.Context(),
		client.PostRbacResourceGroupJSONRequestBody{
			ResourceGroup: rgName,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, createResp.StatusCode())

	// Check Resource Group Exists
	headResp, err := c.HeadRbacResourceGroupWithResponse(
		t.Context(),
		client.HeadRbacResourceGroupJSONRequestBody{
			ResourceGroup: rgName,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, headResp.StatusCode())

	// Get Resource Group
	getResp, err := c.GetRbacResourceGroupWithResponse(
		t.Context(),
		client.GetRbacResourceGroupJSONRequestBody{
			ResourceGroup: rgName,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, getResp.StatusCode())
	assert.NotNil(t, getResp.JSON200)
	assert.Equal(
		t,
		0,
		len(*getResp.JSON200),
		"New resource group should have no endpoints",
	)

	// List all resource groups
	listResp, err := c.GetRbacListResourceGroupsWithResponse(t.Context())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, listResp.StatusCode())
	assert.True(t, slices.Contains(*listResp.JSON200, rgName))

	// Delete Resource Group
	deleteResp, err := c.DeleteRbacResourceGroupWithResponse(
		t.Context(),
		client.DeleteRbacResourceGroupJSONRequestBody{
			ResourceGroup: rgName,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, deleteResp.StatusCode())

	// Verify Resource Group Deletion
	headRespAfterDelete, err := c.HeadRbacResourceGroupWithResponse(
		t.Context(),
		client.HeadRbacResourceGroupJSONRequestBody{
			ResourceGroup: rgName,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, headRespAfterDelete.StatusCode())
}

func TestGetResourceGroupNotFound(t *testing.T) {
	t.Parallel()

	resp, err := c.GetRbacResourceGroupWithResponse(
		t.Context(),
		client.GetRbacResourceGroupJSONRequestBody{
			ResourceGroup: "nonExistentResourceGroup",
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode())
}

func TestDeleteResourceGroupNotFound(t *testing.T) {
	t.Parallel()

	resp, err := c.DeleteRbacResourceGroupWithResponse(
		t.Context(),
		client.DeleteRbacResourceGroupJSONRequestBody{
			ResourceGroup: "nonExistentResourceGroup",
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode())
}

// Test Endpoint to Resource Group assignment
func TestEndpointResourceGroupAssignment(t *testing.T) {
	t.Parallel()
	rgName := "testEndpointRG"
	endpoint := "/test/endpoint"

	// Create Resource Group
	createRGResp, err := c.PostRbacResourceGroupWithResponse(
		t.Context(),
		client.PostRbacResourceGroupJSONRequestBody{
			ResourceGroup: rgName,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, createRGResp.StatusCode())

	// Assign endpoint to resource group
	assignResp, err := c.PostRbacEndpointWithResponse(
		t.Context(),
		client.PostRbacEndpointJSONRequestBody{
			ResourceGroup: rgName,
			Endpoint:      endpoint,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, assignResp.StatusCode())

	// Get resource groups for endpoint
	getEndpointResp, err := c.GetRbacEndpointWithResponse(
		t.Context(),
		client.GetRbacEndpointJSONRequestBody{
			Endpoint: endpoint,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, getEndpointResp.StatusCode())
	assert.True(t, slices.Contains(*getEndpointResp.JSON200, rgName))

	// Get endpoints in resource group
	getRGResp, err := c.GetRbacResourceGroupWithResponse(
		t.Context(),
		client.GetRbacResourceGroupJSONRequestBody{
			ResourceGroup: rgName,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, getRGResp.StatusCode())
	assert.True(t, slices.Contains(*getRGResp.JSON200, endpoint))

	// Remove endpoint from resource group
	removeResp, err := c.DeleteRbacEndpointWithResponse(
		t.Context(),
		client.DeleteRbacEndpointJSONRequestBody{
			ResourceGroup: rgName,
			Endpoint:      endpoint,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, removeResp.StatusCode())

	// Verify endpoint removed
	getRGRespAfter, err := c.GetRbacResourceGroupWithResponse(
		t.Context(),
		client.GetRbacResourceGroupJSONRequestBody{
			ResourceGroup: rgName,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, getRGRespAfter.StatusCode())
	assert.False(t, slices.Contains(*getRGRespAfter.JSON200, endpoint))

	// Cleanup
	_, _ = c.DeleteRbacResourceGroupWithResponse(
		t.Context(),
		client.DeleteRbacResourceGroupJSONRequestBody{ResourceGroup: rgName},
	)
}

func TestAssignEndpointToNonExistentResourceGroup(t *testing.T) {
	t.Parallel()

	resp, err := c.PostRbacEndpointWithResponse(
		t.Context(),
		client.PostRbacEndpointJSONRequestBody{
			ResourceGroup: "nonExistentRG",
			Endpoint:      "/test/endpoint",
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode())
}

func TestRemoveEndpointFromNonExistentResourceGroup(t *testing.T) {
	t.Parallel()

	resp, err := c.DeleteRbacEndpointWithResponse(
		t.Context(),
		client.DeleteRbacEndpointJSONRequestBody{
			ResourceGroup: "nonExistentRG",
			Endpoint:      "/test/endpoint",
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode())
}

// Test User to Role assignment
func TestUserRoleAssignment(t *testing.T) {
	t.Parallel()
	username := "testUserRole"
	password := "test"
	roleName := "testUserRoleAssignment"

	// Create user
	createUserResp, err := c.PostUsersUserWithResponse(
		t.Context(),
		client.PostUsersUserJSONRequestBody{
			Name:     username,
			Password: password,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, createUserResp.StatusCode())
	userId := createUserResp.JSON201.Id

	// Create role
	createRoleResp, err := c.PostRbacRoleWithResponse(
		t.Context(),
		client.PostRbacRoleJSONRequestBody{
			Role: roleName,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, createRoleResp.StatusCode())

	// Assign role to user
	assignResp, err := c.PostRbacUserWithResponse(
		t.Context(),
		client.PostRbacUserJSONRequestBody{
			UserId: userId,
			Role:   roleName,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, assignResp.StatusCode())

	// Get roles for user
	getUserRolesResp, err := c.GetRbacUserWithResponse(
		t.Context(),
		client.GetRbacUserJSONRequestBody{
			UserId: userId,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, getUserRolesResp.StatusCode())
	assert.True(t, slices.Contains(*getUserRolesResp.JSON200, roleName))

	// Get users in role
	getRoleUsersResp, err := c.GetRbacRoleWithResponse(
		t.Context(),
		client.GetRbacRoleJSONRequestBody{
			Role: roleName,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, getRoleUsersResp.StatusCode())
	assert.True(t, slices.Contains(*getRoleUsersResp.JSON200, userId))

	// Remove role from user
	removeResp, err := c.DeleteRbacUserWithResponse(
		t.Context(),
		client.DeleteRbacUserJSONRequestBody{
			UserId: userId,
			Role:   roleName,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, removeResp.StatusCode())

	// Verify role removed
	getUserRolesRespAfter, err := c.GetRbacUserWithResponse(
		t.Context(),
		client.GetRbacUserJSONRequestBody{
			UserId: userId,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, getUserRolesRespAfter.StatusCode())
	assert.False(t, slices.Contains(*getUserRolesRespAfter.JSON200, roleName))

	// Cleanup
	_, _ = c.DeleteRbacRoleWithResponse(
		t.Context(),
		client.DeleteRbacRoleJSONRequestBody{Role: roleName},
	)
	_, _ = c.DeleteUsersUserWithResponse(
		t.Context(),
		client.DeleteUsersUserJSONRequestBody{Id: userId},
	)
}

func TestGetRolesForUserInvalidUUID(t *testing.T) {
	t.Parallel()

	resp, err := c.GetRbacUserWithResponse(
		t.Context(),
		client.GetRbacUserJSONRequestBody{
			UserId: "invalid-uuid",
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode())
}

func TestGetRolesForNonExistentUser(t *testing.T) {
	t.Parallel()

	uuidRandom, _ := uuid.NewRandom()
	resp, err := c.GetRbacUserWithResponse(
		t.Context(),
		client.GetRbacUserJSONRequestBody{
			UserId: uuidRandom.String(),
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode())
}

func TestAssignRoleToUserInvalidUUID(t *testing.T) {
	t.Parallel()

	resp, err := c.PostRbacUserWithResponse(
		t.Context(),
		client.PostRbacUserJSONRequestBody{
			UserId: "invalid-uuid",
			Role:   "someRole",
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode())
}

func TestAssignRoleToNonExistentUser(t *testing.T) {
	t.Parallel()

	uuidRandom, _ := uuid.NewRandom()
	resp, err := c.PostRbacUserWithResponse(
		t.Context(),
		client.PostRbacUserJSONRequestBody{
			UserId: uuidRandom.String(),
			Role:   "someRole",
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode())
}

func TestAssignNonExistentRoleToUser(t *testing.T) {
	t.Parallel()

	// Create user
	createUserResp, err := c.PostUsersUserWithResponse(
		t.Context(),
		client.PostUsersUserJSONRequestBody{
			Name:     "testUserNonExistentRole",
			Password: "test",
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, createUserResp.StatusCode())
	userId := createUserResp.JSON201.Id

	// Try to assign non-existent role
	resp, err := c.PostRbacUserWithResponse(
		t.Context(),
		client.PostRbacUserJSONRequestBody{
			UserId: userId,
			Role:   "nonExistentRole",
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode())

	// Cleanup
	_, _ = c.DeleteUsersUserWithResponse(
		t.Context(),
		client.DeleteUsersUserJSONRequestBody{Id: userId},
	)
}

func TestRemoveRoleFromUserInvalidUUID(t *testing.T) {
	t.Parallel()

	resp, err := c.DeleteRbacUserWithResponse(
		t.Context(),
		client.DeleteRbacUserJSONRequestBody{
			UserId: "invalid-uuid",
			Role:   "someRole",
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode())
}

func TestRemoveRoleFromNonExistentUser(t *testing.T) {
	t.Parallel()

	uuidRandom, _ := uuid.NewRandom()
	resp, err := c.DeleteRbacUserWithResponse(
		t.Context(),
		client.DeleteRbacUserJSONRequestBody{
			UserId: uuidRandom.String(),
			Role:   "someRole",
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode())
}

func TestRemoveNonExistentRoleFromUser(t *testing.T) {
	t.Parallel()

	// Create user
	createUserResp, err := c.PostUsersUserWithResponse(
		t.Context(),
		client.PostUsersUserJSONRequestBody{
			Name:     "testUserRemoveNonExistentRole",
			Password: "test",
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, createUserResp.StatusCode())
	userId := createUserResp.JSON201.Id

	// Try to remove non-existent role
	resp, err := c.DeleteRbacUserWithResponse(
		t.Context(),
		client.DeleteRbacUserJSONRequestBody{
			UserId: userId,
			Role:   "nonExistentRole",
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode())

	// Cleanup
	_, _ = c.DeleteUsersUserWithResponse(
		t.Context(),
		client.DeleteUsersUserJSONRequestBody{Id: userId},
	)
}

// Test Policy CRUD operations
func TestPolicyCRUD(t *testing.T) {
	t.Parallel()
	roleName := "testPolicyRole"
	rgName := "testPolicyRG"
	permission := client.GET

	// Create role
	createRoleResp, err := c.PostRbacRoleWithResponse(
		t.Context(),
		client.PostRbacRoleJSONRequestBody{
			Role: roleName,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, createRoleResp.StatusCode())

	// Create resource group
	createRGResp, err := c.PostRbacResourceGroupWithResponse(
		t.Context(),
		client.PostRbacResourceGroupJSONRequestBody{
			ResourceGroup: rgName,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, createRGResp.StatusCode())

	// Create policy
	createPolicyResp, err := c.PostRbacPolicyWithResponse(
		t.Context(),
		client.PostRbacPolicyJSONRequestBody{
			Role:          roleName,
			ResourceGroup: rgName,
			Permission:    permission,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, createPolicyResp.StatusCode())

	// List policies and verify
	listResp, err := c.GetRbacPolicyWithResponse(t.Context())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, listResp.StatusCode())
	policyExists := slices.ContainsFunc(
		*listResp.JSON200,
		func(p client.RBACPolicy) bool {
			return p.Role == roleName &&
				p.ResourceGroup == rgName &&
				p.Permission == permission
		},
	)
	assert.True(t, policyExists, "Created policy should exist in list")

	// Delete policy
	deleteResp, err := c.DeleteRbacPolicyWithResponse(
		t.Context(),
		client.DeleteRbacPolicyJSONRequestBody{
			Role:          roleName,
			ResourceGroup: rgName,
			Permission:    permission,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, deleteResp.StatusCode())

	// Verify policy deleted
	listRespAfter, err := c.GetRbacPolicyWithResponse(t.Context())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, listRespAfter.StatusCode())
	policyExistsAfter := slices.ContainsFunc(
		*listRespAfter.JSON200,
		func(p client.RBACPolicy) bool {
			return p.Role == roleName &&
				p.ResourceGroup == rgName &&
				p.Permission == permission
		},
	)
	assert.False(t, policyExistsAfter, "Deleted policy should not exist")

	// Cleanup
	_, _ = c.DeleteRbacResourceGroupWithResponse(
		t.Context(),
		client.DeleteRbacResourceGroupJSONRequestBody{ResourceGroup: rgName},
	)
	_, _ = c.DeleteRbacRoleWithResponse(
		t.Context(),
		client.DeleteRbacRoleJSONRequestBody{Role: roleName},
	)
}

func TestCreatePolicyWithWildcards(t *testing.T) {
	t.Parallel()

	// Create policy with wildcards
	createResp, err := c.PostRbacPolicyWithResponse(
		t.Context(),
		client.PostRbacPolicyJSONRequestBody{
			Role:          "*",
			ResourceGroup: "*",
			Permission:    client.Asterisk,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, createResp.StatusCode())

	// List and verify
	listResp, err := c.GetRbacPolicyWithResponse(t.Context())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, listResp.StatusCode())
	policyExists := slices.ContainsFunc(
		*listResp.JSON200,
		func(p client.RBACPolicy) bool {
			return p.Role == "*" &&
				p.ResourceGroup == "*" &&
				p.Permission == client.Asterisk
		},
	)
	assert.True(t, policyExists)

	// Cleanup
	_, _ = c.DeleteRbacPolicyWithResponse(
		t.Context(),
		client.DeleteRbacPolicyJSONRequestBody{
			Role:          "*",
			ResourceGroup: "*",
			Permission:    client.Asterisk,
		},
	)
}

func TestCreatePolicyWithNonExistentRole(t *testing.T) {
	t.Parallel()
	rgName := "testPolicyNonExistentRole"

	// Create resource group
	createRGResp, err := c.PostRbacResourceGroupWithResponse(
		t.Context(),
		client.PostRbacResourceGroupJSONRequestBody{
			ResourceGroup: rgName,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, createRGResp.StatusCode())

	// Try to create policy with non-existent role
	createResp, err := c.PostRbacPolicyWithResponse(
		t.Context(),
		client.PostRbacPolicyJSONRequestBody{
			Role:          "nonExistentRole",
			ResourceGroup: rgName,
			Permission:    client.Asterisk,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, createResp.StatusCode())

	// Cleanup
	_, _ = c.DeleteRbacResourceGroupWithResponse(
		t.Context(),
		client.DeleteRbacResourceGroupJSONRequestBody{ResourceGroup: rgName},
	)
}

func TestCreatePolicyWithNonExistentResourceGroup(t *testing.T) {
	t.Parallel()
	roleName := "testPolicyNonExistentRG"

	// Create role
	createRoleResp, err := c.PostRbacRoleWithResponse(
		t.Context(),
		client.PostRbacRoleJSONRequestBody{
			Role: roleName,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, createRoleResp.StatusCode())

	// Try to create policy with non-existent resource group
	createResp, err := c.PostRbacPolicyWithResponse(
		t.Context(),
		client.PostRbacPolicyJSONRequestBody{
			Role:          roleName,
			ResourceGroup: "nonExistentRG",
			Permission:    client.Asterisk,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, createResp.StatusCode())

	// Cleanup
	_, _ = c.DeleteRbacRoleWithResponse(
		t.Context(),
		client.DeleteRbacRoleJSONRequestBody{Role: roleName},
	)
}

func TestMultiplePoliciesForSameRole(t *testing.T) {
	t.Parallel()
	roleName := "testMultiplePoliciesRole"
	rgName1 := "testMultiplePoliciesRG1"
	rgName2 := "testMultiplePoliciesRG2"

	// Create role
	createRoleResp, err := c.PostRbacRoleWithResponse(
		t.Context(),
		client.PostRbacRoleJSONRequestBody{
			Role: roleName,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, createRoleResp.StatusCode())

	// Create resource groups
	createRG1Resp, err := c.PostRbacResourceGroupWithResponse(
		t.Context(),
		client.PostRbacResourceGroupJSONRequestBody{
			ResourceGroup: rgName1,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, createRG1Resp.StatusCode())

	createRG2Resp, err := c.PostRbacResourceGroupWithResponse(
		t.Context(),
		client.PostRbacResourceGroupJSONRequestBody{
			ResourceGroup: rgName2,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, createRG2Resp.StatusCode())

	// Create multiple policies
	createPolicy1Resp, err := c.PostRbacPolicyWithResponse(
		t.Context(),
		client.PostRbacPolicyJSONRequestBody{
			Role:          roleName,
			ResourceGroup: rgName1,
			Permission:    client.GET,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, createPolicy1Resp.StatusCode())

	createPolicy2Resp, err := c.PostRbacPolicyWithResponse(
		t.Context(),
		client.PostRbacPolicyJSONRequestBody{
			Role:          roleName,
			ResourceGroup: rgName2,
			Permission:    client.POST,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, createPolicy2Resp.StatusCode())

	// List and verify both policies exist
	listResp, err := c.GetRbacPolicyWithResponse(t.Context())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, listResp.StatusCode())

	policy1Exists := slices.ContainsFunc(
		*listResp.JSON200,
		func(p client.RBACPolicy) bool {
			return p.Role == roleName &&
				p.ResourceGroup == rgName1 &&
				p.Permission == client.GET
		},
	)
	assert.True(t, policy1Exists)

	policy2Exists := slices.ContainsFunc(
		*listResp.JSON200,
		func(p client.RBACPolicy) bool {
			return p.Role == roleName &&
				p.ResourceGroup == rgName2 &&
				p.Permission == client.POST
		},
	)
	assert.True(t, policy2Exists)

	// Cleanup
	_, _ = c.DeleteRbacPolicyWithResponse(
		t.Context(),
		client.DeleteRbacPolicyJSONRequestBody{
			Role:          roleName,
			ResourceGroup: rgName1,
			Permission:    client.GET,
		},
	)
	_, _ = c.DeleteRbacPolicyWithResponse(
		t.Context(),
		client.DeleteRbacPolicyJSONRequestBody{
			Role:          roleName,
			ResourceGroup: rgName2,
			Permission:    client.POST,
		},
	)
	_, _ = c.DeleteRbacResourceGroupWithResponse(
		t.Context(),
		client.DeleteRbacResourceGroupJSONRequestBody{ResourceGroup: rgName1},
	)
	_, _ = c.DeleteRbacResourceGroupWithResponse(
		t.Context(),
		client.DeleteRbacResourceGroupJSONRequestBody{ResourceGroup: rgName2},
	)
	_, _ = c.DeleteRbacRoleWithResponse(
		t.Context(),
		client.DeleteRbacRoleJSONRequestBody{Role: roleName},
	)
}
