package usecase

import (
	"backend/internal/app/dto"
	"backend/internal/app/errs"
	"backend/internal/domain/port"
	pkgerrs "backend/pkg/errs"
	"context"
	"errors"
)

type LoginUC struct {
	user     port.UserRepository
	password port.PasswordHasher
	token    port.TokenGenerator
}

func NewLoginUC(user port.UserRepository,
	password port.PasswordHasher,
	token port.TokenGenerator,
) *LoginUC {
	return &LoginUC{
		user:     user,
		password: password,
		token:    token,
	}
}

func (uc *LoginUC) Execute(ctx context.Context, in dto.LoginInput) (dto.LoginOutput, error) {
	// Find the user
	user, err := uc.user.GetByEmail(ctx, in.Email)
	if err != nil {
		if errors.Is(err, pkgerrs.ErrObjectNotFound) {
			return dto.LoginOutput{}, errs.ErrInvalidCredentials
		}
		return dto.LoginOutput{}, errs.Wrap(
			errs.ErrGetUserByEmailDB, err,
		)
	}

	// Validate password
	if !uc.password.Compare(user.PasswordHash(), in.Password) {
		return dto.LoginOutput{}, errs.ErrInvalidCredentials
	}

	// Generate the JWT token
	token, err := uc.token.Generate(user.ID(), user.Role().String())
	if err != nil {
		return dto.LoginOutput{}, errs.Wrap(
			errs.ErrGenerateToken,
			err,
		)
	}

	return dto.LoginOutput{Token: token}, nil
}
