//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchedule_Create(t *testing.T) {
	app := setupE2E(t)
	adminToken := app.getAdminToken(t)
	userToken := app.getUserToken(t)

	/*
		Prepare test data - create 2 rooms
		The first one is for successful scenario and the second one is for failed ones
	*/
	firstRoomID := app.createRoom(t)
	secondRoomID := app.createRoom(t)

	type testCase struct {
		name           string
		token          string
		roomID         string
		payload        map[string]interface{}
		expectedStatus int
		expectedError  string
	}

	var tests = []testCase{
		{
			name:   "Success - Create Schedule",
			token:  adminToken,
			roomID: firstRoomID,
			payload: map[string]interface{}{
				"days_of_week": []int{1, 2, 5},
				"start_time":   "8:30",
				"end_time":     "9:30",
			},
			expectedStatus: http.StatusCreated,
			expectedError:  "",
		},
		{
			name:           "Bad Request - Invalid Identifier",
			token:          adminToken,
			roomID:         "not-an-uuid",
			payload:        map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid identifier format",
		},
		{
			name:           "Bad Request - Invalid Days Of Week (the array is empty)",
			token:          adminToken,
			roomID:         secondRoomID,
			payload:        map[string]interface{}{"days_of_week": []int{}},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid input",
		},
		{
			name:           "Bad Request - Invalid Days Of Week (contains invalid values)",
			token:          adminToken,
			roomID:         secondRoomID,
			payload:        map[string]interface{}{"days_of_week": []int{8, 10, 50}},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid input",
		},
		{
			name:   "Bad Request - Invalid Start Time",
			token:  adminToken,
			roomID: secondRoomID,
			payload: map[string]interface{}{
				"days_of_week": []int{7},
				"start_time":   "80:579",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid input",
		},
		{
			name:   "Bad Request - Invalid End Time",
			token:  adminToken,
			roomID: secondRoomID,
			payload: map[string]interface{}{
				"days_of_week": []int{7},
				"end_time":     "947",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid input",
		},
		{
			name:   "Bad Request - Invalid Working Hours",
			token:  adminToken,
			roomID: secondRoomID,
			payload: map[string]interface{}{
				"days_of_week": []int{7},
				"start_time":   "12:30",
				"end_time":     "9:00",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid input",
		},
		{
			name:   "Unauthorized - Invalid Token",
			token:  "invalid-token",
			roomID: secondRoomID,
			payload: map[string]interface{}{
				"days_of_week": []int{1, 2, 5},
				"start_time":   "8:30",
				"end_time":     "9:30",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid or expired token",
		},
		{
			name:   "Forbidden - Insufficient Permissions",
			token:  userToken,
			roomID: secondRoomID,
			payload: map[string]interface{}{
				"days_of_week": []int{1, 2, 5},
				"start_time":   "8:30",
				"end_time":     "9:30",
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "insufficient permissions",
		},
		{
			name:   "Not Found - Room Doesn't Exist",
			token:  adminToken,
			roomID: uuid.New().String(),
			payload: map[string]interface{}{
				"days_of_week": []int{1, 2, 5},
				"start_time":   "8:30",
				"end_time":     "9:30",
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "room not found",
		},
		{
			name:   "Conflict - Room Already Has Schedule",
			token:  adminToken,
			roomID: firstRoomID,
			payload: map[string]interface{}{
				"days_of_week": []int{1, 2, 5},
				"start_time":   "8:30",
				"end_time":     "9:30",
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "schedule for this room already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := fmt.Sprintf("/rooms/%s/schedule/create", tt.roomID)

			resp, err := app.makeRequestAuth(
				http.MethodPost, path,
				tt.payload, tt.token,
			)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if resp.StatusCode == http.StatusCreated {
				var schedule map[string]map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&schedule)
				require.NoError(t, err)

				// Check the schedule id
				scheduleID := schedule["schedule"]["id"].(string)
				assert.NotEmpty(t, scheduleID)

				_, err = uuid.Parse(scheduleID)
				assert.NoError(t, err)

				// Check the room id, days of week and working hours
				assert.Equal(t, tt.roomID, schedule["schedule"]["room_id"])
				assert.ObjectsAreEqual(
					tt.payload["days_of_week"],
					schedule["schedule"]["days_of_week"],
				)
				assert.Contains(t,
					schedule["schedule"]["start_time"],
					tt.payload["start_time"],
				)
				assert.Contains(t,
					schedule["schedule"]["end_time"],
					tt.payload["end_time"],
				)
			} else {
				var errResp map[string]string
				err = json.NewDecoder(resp.Body).Decode(&errResp)
				require.NoError(t, err)

				assert.Contains(t, errResp["error"], tt.expectedError)
			}
		})
	}
}
