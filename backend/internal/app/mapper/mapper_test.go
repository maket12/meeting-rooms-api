package mapper_test

import (
	"backend/internal/app/dto"
	"backend/internal/app/mapper"
	"backend/internal/domain/model"
	"backend/pkg/utils"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// ==================== User Tests ====================

func TestMapDomainToUserDTO(t *testing.T) {
	tests := []struct {
		name      string
		givenUser *model.User
		expected  dto.User
	}{
		{
			name: "admin_user_mapping",
			givenUser: model.RestoreUser(
				uuid.MustParse("12345678-1234-1234-1234-123456789012"),
				"admin@example.com",
				"hashedpassword123",
				model.RoleAdmin,
				time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC),
			),
			expected: dto.User{
				ID:        uuid.MustParse("12345678-1234-1234-1234-123456789012"),
				Email:     "admin@example.com",
				Role:      "admin",
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC),
			},
		},
		{
			name: "regular_user_mapping",
			givenUser: model.RestoreUser(
				uuid.MustParse("87654321-4321-4321-4321-210987654321"),
				"user@example.com",
				"hashedpassword456",
				model.RoleUser,
				time.Date(2024, 2, 20, 14, 25, 30, 0, time.UTC),
			),
			expected: dto.User{
				ID:        uuid.MustParse("87654321-4321-4321-4321-210987654321"),
				Email:     "user@example.com",
				Role:      "user",
				CreatedAt: time.Date(2024, 2, 20, 14, 25, 30, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapper.MapDomainToUserDTO(tt.givenUser)

			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.Email, result.Email)
			assert.Equal(t, tt.expected.Role, result.Role)
			assert.Equal(t, tt.expected.CreatedAt, result.CreatedAt)
		})
	}
}

func TestMapDomainToRegisterDTO(t *testing.T) {
	tests := []struct {
		name         string
		givenUser    *model.User
		expectedUser dto.User
	}{
		{
			name: "register_output_contains_correct_user",
			givenUser: model.RestoreUser(
				uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
				"newuser@example.com",
				"hash",
				model.RoleUser,
				time.Date(2024, 3, 10, 9, 0, 0, 0, time.UTC),
			),
			expectedUser: dto.User{
				ID:        uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
				Email:     "newuser@example.com",
				Role:      "user",
				CreatedAt: time.Date(2024, 3, 10, 9, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapper.MapDomainToRegisterDTO(tt.givenUser)

			assert.Equal(t, tt.expectedUser.ID, result.User.ID)
			assert.Equal(t, tt.expectedUser.Email, result.User.Email)
			assert.Equal(t, tt.expectedUser.Role, result.User.Role)
			assert.Equal(t, tt.expectedUser.CreatedAt, result.User.CreatedAt)
		})
	}
}

// ==================== Room Tests ====================

func TestMapDomainToRoomDTO(t *testing.T) {
	tests := []struct {
		name              string
		givenRoom         *model.Room
		expectedID        uuid.UUID
		expectedName      string
		expectedDesc      *string
		expectedCapacity  *int
		expectedCreatedAt time.Time
	}{
		{
			name: "room_with_all_fields",
			givenRoom: model.RestoreRoom(
				uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
				"Conference Room A",
				utils.VPtr("Large meeting room with projection"),
				utils.VPtr(20),
				time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			),
			expectedID:        uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
			expectedName:      "Conference Room A",
			expectedDesc:      utils.VPtr("Large meeting room with projection"),
			expectedCapacity:  utils.VPtr(20),
			expectedCreatedAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		},
		{
			name: "room_with_nil_description_and_capacity",
			givenRoom: model.RestoreRoom(
				uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc"),
				"Small Room",
				nil,
				nil,
				time.Date(2024, 1, 5, 8, 30, 0, 0, time.UTC),
			),
			expectedID:        uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc"),
			expectedName:      "Small Room",
			expectedDesc:      nil,
			expectedCapacity:  nil,
			expectedCreatedAt: time.Date(2024, 1, 5, 8, 30, 0, 0, time.UTC),
		},
		{
			name: "room_with_only_description",
			givenRoom: model.RestoreRoom(
				uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd"),
				"Meeting Room B",
				utils.VPtr("Board room"),
				nil,
				time.Date(2024, 2, 10, 15, 45, 0, 0, time.UTC),
			),
			expectedID:        uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd"),
			expectedName:      "Meeting Room B",
			expectedDesc:      utils.VPtr("Board room"),
			expectedCapacity:  nil,
			expectedCreatedAt: time.Date(2024, 2, 10, 15, 45, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapper.MapDomainToRoomDTO(tt.givenRoom)

			assert.Equal(t, tt.expectedID, result.ID)
			assert.Equal(t, tt.expectedName, result.Name)
			if tt.expectedDesc == nil {
				assert.Nil(t, result.Description)
			} else {
				assert.NotNil(t, result.Description)
				assert.Equal(t, *tt.expectedDesc, *result.Description)
			}
			if tt.expectedCapacity == nil {
				assert.Nil(t, result.Capacity)
			} else {
				assert.NotNil(t, result.Capacity)
				assert.Equal(t, *tt.expectedCapacity, *result.Capacity)
			}
			assert.Equal(t, tt.expectedCreatedAt, result.CreatedAt)
		})
	}
}

func TestMapDomainToCreateRoomDTO(t *testing.T) {
	tests := []struct {
		name      string
		givenRoom *model.Room
	}{
		{
			name: "create_room_output_structure",
			givenRoom: model.RestoreRoom(
				uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"),
				"New Room",
				utils.VPtr("Description"),
				utils.VPtr(10),
				time.Date(2024, 1, 20, 11, 0, 0, 0, time.UTC),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapper.MapDomainToCreateRoomDTO(tt.givenRoom)

			assert.Equal(t, tt.givenRoom.ID(), result.Room.ID)
			assert.Equal(t, tt.givenRoom.Name(), result.Room.Name)
			assert.NotNil(t, result.Room.Description)
			assert.Equal(t, *tt.givenRoom.Description(), *result.Room.Description)
			assert.NotNil(t, result.Room.Capacity)
			assert.Equal(t, *tt.givenRoom.Capacity(), *result.Room.Capacity)
			assert.Equal(t, tt.givenRoom.CreatedAt(), result.Room.CreatedAt)
		})
	}
}

func TestMapDomainToListRoomsDTO(t *testing.T) {
	tests := []struct {
		name       string
		givenRooms []*model.Room
		expected   int // capacity
	}{
		{
			name: "list_with_multiple_rooms",
			givenRooms: []*model.Room{
				model.RestoreRoom(
					uuid.MustParse("11111111-1111-1111-1111-111111111111"),
					"Room 1",
					utils.VPtr("Desc 1"),
					utils.VPtr(5),
					time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				),
				model.RestoreRoom(
					uuid.MustParse("22222222-2222-2222-2222-222222222222"),
					"Room 2",
					nil,
					utils.VPtr(10),
					time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
				),
				model.RestoreRoom(
					uuid.MustParse("33333333-3333-3333-3333-333333333333"),
					"Room 3",
					utils.VPtr("Desc 3"),
					nil,
					time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
				),
			},
			expected: 3,
		},
		{
			name:       "empty_list",
			givenRooms: []*model.Room{},
			expected:   0,
		},
		{
			name: "single_room",
			givenRooms: []*model.Room{
				model.RestoreRoom(
					uuid.MustParse("44444444-4444-4444-4444-444444444444"),
					"Only Room",
					utils.VPtr("Only one"),
					utils.VPtr(1),
					time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC),
				),
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapper.MapDomainToListRoomsDTO(tt.givenRooms)

			assert.Len(t, result.Rooms, tt.expected)

			for i, domainRoom := range tt.givenRooms {
				assert.Equal(t, domainRoom.ID(), result.Rooms[i].ID)
				assert.Equal(t, domainRoom.Name(), result.Rooms[i].Name)
				if domainRoom.Description() == nil {
					assert.Nil(t, result.Rooms[i].Description)
				} else {
					assert.NotNil(t, result.Rooms[i].Description)
					assert.Equal(t, *domainRoom.Description(), *result.Rooms[i].Description)
				}
				if domainRoom.Capacity() == nil {
					assert.Nil(t, result.Rooms[i].Capacity)
				} else {
					assert.NotNil(t, result.Rooms[i].Capacity)
					assert.Equal(t, *domainRoom.Capacity(), *result.Rooms[i].Capacity)
				}
				assert.Equal(t, domainRoom.CreatedAt(), result.Rooms[i].CreatedAt)
			}
		})
	}
}

// ==================== Schedule Tests ====================

func TestMapDomainToScheduleDTO(t *testing.T) {
	tests := []struct {
		name          string
		givenSchedule *model.Schedule
		expectedDays  []int
		expectedStart string
		expectedEnd   string
	}{
		{
			name: "weekday_schedule_9_to_18",
			givenSchedule: model.RestoreSchedule(
				uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff"),
				uuid.MustParse("55555555-5555-5555-5555-555555555555"),
				[]int{1, 2, 3, 4, 5},
				model.RestoreDayTime(9*60),  // 09:00
				model.RestoreDayTime(18*60), // 18:00
			),
			expectedDays:  []int{1, 2, 3, 4, 5},
			expectedStart: "09:00",
			expectedEnd:   "18:00",
		},
		{
			name: "weekend_schedule_10_to_16",
			givenSchedule: model.RestoreSchedule(
				uuid.MustParse("10101010-1010-1010-1010-101010101010"),
				uuid.MustParse("66666666-6666-6666-6666-666666666666"),
				[]int{6, 7},
				model.RestoreDayTime(10*60), // 10:00
				model.RestoreDayTime(16*60), // 16:00
			),
			expectedDays:  []int{6, 7},
			expectedStart: "10:00",
			expectedEnd:   "16:00",
		},
		{
			name: "single_day_with_half_hours",
			givenSchedule: model.RestoreSchedule(
				uuid.MustParse("20202020-2020-2020-2020-202020202020"),
				uuid.MustParse("77777777-7777-7777-7777-777777777777"),
				[]int{3},
				model.RestoreDayTime(8*60+30),  // 08:30
				model.RestoreDayTime(17*60+30), // 17:30
			),
			expectedDays:  []int{3},
			expectedStart: "08:30",
			expectedEnd:   "17:30",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapper.MapDomainToScheduleDTO(tt.givenSchedule)

			assert.Equal(t, tt.givenSchedule.ID(), result.ID)
			assert.Equal(t, tt.givenSchedule.RoomID(), result.RoomID)
			assert.Equal(t, tt.expectedDays, result.DaysOfWeek)
			assert.Equal(t, tt.expectedStart, result.StartTime)
			assert.Equal(t, tt.expectedEnd, result.EndTime)
		})
	}
}

func TestMapDomainToCreateScheduleDTO(t *testing.T) {
	tests := []struct {
		name          string
		givenSchedule *model.Schedule
	}{
		{
			name: "create_schedule_output_wraps_schedule",
			givenSchedule: model.RestoreSchedule(
				uuid.MustParse("30303030-3030-3030-3030-303030303030"),
				uuid.MustParse("88888888-8888-8888-8888-888888888888"),
				[]int{1, 2, 3},
				model.RestoreDayTime(9*60),
				model.RestoreDayTime(17*60),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapper.MapDomainToCreateScheduleDTO(tt.givenSchedule)

			assert.Equal(t, tt.givenSchedule.ID(), result.Schedule.ID)
			assert.Equal(t, tt.givenSchedule.RoomID(), result.Schedule.RoomID)
			assert.Equal(t, tt.givenSchedule.DaysOfWeek(), result.Schedule.DaysOfWeek)
			assert.Equal(t, tt.givenSchedule.StartTime().String(), result.Schedule.StartTime)
			assert.Equal(t, tt.givenSchedule.EndTime().String(), result.Schedule.EndTime)
		})
	}
}

// ==================== Slot Tests ====================

func TestMapDomainToSlotDTO(t *testing.T) {
	tests := []struct {
		name      string
		givenSlot *model.Slot
	}{
		{
			name: "slot_mapping_preserves_all_fields",
			givenSlot: model.RestoreSlot(
				uuid.MustParse("40404040-4040-4040-4040-404040404040"),
				uuid.MustParse("99999999-9999-9999-9999-999999999999"),
				time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC),
				time.Date(2024, 1, 15, 9, 30, 0, 0, time.UTC),
			),
		},
		{
			name: "slot_with_different_times",
			givenSlot: model.RestoreSlot(
				uuid.MustParse("50505050-5050-5050-5050-505050505050"),
				uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
				time.Date(2024, 2, 20, 14, 30, 0, 0, time.UTC),
				time.Date(2024, 2, 20, 15, 0, 0, 0, time.UTC),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapper.MapDomainToSlotDTO(tt.givenSlot)

			assert.Equal(t, tt.givenSlot.ID(), result.ID)
			assert.Equal(t, tt.givenSlot.RoomID(), result.RoomID)
			assert.Equal(t, tt.givenSlot.Start(), result.Start)
			assert.Equal(t, tt.givenSlot.End(), result.End)
		})
	}
}

func TestMapDomainToListSlotsDTO(t *testing.T) {
	tests := []struct {
		name       string
		givenSlots []*model.Slot
		expected   int
	}{
		{
			name: "multiple_slots",
			givenSlots: []*model.Slot{
				model.RestoreSlot(
					uuid.MustParse("60606060-6060-6060-6060-606060606060"),
					uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
					time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC),
					time.Date(2024, 1, 15, 9, 30, 0, 0, time.UTC),
				),
				model.RestoreSlot(
					uuid.MustParse("61616161-6161-6161-6161-616161616161"),
					uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
					time.Date(2024, 1, 15, 9, 30, 0, 0, time.UTC),
					time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
				),
				model.RestoreSlot(
					uuid.MustParse("62626262-6262-6262-6262-626262626262"),
					uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
					time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
					time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				),
			},
			expected: 3,
		},
		{
			name:       "empty_slots_list",
			givenSlots: []*model.Slot{},
			expected:   0,
		},
		{
			name: "single_slot",
			givenSlots: []*model.Slot{
				model.RestoreSlot(
					uuid.MustParse("63636363-6363-6363-6363-636363636363"),
					uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc"),
					time.Date(2024, 1, 16, 11, 0, 0, 0, time.UTC),
					time.Date(2024, 1, 16, 11, 30, 0, 0, time.UTC),
				),
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapper.MapDomainToListSlotsDTO(tt.givenSlots)

			assert.Len(t, result.Slots, tt.expected)

			for i, domainSlot := range tt.givenSlots {
				assert.Equal(t, domainSlot.ID(), result.Slots[i].ID)
				assert.Equal(t, domainSlot.RoomID(), result.Slots[i].RoomID)
				assert.Equal(t, domainSlot.Start(), result.Slots[i].Start)
				assert.Equal(t, domainSlot.End(), result.Slots[i].End)
			}
		})
	}
}

// ==================== Booking Tests ====================

func TestMapDomainToBookingDTO(t *testing.T) {
	tests := []struct {
		name         string
		givenBooking *model.Booking
	}{
		{
			name: "active_booking_with_conference_link",
			givenBooking: model.RestoreBooking(
				uuid.MustParse("70707070-7070-7070-7070-707070707070"),
				uuid.MustParse("71717171-7171-7171-7171-717171717171"),
				uuid.MustParse("72727272-7272-7272-7272-727272727272"),
				model.BookingActive,
				utils.VPtr("https://meet.google.com/abc-defg-hij"),
				time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC),
			),
		},
		{
			name: "cancelled_booking_without_link",
			givenBooking: model.RestoreBooking(
				uuid.MustParse("73737373-7373-7373-7373-737373737373"),
				uuid.MustParse("74747474-7474-7474-7474-747474747474"),
				uuid.MustParse("75757575-7575-7575-7575-757575757575"),
				model.BookingCancelled,
				nil,
				time.Date(2024, 1, 10, 14, 30, 0, 0, time.UTC),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapper.MapDomainToBookingDTO(tt.givenBooking)

			assert.Equal(t, tt.givenBooking.ID(), result.ID)
			assert.Equal(t, tt.givenBooking.SlotID(), result.SlotID)
			assert.Equal(t, tt.givenBooking.UserID(), result.UserID)
			assert.Equal(t, tt.givenBooking.Status().String(), result.Status)
			assert.Equal(t, tt.givenBooking.ConferenceLink(), result.ConferenceLink)
			assert.Equal(t, tt.givenBooking.CreatedAt(), result.CreatedAt)
		})
	}
}

func TestMapDomainToCreateBookingDTO(t *testing.T) {
	tests := []struct {
		name         string
		givenBooking *model.Booking
	}{
		{
			name: "create_booking_output_wraps_booking",
			givenBooking: model.RestoreBooking(
				uuid.MustParse("76767676-7676-7676-7676-767676767676"),
				uuid.MustParse("77777777-7777-7777-7777-777777777777"),
				uuid.MustParse("78787878-7878-7878-7878-787878787878"),
				model.BookingActive,
				utils.VPtr("https://example.com/meeting"),
				time.Date(2024, 1, 20, 10, 0, 0, 0, time.UTC),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapper.MapDomainToCreateBookingDTO(tt.givenBooking)

			assert.Equal(t, tt.givenBooking.ID(), result.Booking.ID)
			assert.Equal(t, tt.givenBooking.SlotID(), result.Booking.SlotID)
			assert.Equal(t, tt.givenBooking.UserID(), result.Booking.UserID)
			assert.Equal(t, tt.givenBooking.Status().String(), result.Booking.Status)
			assert.Equal(t, tt.givenBooking.ConferenceLink(), result.Booking.ConferenceLink)
			assert.Equal(t, tt.givenBooking.CreatedAt(), result.Booking.CreatedAt)
		})
	}
}

func TestMapDomainToListBookings(t *testing.T) {
	tests := []struct {
		name          string
		givenBookings []*model.Booking
		expected      int
	}{
		{
			name: "multiple_bookings",
			givenBookings: []*model.Booking{
				model.RestoreBooking(
					uuid.MustParse("79797979-7979-7979-7979-797979797979"),
					uuid.MustParse("7a7a7a7a-7a7a-7a7a-7a7a-7a7a7a7a7a7a"),
					uuid.MustParse("7b7b7b7b-7b7b-7b7b-7b7b-7b7b7b7b7b7b"),
					model.BookingActive,
					utils.VPtr("link1"),
					time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC),
				),
				model.RestoreBooking(
					uuid.MustParse("7c7c7c7c-7c7c-7c7c-7c7c-7c7c7c7c7c7c"),
					uuid.MustParse("7d7d7d7d-7d7d-7d7d-7d7d-7d7d7d7d7d7d"),
					uuid.MustParse("7e7e7e7e-7e7e-7e7e-7e7e-7e7e7e7e7e7e"),
					model.BookingCancelled,
					nil,
					time.Date(2024, 1, 20, 14, 30, 0, 0, time.UTC),
				),
			},
			expected: 2,
		},
		{
			name:          "empty_bookings_list",
			givenBookings: []*model.Booking{},
			expected:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapper.MapDomainToListBookings(tt.givenBookings)

			assert.Len(t, result.Bookings, tt.expected)

			for i, domainBooking := range tt.givenBookings {
				assert.Equal(t, domainBooking.ID(), result.Bookings[i].ID)
				assert.Equal(t, domainBooking.SlotID(), result.Bookings[i].SlotID)
				assert.Equal(t, domainBooking.UserID(), result.Bookings[i].UserID)
				assert.Equal(t, domainBooking.Status().String(), result.Bookings[i].Status)
				assert.Equal(t, domainBooking.ConferenceLink(), result.Bookings[i].ConferenceLink)
				assert.Equal(t, domainBooking.CreatedAt(), result.Bookings[i].CreatedAt)
			}
		})
	}
}

func TestMapDomainToListMyBookings(t *testing.T) {
	tests := []struct {
		name          string
		givenBookings []*model.Booking
		expected      int
	}{
		{
			name: "my_bookings_with_user_specific_data",
			givenBookings: []*model.Booking{
				model.RestoreBooking(
					uuid.MustParse("7f7f7f7f-7f7f-7f7f-7f7f-7f7f7f7f7f7f"),
					uuid.MustParse("80808080-8080-8080-8080-808080808080"),
					uuid.MustParse("81818181-8181-8181-8181-818181818181"),
					model.BookingActive,
					utils.VPtr("link1"),
					time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC),
				),
			},
			expected: 1,
		},
		{
			name:          "empty_my_bookings",
			givenBookings: []*model.Booking{},
			expected:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapper.MapDomainToListMyBookings(tt.givenBookings)

			assert.Len(t, result.Bookings, tt.expected)

			for i, domainBooking := range tt.givenBookings {
				assert.Equal(t, domainBooking.ID(), result.Bookings[i].ID)
				assert.Equal(t, domainBooking.SlotID(), result.Bookings[i].SlotID)
				assert.Equal(t, domainBooking.UserID(), result.Bookings[i].UserID)
				assert.Equal(t, domainBooking.Status().String(), result.Bookings[i].Status)
				assert.Equal(t, domainBooking.ConferenceLink(), result.Bookings[i].ConferenceLink)
				assert.Equal(t, domainBooking.CreatedAt(), result.Bookings[i].CreatedAt)
			}
		})
	}
}
