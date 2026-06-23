package dto

type RegisterInput struct {
	Email    string
	Password string
	Role     string
}

type RegisterOutput struct {
	User User
}
