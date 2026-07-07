package usecase

import (
	"context"
	"errors"
	"github.com/maket12/meeting-rooms-api/internal/app/dto"
	ucerrs "github.com/maket12/meeting-rooms-api/internal/app/errs"
	"github.com/maket12/meeting-rooms-api/internal/app/mapper"
	"github.com/maket12/meeting-rooms-api/internal/domain/model"
	"github.com/maket12/meeting-rooms-api/internal/domain/port"
	pkgerrs "github.com/maket12/meeting-rooms-api/pkg/errs"
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
