///go:build e2e

package e2e

import (
	"backend/pkg/utils"
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
		const path = "/register"

		payload := map[string]interface{}{
			"email":    email,
			"password": password,
			"role":     "admin",
		}

		resp, err := app.makeRequest(http.MethodPost, path, payload)
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
		const path = "/login"

		payload := map[string]interface{}{
			"email":    email,
			"password": password,
		}

		resp, err := app.makeRequest(http.MethodPost, path, payload)
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
		const path = "/dummyLogin"

		var userToken, adminToken string

		userPayload := map[string]interface{}{"role": "user"}
		adminPayload := map[string]interface{}{"role": "admin"}

		resp, err := app.makeRequest(http.MethodPost, path, userPayload)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		defer func() { _ = resp.Body.Close() }()

		var token map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&token)
		require.NoError(t, err)

		userToken = token["token"].(string)
		assert.NotEmpty(t, userToken)

		resp, err = app.makeRequest(http.MethodPost, "/dummyLogin", adminPayload)
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

func TestAuth_ValidateAndConflicts(t *testing.T) {
	const (
		testEmail = "test129@gmail.com"
		testPass  = "test129"
	)

	app := setupE2E(t)
	_, _ = app.createUser(
		t, utils.VPtr(testEmail),
		utils.VPtr(testPass), nil,
	)

	t.Run("Register - Bad Cases", func(t *testing.T) {
		const path = "/register"

		type testCase struct {
			name           string
			payload        map[string]interface{}
			expectedStatus int
			expectedError  string
		}

		var tests = []testCase{
			{
				name: "Bad Request - Email Not Specified",
				payload: map[string]interface{}{
					"email":    "",
					"password": testPass,
					"role":     "user",
				},
				expectedStatus: http.StatusBadRequest,
				expectedError:  "invalid input",
			},
			{
				name: "Bad Request - Password Not Specified",
				payload: map[string]interface{}{
					"email":    "test123@gmail.com",
					"password": "",
					"role":     "user",
				},
				expectedStatus: http.StatusBadRequest,
				expectedError:  "invalid input",
			},
			{
				name: "Bad Request - Role Not Specified",
				payload: map[string]interface{}{
					"email":    "test123@gmail.com",
					"password": testPass,
					"role":     "",
				},
				expectedStatus: http.StatusBadRequest,
				expectedError:  "invalid input",
			},
			{
				name: "Bad Request - Invalid Role",
				payload: map[string]interface{}{
					"email":    "test123@gmail.com",
					"password": testPass,
					"role":     "backpacker",
				},
				expectedStatus: http.StatusBadRequest,
				expectedError:  "invalid input",
			},
			{
				name: "Conflict - User Already Exists",
				payload: map[string]interface{}{
					"email":    testEmail,
					"password": testPass,
					"role":     "user",
				},
				expectedStatus: http.StatusConflict,
				expectedError:  "user with given email already exists",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				resp, err := app.makeRequest(http.MethodPost, path, tt.payload)
				require.NoError(t, err)

				defer func() { _ = resp.Body.Close() }()

				assert.Equal(t, tt.expectedStatus, resp.StatusCode)

				var errResp map[string]string
				_ = json.NewDecoder(resp.Body).Decode(&errResp)
				assert.Contains(t, errResp["error"], tt.expectedError)
			})
		}
	})

	t.Run("Login - Bad Cases", func(t *testing.T) {
		const path = "/login"

		type testCase struct {
			name           string
			payload        map[string]interface{}
			expectedStatus int
			expectedError  string
		}

		var tests = []testCase{
			{
				name: "Unauthorized - Invalid Password",
				payload: map[string]interface{}{
					"email":    testEmail,
					"password": "invalid123",
				},
				expectedStatus: http.StatusUnauthorized,
				expectedError:  "invalid email or password",
			},
			{
				name: "Not Found - User Does Not Exist",
				payload: map[string]interface{}{
					"email":    "test123@gmail.com",
					"password": testPass,
				},
				expectedStatus: http.StatusUnauthorized,
				expectedError:  "invalid email or password",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				resp, err := app.makeRequest(http.MethodPost, path, tt.payload)
				require.NoError(t, err)

				defer func() { _ = resp.Body.Close() }()

				assert.Equal(t, tt.expectedStatus, resp.StatusCode)

				var errResp map[string]string
				_ = json.NewDecoder(resp.Body).Decode(&errResp)
				assert.Contains(t, errResp["error"], tt.expectedError)
			})
		}
	})

	t.Run("Dummy Login - Bad Cases", func(t *testing.T) {
		const path = "/dummyLogin"

		type testCase struct {
			name           string
			payload        map[string]interface{}
			expectedStatus int
			expectedError  string
		}

		var tests = []testCase{
			{
				name:           "Bad Request - Invalid Role",
				payload:        map[string]interface{}{"role": "backpacker"},
				expectedStatus: http.StatusBadRequest,
				expectedError:  "invalid input",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				resp, err := app.makeRequest(http.MethodPost, path, tt.payload)
				require.NoError(t, err)

				defer func() { _ = resp.Body.Close() }()

				assert.Equal(t, tt.expectedStatus, resp.StatusCode)

				var errResp map[string]string
				_ = json.NewDecoder(resp.Body).Decode(&errResp)
				assert.Contains(t, errResp["error"], tt.expectedError)
			})
		}
	})
}
