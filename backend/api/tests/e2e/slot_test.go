//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSlot_List(t *testing.T) {
	app := setupE2E(t)
	token := app.getUserToken(t)

	/*
		Prepare test data:
		1) Create a room.
		2) Create a schedule for this room.
		3) Specify the test date
	*/
	roomID := app.createRoom(t)
	app.createSchedule(t, roomID)
	testDate := time.Now().Add(24 * time.Hour).Format(time.DateOnly)

	type testCase struct {
		name           string
		token          string
		roomID         string
		date           string
		expectedStatus int
		expectedError  string
	}

	var tests = []testCase{
		{
			name:           "Success - Get Slots",
			token:          token,
			roomID:         roomID,
			date:           testDate,
			expectedStatus: http.StatusOK,
			expectedError:  "",
		},
		{
			name:           "Bad Request - Invalid Identifier",
			token:          token,
			roomID:         "not-an-uuid",
			date:           testDate,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid identifier format",
		},
		{
			name:           "Bad Request - Invalid Date",
			token:          token,
			roomID:         roomID,
			date:           "not-a-date-at-all",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid input",
		},
		{
			name:           "Unauthorized - Invalid Token",
			token:          "invalid-token",
			roomID:         roomID,
			date:           testDate,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid or expired token",
		},
		{
			name:           "Not Found - Room Doesn't Exist",
			token:          token,
			roomID:         uuid.New().String(),
			date:           testDate,
			expectedStatus: http.StatusNotFound,
			expectedError:  "room not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := fmt.Sprintf("/rooms/%s/slots/list?date=%s",
				tt.roomID, tt.date,
			)

			resp, err := app.makeRequestAuth(http.MethodGet, path, nil, tt.token)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if resp.StatusCode == http.StatusOK {
				var slots map[string][]map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&slots)
				require.NoError(t, err)

				assert.NotZero(t, len(slots["slots"]))
			} else {
				var errResp map[string]string
				err = json.NewDecoder(resp.Body).Decode(&errResp)
				require.NoError(t, err)

				assert.Contains(t, errResp["error"], tt.expectedError)
			}
		})
	}
}
