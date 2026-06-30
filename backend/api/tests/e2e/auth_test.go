///go:build e2e

package e2e

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuth_AllEndpoints(t *testing.T) {
	app := setupE2E(t)

	const (
		email    = "vovavova@gmail.com"
		password = "vovavova61"
	)

	t.Run("Register", func(t *testing.T) {
		payload := map[string]interface{}{
			"email":    email,
			"password": password,
			"role":     "admin",
		}

		resp, err := app.doRequest("POST", "/register", payload)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		defer func() { _ = resp.Body.Close() }()

		var user map[string]map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&user)
		require.NoError(t, err)

		// Check the user id
		userID := user["user"]["id"].(string)
		assert.NotEmpty(t, userID)

		_, err = uuid.Parse(userID)
		assert.NoError(t, err)

		// Check the email and the role
		assert.Equal(t, payload["email"], user["user"]["email"])
		assert.Equal(t, payload["role"], user["user"]["role"])
	})

	t.Run("Login", func(t *testing.T) {
		payload := map[string]interface{}{
			"email":    email,
			"password": password,
		}

		resp, err := app.doRequest("POST", "/login", payload)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		defer func() { _ = resp.Body.Close() }()

		var token map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&token)
		require.NoError(t, err)

		// Check the token
		tokenStr := token["token"].(string)
		assert.NotEmpty(t, tokenStr)
	})

	t.Run("Dummy Login", func(t *testing.T) {
		var userToken, adminToken string

		userPayload := map[string]interface{}{"role": "user"}
		adminPayload := map[string]interface{}{"role": "admin"}

		resp, err := app.doRequest("POST", "/dummyLogin", userPayload)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		defer func() { _ = resp.Body.Close() }()

		var token map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&token)
		require.NoError(t, err)

		userToken = token["token"].(string)
		assert.NotEmpty(t, userToken)

		resp, err = app.doRequest("POST", "/login", adminPayload)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		err = json.NewDecoder(resp.Body).Decode(&token)
		require.NoError(t, err)

		adminToken = token["token"].(string)
		assert.NotEmpty(t, adminToken)

		// Ensure that tokens are different
		assert.NotEqual(t, userToken, adminToken,
			"expected different tokens",
		)
	})
}
