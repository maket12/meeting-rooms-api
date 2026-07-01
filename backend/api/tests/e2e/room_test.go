package e2e

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoom_AllEndpoints(t *testing.T) {
	app := setupE2E(t)
	adminToken := app.getAdminToken(t)
	userToken := app.getUserToken(t)

	t.Run("Create Room", func(t *testing.T) {
		const path = "/rooms/create"

		payload := map[string]interface{}{
			"name":        "B709",
			"description": "The room is right in front of the elevator",
			"capacity":    30,
		}

		resp, err := app.makeRequestAuth(http.MethodPost, path, payload, adminToken)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		defer func() { _ = resp.Body.Close() }()

		var room map[string]map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&room)
		require.NoError(t, err)

		// Check the room id
		roomID := room["room"]["id"].(string)
		assert.NotEmpty(t, roomID)

		_, err = uuid.Parse(roomID)
		assert.NoError(t, err)

		// Check the name, the description and the capacity
		assert.Equal(t, payload["name"], room["room"]["name"])
		assert.Equal(t, payload["description"], room["room"]["description"])
		assert.Equal(t, payload["capacity"], int(room["room"]["capacity"].(float64)))

		// Check the metadata
		assert.NotEmpty(t, room["room"]["created_at"])
	})

	t.Run("List Rooms", func(t *testing.T) {
		const path = "/rooms/list"

		// Try it with 2 types of tokens - admin and user
		resp, err := app.makeRequestAuth(http.MethodGet, path, nil, adminToken)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		defer func() { _ = resp.Body.Close() }()

		var (
			rooms   map[string][]map[string]interface{}
			prevLen int
		)

		err = json.NewDecoder(resp.Body).Decode(&rooms)
		require.NoError(t, err)

		prevLen = len(rooms["rooms"])
		assert.NotZero(t, prevLen)

		// Now make the sane request but with user rights
		resp, err = app.makeRequestAuth(http.MethodGet, path, nil, userToken)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		err = json.NewDecoder(resp.Body).Decode(&rooms)
		require.NoError(t, err)
		assert.NotEmpty(t, rooms["rooms"])

		assert.Equal(t, prevLen, len(rooms["rooms"]))
	})
}

func TestRoom_ValidateAndConflicts(t *testing.T) {
	app := setupE2E(t)
	adminToken := app.getAdminToken(t)
	userToken := app.getUserToken(t)

	t.Run("Create Room - Bad Cases", func(t *testing.T) {
		const path = "/rooms/create"

		type testCase struct {
			name           string
			token          string
			payload        map[string]interface{}
			expectedStatus int
			expectedError  string
		}

		var tests = []testCase{
			{
				name:           "Bad Request - Name Not Specified",
				token:          adminToken,
				payload:        map[string]interface{}{"name": ""},
				expectedStatus: http.StatusBadRequest,
				expectedError:  "invalid input",
			},
			{
				name:  "Bad Request - Empty Description",
				token: adminToken,
				payload: map[string]interface{}{
					"name":        "A002",
					"description": "",
				},
				expectedStatus: http.StatusBadRequest,
				expectedError:  "invalid input",
			},
			{
				name:  "Bad Request - Negative Capacity",
				token: adminToken,
				payload: map[string]interface{}{
					"name":     "A002",
					"capacity": -50,
				},
				expectedStatus: http.StatusBadRequest,
				expectedError:  "invalid input",
			},
			{
				name:           "Unauthorized - Invalid Token",
				token:          "invalid-token",
				payload:        map[string]interface{}{},
				expectedStatus: http.StatusUnauthorized,
				expectedError:  "invalid or expired token",
			},
			{
				name:           "Forbidden - Insufficient Permissions",
				token:          userToken,
				payload:        map[string]interface{}{"name": "A112"},
				expectedStatus: http.StatusForbidden,
				expectedError:  "insufficient permissions",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				resp, err := app.makeRequestAuth(http.MethodPost, path, tt.payload, tt.token)
				require.NoError(t, err)

				defer func() { _ = resp.Body.Close() }()

				assert.Equal(t, tt.expectedStatus, resp.StatusCode)

				var errResp map[string]string
				_ = json.NewDecoder(resp.Body).Decode(&errResp)
				assert.Contains(t, errResp["error"], tt.expectedError)
			})
		}
	})

	t.Run("List Rooms - Bad Cases", func(t *testing.T) {
		const path = "/rooms/list"

		type testCase struct {
			name           string
			token          string
			payload        map[string]interface{}
			expectedStatus int
			expectedError  string
		}

		var tests = []testCase{
			{
				name:           "Unauthorized - Invalid Token",
				token:          "invalid-token",
				payload:        map[string]interface{}{},
				expectedStatus: http.StatusUnauthorized,
				expectedError:  "invalid or expired token",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				resp, err := app.makeRequestAuth(http.MethodGet, path, tt.payload, tt.token)
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
