package usecase

import (
	"backend/internal/app/dto"
	ucerrs "backend/internal/app/errs"
	"backend/internal/app/mapper"
	"backend/internal/domain/model"
	"backend/internal/domain/port"
	pkgerrs "backend/pkg/errs"
	"context"
	"errors"
)

type RegisterUC struct {
	user     port.UserRepository
	password port.PasswordHasher
}

func NewRegisterUC(
	user port.UserRepository,
	password port.PasswordHasher,
) *RegisterUC {
	return &RegisterUC{
		user:     user,
		password: password,
	}
}

func (uc *RegisterUC) Execute(ctx context.Context, in dto.RegisterInput) (dto.RegisterOutput, error) {
	// Hashing the password
	hashedPassword, err := uc.password.Hash(in.Password)
	if err != nil {
		return dto.RegisterOutput{}, ucerrs.Wrap(
			ucerrs.ErrHashPassword, err,
		)
	}

	// Creating rich-models with validation
	user, err := model.NewUser(in.Email, hashedPassword, in.Role)
	if err != nil {
		return dto.RegisterOutput{}, ucerrs.Wrap(
			ucerrs.ErrInvalidInput, err,
		)
	}

	// Save it into database
	user, err = uc.user.Create(ctx, user)
	if err != nil {
		if errors.Is(err, pkgerrs.ErrObjectAlreadyExists) {
			return dto.RegisterOutput{}, ucerrs.ErrUserAlreadyExists
		}
		return dto.RegisterOutput{}, ucerrs.Wrap(
			ucerrs.ErrCreateUserDB, err,
		)
	}

	return mapper.MapDomainToRegisterDTO(user), nil
}
