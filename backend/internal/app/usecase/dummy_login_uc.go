package usecase

import (
	"context"
	"errors"

	"github.com/maket12/meeting-rooms-api/internal/app/dto"
	ucerrs "github.com/maket12/meeting-rooms-api/internal/app/errs"
	"github.com/maket12/meeting-rooms-api/internal/domain/port"
	pkgerrs "github.com/maket12/meeting-rooms-api/pkg/errs"

	"github.com/google/uuid"
)

type DummyLoginUC struct {
	user    port.UserRepository
	jwtGen  port.TokenGenerator
	adminID uuid.UUID
	userID  uuid.UUID
}

func NewDummyLoginUC(
	user port.UserRepository,
	jwtGen port.TokenGenerator,
	adminID uuid.UUID,
	userID uuid.UUID,
) *DummyLoginUC {
	return &DummyLoginUC{
		user:    user,
		jwtGen:  jwtGen,
		adminID: adminID,
		userID:  userID,
	}
}

func (uc *DummyLoginUC) Execute(ctx context.Context, in dto.DummyLoginInput) (dto.DummyLoginOutput, error) {
	var uID uuid.UUID

	// Input validation
	switch in.Role {
	case "admin":
		uID = uc.adminID
	case "user":
		uID = uc.userID
	default:
		return dto.DummyLoginOutput{}, ucerrs.Wrap(
			ucerrs.ErrInvalidInput, errors.New("invalid role"),
		)
	}

	// Get the dummy user
	user, err := uc.user.GetByID(ctx, uID)
	if err != nil {
		if errors.Is(err, pkgerrs.ErrObjectNotFound) {
			return dto.DummyLoginOutput{}, ucerrs.ErrUserNotFound
		}
		return dto.DummyLoginOutput{}, ucerrs.Wrap(
			ucerrs.ErrGetUserByIDDB, err,
		)
	}

	// Generate the token for the gotten user
	token, err := uc.jwtGen.Generate(user.ID(), user.Role().String())
	if err != nil {
		return dto.DummyLoginOutput{}, ucerrs.Wrap(
			ucerrs.ErrGenerateToken, err,
		)
	}

	return dto.DummyLoginOutput{Token: token}, nil
}
