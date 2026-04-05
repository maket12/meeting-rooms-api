package dto

type CreateRoomInput struct {
	Name        string
	Description *string
	Capacity    *int
}

type CreateRoomOutput struct {
	Room Room
}
