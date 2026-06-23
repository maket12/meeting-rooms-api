package usecase

import (
	"backend/internal/app/dto"
	"backend/internal/app/errs"
	"backend/internal/app/mapper"
	"backend/internal/domain/model"
	port2 "backend/internal/domain/port"
	pkgerrs "backend/pkg/errs"
	"context"
	"errors"
)

type RegisterUC struct {
	user     port2.UserRepository
	password port2.PasswordHasher
}

func NewRegisterUC(
	user port2.UserRepository,
	password port2.PasswordHasher,
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
		return dto.RegisterOutput{}, errs.Wrap(
			errs.ErrHashPassword, err,
		)
	}

	// Creating rich-models with validation
	user, err := model.NewUser(in.Email, hashedPassword, model.UserRole(in.Role))
	if err != nil {
		return dto.RegisterOutput{}, errs.Wrap(
			errs.ErrInvalidInput, err,
		)
	}

	// Save it into database
	user, err = uc.user.Create(ctx, user)
	if err != nil {
		if errors.Is(err, pkgerrs.ErrObjectAlreadyExists) {
			return dto.RegisterOutput{}, errs.ErrUserAlreadyExists
		}
		return dto.RegisterOutput{}, errs.Wrap(
			errs.ErrCreateUserDB, err,
		)
	}

	return mapper.MapDomainToRegisterDTO(user), nil
}
