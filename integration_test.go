package main

import (
	"api-server/client"
	"context"
	"net/http"
	"os"
	"slices"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

const (
	adminPassword    = "admin"
	adminUsername    = "admin"
	adminDisplayName = "Administrator"
	defaultPassword  = "test"
)

var c *client.ClientWithResponses

func TestMain(m *testing.M) {
	_ = os.Setenv("ENCLAVE_ADMIN_USERNAME", adminUsername)
	_ = os.Setenv("ENCLAVE_ADMIN_PASSWORD", adminPassword)
	_ = os.Setenv("ENCLAVE_ADMIN_DISPLAY_NAME", adminDisplayName)
	_ = os.Setenv("ENCLAVE_DATABASE_HOST", "localhost")
	_ = os.Setenv("ENCLAVE_RETRY_RETENTION", "24h")

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
	for {
		_, err := c.GetV1UserMeWithResponse(context.Background())
		if err == nil {
			break
		}
	}

	code := m.Run()

	_ = os.Unsetenv("ENCLAVE_ADMIN_USERNAME")
	_ = os.Unsetenv("ENCLAVE_ADMIN_PASSWORD")
	_ = os.Unsetenv("ENCLAVE_ADMIN_DISPLAY_NAME")
	_ = os.Unsetenv("ENCLAVE_DATABASE_HOST")
	_ = os.Unsetenv("ENCLAVE_RETRY_RETENTION")

	os.Exit(code)
}

func createUser(
	t *testing.T,
	username, displayName, password string,
	roles ...string,
) {
	t.Helper()

	body := client.PutV1UserUsernameJSONRequestBody{
		Password:    password,
		DisplayName: displayName,
	}
	if len(roles) > 0 {
		body.Roles = &roles
	}

	resp, err := c.PutV1UserUsernameWithResponse(
		t.Context(),
		username,
		body,
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode())
}

func deleteUser(t *testing.T, username string) {
	t.Helper()

	_, _ = c.DeleteV1UserUsernameWithResponse(t.Context(), username)
}

func TestAdminUserExists(t *testing.T) {
	t.Parallel()

	resp, err := c.GetV1UserWithResponse(t.Context(), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.NotNil(t, resp.JSON200)

	adminExists := slices.ContainsFunc(
		*resp.JSON200,
		func(element client.UserResponse) bool {
			return element.Name == adminUsername &&
				element.DisplayName == adminDisplayName
		},
	)
	assert.True(t, adminExists)
}

func TestUserCRUDByUsername(t *testing.T) {
	t.Parallel()

	username := "test-user-crud"
	displayName := "Test User CRUD"
	createUser(t, username, displayName, defaultPassword)
	defer deleteUser(t, username)

	headResp, err := c.HeadV1UserUsernameWithResponse(t.Context(), username)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, headResp.StatusCode())

	getResp, err := c.GetV1UserUsernameWithResponse(t.Context(), username)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, getResp.StatusCode())
	assert.Equal(t, username, getResp.JSON200.Name)
	assert.Equal(t, displayName, getResp.JSON200.DisplayName)

	updatedDisplayName := "Updated Test User CRUD"
	patchResp, err := c.PatchV1UserUsernameWithResponse(
		t.Context(),
		username,
		client.PatchV1UserUsernameJSONRequestBody{DisplayName: &updatedDisplayName},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, patchResp.StatusCode())
	assert.Equal(t, username, patchResp.JSON200.Name)
	assert.Equal(t, updatedDisplayName, patchResp.JSON200.DisplayName)

	deleteResp, err := c.DeleteV1UserUsernameWithResponse(t.Context(), username)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, deleteResp.StatusCode())
	assert.Equal(t, username, deleteResp.JSON200.Name)

	headAfterDelete, err := c.HeadV1UserUsernameWithResponse(
		t.Context(),
		username,
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, headAfterDelete.StatusCode())
}

func TestPatchUserPasswordOnly(t *testing.T) {
	t.Parallel()

	username := "test-user-password-only"
	createUser(t, username, "Password User", "oldpassword")
	defer deleteUser(t, username)

	newPassword := "newpassword"
	patchResp, err := c.PatchV1UserUsernameWithResponse(
		t.Context(),
		username,
		client.PatchV1UserUsernameJSONRequestBody{Password: &newPassword},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, patchResp.StatusCode())
	assert.Equal(t, username, patchResp.JSON200.Name)
}

func TestPatchUserRolesOnly(t *testing.T) {
	t.Parallel()

	username := "test-user-roles-only"
	createUser(t, username, "Roles User", defaultPassword)
	defer deleteUser(t, username)

	role := "enclave_admin"
	patchResp, err := c.PatchV1UserUsernameWithResponse(
		t.Context(),
		username,
		client.PatchV1UserUsernameJSONRequestBody{Roles: &[]string{role}},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, patchResp.StatusCode())
	assert.Equal(t, username, patchResp.JSON200.Name)
	assert.NotNil(t, patchResp.JSON200.Roles)
	assert.Contains(t, *patchResp.JSON200.Roles, role)
}

func TestCreateUserWithRoles(t *testing.T) {
	t.Parallel()

	username := "test-user-with-roles"
	role := "enclave_admin"
	createUser(t, username, "Role User", defaultPassword, role)
	defer deleteUser(t, username)

	getResp, err := c.GetV1UserUsernameWithResponse(t.Context(), username)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, getResp.StatusCode())
	assert.NotNil(t, getResp.JSON200.Roles)
	assert.Contains(t, *getResp.JSON200.Roles, role)

	roleResp, err := c.GetV1RbacRoleRoleWithResponse(t.Context(), role)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, roleResp.StatusCode())
	assert.Contains(t, roleResp.JSON200.Users, username)
}

func TestGetUsersMeAllowedForRegularUser(t *testing.T) {
	t.Parallel()

	username := "test-user-me"
	password := "password"
	createUser(t, username, "Me User", password)
	defer deleteUser(t, username)

	userClient, err := client.NewClientWithResponses("http://localhost:8080",
		func(c *client.Client) error {
			c.RequestEditors = []client.RequestEditorFn{
				func(ctx context.Context, req *http.Request) error {
					_, _, ok := req.BasicAuth()
					if !ok {
						req.SetBasicAuth(username, password)
					}

					return nil
				},
			}

			return nil
		},
	)
	assert.NoError(t, err)

	meResp, err := userClient.GetV1UserMeWithResponse(t.Context())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, meResp.StatusCode())

	newDisplayName := "Updated Me User"
	patchResp, err := userClient.PatchV1UserMeWithResponse(
		t.Context(),
		client.PatchV1UserMeJSONRequestBody{DisplayName: &newDisplayName},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, patchResp.StatusCode())
}

func TestDeleteUsersMeUnauthenticated(t *testing.T) {
	t.Parallel()

	noAuthClient, err := client.NewClientWithResponses("http://localhost:8080")
	assert.NoError(t, err)

	resp, err := noAuthClient.DeleteV1UserMeWithResponse(t.Context())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode())
}

func TestRbacRoleUserAssignmentByUsername(t *testing.T) {
	t.Parallel()

	username := "test-rbac-user"
	role := "test-rbac-role"
	createUser(t, username, "RBAC User", defaultPassword)
	defer deleteUser(t, username)
	defer func() {
		_, _ = c.DeleteV1RbacRoleRoleWithResponse(t.Context(), role)
	}()

	putResp, err := c.PutV1RbacRoleRoleWithResponse(
		t.Context(),
		role,
		client.PutV1RbacRoleRoleJSONRequestBody{Users: []string{username}},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, putResp.StatusCode())
	if assert.NotNil(t, putResp.JSON201) {
		assert.True(t, slices.Contains(putResp.JSON201.Users, username))
	}

	getResp, err := c.GetV1RbacRoleRoleWithResponse(t.Context(), role)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, getResp.StatusCode())
	if assert.NotNil(t, getResp.JSON200) {
		assert.True(t, slices.Contains(getResp.JSON200.Users, username))
	}

	clearResp, err := c.PutV1RbacRoleRoleWithResponse(
		t.Context(),
		role,
		client.PutV1RbacRoleRoleJSONRequestBody{Users: []string{}},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, clearResp.StatusCode())
}

func TestRbacPolicyCRUD(t *testing.T) {
	t.Parallel()

	role := "test-policy-role"
	resourceGroup := "test-policy-rg"
	method := client.GET

	defer func() {
		_, _ = c.DeleteV1RbacRoleRoleWithResponse(t.Context(), role)
		_, _ = c.DeleteV1RbacResourceGroupResourceGroupWithResponse(
			t.Context(),
			resourceGroup,
		)
	}()

	_, err := c.PutV1RbacRoleRoleWithResponse(
		t.Context(),
		role,
		client.PutV1RbacRoleRoleJSONRequestBody{Users: []string{}},
	)
	assert.NoError(t, err)

	_, err = c.PutV1RbacResourceGroupResourceGroupWithResponse(
		t.Context(),
		resourceGroup,
		client.PutV1RbacResourceGroupResourceGroupJSONRequestBody{
			Endpoints: []string{"/v1/user"},
		},
	)
	assert.NoError(t, err)

	putPolicyResp, err := c.PutV1RbacPolicyWithResponse(
		t.Context(),
		client.PutV1RbacPolicyJSONRequestBody{
			Role:          role,
			ResourceGroup: resourceGroup,
			Method:        method,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, putPolicyResp.StatusCode())

	listResp, err := c.GetV1RbacPolicyWithResponse(t.Context(), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, listResp.StatusCode())
	assert.NotNil(t, listResp.JSON200)

	exists := slices.ContainsFunc(
		*listResp.JSON200,
		func(p client.RBACPolicy) bool {
			return p.Role == role && p.ResourceGroup == resourceGroup &&
				p.Method == method
		},
	)
	assert.True(t, exists)

	deleteResp, err := c.DeleteV1RbacPolicyWithResponse(
		t.Context(),
		client.DeleteV1RbacPolicyJSONRequestBody{
			Role:          role,
			ResourceGroup: resourceGroup,
			Method:        method,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, deleteResp.StatusCode())
}
