package errs

import "errors"

/*
================ Validation failures ================
*/
var (
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrCannotCreateBooking = errors.New("specified slot is in the past")
	ErrCannotCancelBooking = errors.New("booking can be cancelled only by owner")

	ErrInvalidInput = errors.New("invalid input") // for rich models
)

/*
================ Adapter failures ================
*/
var (
	ErrHashPassword  = errors.New("failed to hash password")
	ErrGenerateToken = errors.New("failed to generate token")
	ErrCreateMeeting = errors.New("failed to create conference link")
)

/*
================ Database failures ================
*/
var (
	ErrCreateUserDB          = errors.New("failed to create user using db")
	ErrUserAlreadyExists     = errors.New("user with given email already exists")
	ErrGetUserByIDDB         = errors.New("failed to get user by id using db")
	ErrGetUserByEmailDB      = errors.New("failed to get user by email using db")
	ErrCreateRoomDB          = errors.New("failed to create room using db")
	ErrGetRoomDB             = errors.New("failed to get room using db")
	ErrListRoomsDB           = errors.New("failed to get a list of rooms using db")
	ErrCreateScheduleDB      = errors.New("failed to create schedule using db")
	ErrGetScheduleDB         = errors.New("failed to get schedule using db")
	ErrCreateSlotsDB         = errors.New("failed to create slots using db")
	ErrGetSlotDB             = errors.New("failed to get slot using db")
	ErrListSlotsDB           = errors.New("failed to get a list of slots using db")
	ErrCreateBookingDB       = errors.New("failed to create booking using db")
	ErrGetBookingDB          = errors.New("failed to get booking using db")
	ErrUpdateBookingStatusDB = errors.New("failed to update booking status using db")
	ErrListBookingsDB        = errors.New("failed to get a list of bookings using db")
	ErrListMyBookingsDB      = errors.New("failed to get a list of bookings by user id using db")

	ErrRoomNotFound     = errors.New("room not found")
	ErrScheduleNotFound = errors.New("schedule not found")
	ErrSlotNotFound     = errors.New("slot not found")
	ErrBookingNotFound  = errors.New("booking not found")
)
