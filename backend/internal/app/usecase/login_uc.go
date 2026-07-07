package usecase

import (
	"context"
	"errors"
	"github.com/maket12/meeting-rooms-api/internal/app/dto"
	ucerrs "github.com/maket12/meeting-rooms-api/internal/app/errs"
	"github.com/maket12/meeting-rooms-api/internal/domain/port"
	pkgerrs "github.com/maket12/meeting-rooms-api/pkg/errs"
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
			return dto.LoginOutput{}, ucerrs.ErrInvalidCredentials
		}
		return dto.LoginOutput{}, ucerrs.Wrap(ucerrs.ErrGetUserByEmailDB, err)
	}

	// Validate password
	if !uc.password.Compare(user.PasswordHash(), in.Password) {
		return dto.LoginOutput{}, ucerrs.ErrInvalidCredentials
	}

	// Generate the JWT token
	token, err := uc.token.Generate(user.ID(), user.Role().String())
	if err != nil {
		return dto.LoginOutput{}, ucerrs.Wrap(ucerrs.ErrGenerateToken, err)
	}

	return dto.LoginOutput{Token: token}, nil
}
