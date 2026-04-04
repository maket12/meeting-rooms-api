package dto

type LoginInput struct {
	Email    string
	Password string
}

type LoginOutput struct {
	Token string
}
