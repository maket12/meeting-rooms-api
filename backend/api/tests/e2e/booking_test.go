package e2e

import (
	"backend/pkg/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBooking_LifeCycle(t *testing.T) {
	app := setupE2E(t)
	adminToken := app.getAdminToken(t)
	userToken := app.getUserToken(t)

	/*
		Prepare test data:
		1) Create a room.
		2) Create a schedule for the room.
		3) Get slots for the room
	*/
	roomID := app.createRoom(t)
	app.createSchedule(t, roomID)
	slots := app.getSlots(t, roomID)

	var bookingID string

	t.Run("Create Booking", func(t *testing.T) {
		const path = "/bookings/create"

		payload := map[string]interface{}{
			"slot_id":                slots[0]["id"],
			"create_conference_link": false,
		}

		resp, err := app.makeRequestAuth(http.MethodPost, path, payload, userToken)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		defer func() { _ = resp.Body.Close() }()

		var booking map[string]map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&booking)
		require.NoError(t, err)

		// Check the booking id
		id := booking["booking"]["id"].(string)
		assert.NotEmpty(t, id)

		_, err = uuid.Parse(id)
		assert.NoError(t, err)

		bookingID = id

		// Check the slot id
		slotID := booking["booking"]["slot_id"].(string)
		assert.NotEmpty(t, slotID)

		_, err = uuid.Parse(slotID)
		assert.NoError(t, err)

		assert.Equal(t, slots[0]["id"], slotID)

		// Check the user id
		userID := booking["booking"]["user_id"].(string)
		assert.NotEmpty(t, userID)

		_, err = uuid.Parse(userID)
		assert.NoError(t, err)

		// Check the status and the conference link
		assert.NotEmpty(t, booking["booking"]["status"])
		assert.True(t, booking["booking"]["status"] == "active")
		assert.Empty(t, booking["booking"]["conference_link"])

		// Check the metadata
		assert.NotEmpty(t, booking["booking"]["created_at"])
	})

	t.Run("List My Bookings", func(t *testing.T) {
		const path = "/bookings/my"

		resp, err := app.makeRequestAuth(http.MethodGet, path, nil, userToken)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		defer func() { _ = resp.Body.Close() }()

		var bookings map[string][]map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&bookings)
		require.NoError(t, err)

		assert.NotEmpty(t, bookings["bookings"])

		// Ensure that there are only active bookings
		assert.True(t, bookings["bookings"][0]["status"] == "active")
	})

	t.Run("List All Bookings", func(t *testing.T) {
		const path = "/bookings/list"

		resp, err := app.makeRequestAuth(http.MethodGet, path, nil, adminToken)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		defer func() { _ = resp.Body.Close() }()

		var respMapped map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&respMapped)
		require.NoError(t, err)

		// Check bookings and pagination
		bookings, ok := respMapped["bookings"].([]interface{})
		require.True(t, ok, "bookings must be an array")

		pagination, ok := respMapped["pagination"].(map[string]interface{})
		require.True(t, ok, "pagination must be an object")

		assert.NotEmpty(t, bookings)

		total, ok := pagination["total"].(float64)
		require.True(t, ok, "total must be a number")

		assert.Equal(t, len(bookings), int(total))
	})

	t.Run("Cancel Booking", func(t *testing.T) {
		path := fmt.Sprintf("/bookings/%s/cancel", bookingID)

		resp, err := app.makeRequestAuth(http.MethodPost, path, nil, userToken)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		defer func() { _ = resp.Body.Close() }()

		var booking map[string]map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&booking)
		require.NoError(t, err)

		// Check the status
		assert.NotEmpty(t, booking["booking"]["status"])
		assert.True(t, booking["booking"]["status"] == "cancelled")

		/*
			Check the idempotency of the operation.
			Try to cancel it again (must be OK).
		*/
		resp, err = app.makeRequestAuth(http.MethodPost, path, nil, userToken)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestBooking_ValidateAndConflicts(t *testing.T) {
	app := setupE2E(t)
	adminToken := app.getAdminToken(t)
	userToken := app.getUserToken(t)

	/*
		Prepare test data:
		1) Create the new user
		2) Create a room.
		3) Create a schedule for the room.
		4) Get slots for the room.
		5) Book one of the slots.
	*/
	_, anotherUserToken := app.createUser(t, utils.VPtr("new@gmail.com"), nil, nil)

	roomID := app.createRoom(t)
	app.createSchedule(t, roomID)
	slots := app.getSlots(t, roomID)

	bookedSlotID := slots[1]["id"].(string)
	bookingID := app.createBooking(t, bookedSlotID)

	t.Run("Create Booking - Bad Cases", func(t *testing.T) {
		const path = "/bookings/create"

		type testCase struct {
			name           string
			token          string
			payload        map[string]interface{}
			expectedStatus int
			expectedError  string
		}

		var tests = []testCase{
			{
				name:  "Bad Request - Invalid Slot ID",
				token: userToken,
				payload: map[string]interface{}{
					"slot_id":                "not-a-uuid",
					"create_conference_link": false,
				},
				expectedStatus: http.StatusBadRequest,
				expectedError:  "invalid identifier format",
			},
			{
				name:           "Unauthorized - Invalid Token",
				token:          "invalid-token",
				payload:        map[string]interface{}{},
				expectedStatus: http.StatusUnauthorized,
				expectedError:  "invalid or expired token",
			},
			{
				name:  "Forbidden - Insufficient Permissions",
				token: adminToken,
				payload: map[string]interface{}{
					"slot_id":                slots[0]["id"],
					"create_conference_link": false,
				},
				expectedStatus: http.StatusForbidden,
				expectedError:  "insufficient permissions",
			},
			{
				name:  "Not Found - Slot Doesn't Exist",
				token: userToken,
				payload: map[string]interface{}{
					"slot_id":                uuid.New().String(),
					"create_conference_link": false,
				},
				expectedStatus: http.StatusNotFound,
				expectedError:  "slot not found",
			},
			{
				name:  "Conflict - Slot Is Already Booked",
				token: userToken,
				payload: map[string]interface{}{
					"slot_id":                bookedSlotID,
					"create_conference_link": false,
				},
				expectedStatus: http.StatusConflict,
				expectedError:  "booking for this slot already exists",
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

	t.Run("List My Bookings - Bad Cases", func(t *testing.T) {
		const path = "/bookings/my"

		type testCase struct {
			name           string
			token          string
			expectedStatus int
			expectedError  string
		}

		var tests = []testCase{
			{
				name:           "Unauthorized - Invalid Token",
				token:          "invalid-token",
				expectedStatus: http.StatusUnauthorized,
				expectedError:  "invalid or expired token",
			},
			{
				name:           "Forbidden - Insufficient Permissions",
				token:          adminToken,
				expectedStatus: http.StatusForbidden,
				expectedError:  "insufficient permissions",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				resp, err := app.makeRequestAuth(http.MethodGet, path, nil, tt.token)
				require.NoError(t, err)

				defer func() { _ = resp.Body.Close() }()

				assert.Equal(t, tt.expectedStatus, resp.StatusCode)

				var errResp map[string]string
				_ = json.NewDecoder(resp.Body).Decode(&errResp)
				assert.Contains(t, errResp["error"], tt.expectedError)
			})
		}
	})

	t.Run("List All Bookings - Bad Cases", func(t *testing.T) {
		type testCase struct {
			name           string
			token          string
			page           int
			pageSize       int
			expectedStatus int
			expectedError  string
		}

		var tests = []testCase{
			{
				name:           "Bad Request - Invalid Page",
				token:          adminToken,
				page:           -10,
				expectedStatus: http.StatusBadRequest,
				expectedError:  "invalid input",
			},
			{
				name:           "Bad Request - Invalid Page Size",
				token:          adminToken,
				page:           1,
				pageSize:       -10,
				expectedStatus: http.StatusBadRequest,
				expectedError:  "invalid input",
			},
			{
				name:           "Unauthorized - Invalid Token",
				token:          "invalid-token",
				expectedStatus: http.StatusUnauthorized,
				expectedError:  "invalid or expired token",
			},
			{
				name:           "Forbidden - Insufficient Permissions",
				token:          userToken,
				expectedStatus: http.StatusForbidden,
				expectedError:  "insufficient permissions",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var builder strings.Builder

				builder.Grow(40)
				builder.WriteString("/bookings/list")

				if tt.page != 0 {
					builder.WriteString(fmt.Sprintf(
						"?page=%d", tt.page,
					))
				}
				if tt.pageSize != 0 {
					builder.WriteString(fmt.Sprintf(
						"&page_size=%d", tt.pageSize,
					))
				}

				resp, err := app.makeRequestAuth(
					http.MethodGet, builder.String(),
					nil, tt.token,
				)
				require.NoError(t, err)

				defer func() { _ = resp.Body.Close() }()

				assert.Equal(t, tt.expectedStatus, resp.StatusCode)

				var errResp map[string]string
				_ = json.NewDecoder(resp.Body).Decode(&errResp)
				assert.Contains(t, errResp["error"], tt.expectedError)
			})
		}
	})

	t.Run("Cancel Booking - Bad Cases", func(t *testing.T) {
		type testCase struct {
			name           string
			token          string
			bookingID      string
			expectedStatus int
			expectedError  string
		}

		var tests = []testCase{
			{
				name:           "Bad Request - Invalid Booking ID",
				token:          userToken,
				bookingID:      "not-a-uuid",
				expectedStatus: http.StatusBadRequest,
				expectedError:  "invalid identifier format",
			},
			{
				name:           "Unauthorized - Invalid Token",
				token:          "invalid-token",
				bookingID:      bookingID,
				expectedStatus: http.StatusUnauthorized,
				expectedError:  "invalid or expired token",
			},
			{
				name:           "Forbidden - Not a User",
				token:          adminToken,
				bookingID:      bookingID,
				expectedStatus: http.StatusForbidden,
				expectedError:  "insufficient permissions",
			},
			{
				name:           "Forbidden - Not an Owner",
				token:          anotherUserToken,
				bookingID:      bookingID,
				expectedStatus: http.StatusForbidden,
				expectedError:  "insufficient permissions",
			},
			{
				name:           "Not Found - Booking Doesn't Exist",
				token:          userToken,
				bookingID:      uuid.New().String(),
				expectedStatus: http.StatusNotFound,
				expectedError:  "booking not found",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				path := fmt.Sprintf("/bookings/%s/cancel", tt.bookingID)

				resp, err := app.makeRequestAuth(http.MethodPost, path, nil, tt.token)
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
