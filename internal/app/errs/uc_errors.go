package errs

import "errors"

/*
================ Validation failures ================
*/
var (
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrCannotLogin         = errors.New("account either is blocked or not exists")
	ErrInvalidAccountID    = errors.New("account id is invalid or account with this id not found")
	ErrCannotAssign        = errors.New("account can not be assigned to this role")
	ErrInvalidRefreshToken = errors.New("refresh token is invalid or not found")
	ErrCannotRevoke        = errors.New("refresh token has been already rotated or invalid")
	ErrInvalidAccessToken  = errors.New("access token is invalid")

	ErrInvalidInput = errors.New("invalid input") // for rich models
)

/*
================ Adapter failures ================
*/
var (
	ErrHashPassword         = errors.New("failed to hash password")
	ErrGenerateToken        = errors.New("failed to generate token")
	ErrGenerateRefreshToken = errors.New("failed to generate refresh token")
	ErrPublishEvent         = errors.New("failed to publish event")
)

/*
================ Database failures ================
*/
var (
	ErrCreateUserDB            = errors.New("failed to create user using db")
	ErrUserAlreadyExists       = errors.New("user with given email already exists")
	ErrGetUserByIDDB           = errors.New("failed to get user by id using db")
	ErrGetUserByEmailDB        = errors.New("failed to get user by email using db")
	ErrCreateRoomDB            = errors.New("failed to create room using db")
	ErrListRoomsDB             = errors.New("failed to get a list of rooms using db")
	ErrCreateScheduleDB        = errors.New("failed to create schedule using db")
	ErrCreateSlotsDB           = errors.New("failed to create slots using db")
	ErrListSlotsDB             = errors.New("failed to get a list of slots using db")
	ErrUpdateAccountDB         = errors.New("failed to update account using db")
	ErrGetAccountRoleDB        = errors.New("failed to get account role using db")
	ErrUpdateAccountRoleDB     = errors.New("failed to update account role using db")
	ErrCreateRefreshSessionDB  = errors.New("failed to create refresh session using db")
	ErrGetRefreshSessionByIDDB = errors.New("failed to get refresh session by id using db")
	ErrRevokeRefreshSessionDB  = errors.New("failed to revoke refresh session using db")
)
