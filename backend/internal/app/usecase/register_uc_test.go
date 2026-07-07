package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/maket12/meeting-rooms-api/internal/app/dto"
	ucerrs "github.com/maket12/meeting-rooms-api/internal/app/errs"
	"github.com/maket12/meeting-rooms-api/internal/app/usecase"
	"github.com/maket12/meeting-rooms-api/internal/domain/model"
	"github.com/maket12/meeting-rooms-api/internal/domain/port/mocks"
	pkgerrs "github.com/maket12/meeting-rooms-api/pkg/errs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegisterUC_Execute(t *testing.T) {
	type adapter struct {
		user     *mocks.MockUserRepository
		password *mocks.MockPasswordHasher
	}

	type testCase struct {
		name          string
		input         dto.RegisterInput
		mockBehaviour func(a adapter)
		expectErr     error
	}

	var tests = []testCase{
		{
			name: "Success",
			input: dto.RegisterInput{
				Email:    "test@avito.ru",
				Password: "password",
				Role:     "user",
			},
			mockBehaviour: func(a adapter) {
				a.password.EXPECT().Hash("password").
					Return("hashed", nil)
				a.user.EXPECT().Create(mock.Anything, mock.AnythingOfType("*model.User")).
					RunAndReturn(func(ctx context.Context, u *model.User) (*model.User, error) {
						return u, nil
					})
			},
			expectErr: nil,
		},
		{
			name: "Failure - hasher error",
			input: dto.RegisterInput{
				Email:    "test@avito.ru",
				Password: "wrong-pass",
				Role:     "user",
			},
			mockBehaviour: func(a adapter) {
				a.password.EXPECT().Hash("wrong-pass").
					Return("", errors.New("failed to hash"))
			},
			expectErr: ucerrs.ErrHashPassword,
		},
		{
			name: "Failure - invalid input",
			input: dto.RegisterInput{
				Email:    "",
				Password: "",
				Role:     "user",
			},
			mockBehaviour: func(a adapter) {
				a.password.EXPECT().Hash("").
					Return("hashed", nil)
			},
			expectErr: ucerrs.ErrInvalidInput,
		},
		{
			name: "Failure - already exists",
			input: dto.RegisterInput{
				Email:    "test@avito.ru",
				Password: "password",
				Role:     "user",
			},
			mockBehaviour: func(a adapter) {
				a.password.EXPECT().Hash("password").
					Return("hashed", nil)
				a.user.EXPECT().Create(mock.Anything, mock.AnythingOfType("*model.User")).
					Return(nil, pkgerrs.ErrObjectAlreadyExists)
			},
			expectErr: ucerrs.ErrUserAlreadyExists,
		},
		{
			name: "Failure - repository error",
			input: dto.RegisterInput{
				Email:    "test@avito.ru",
				Password: "password",
				Role:     "user",
			},
			mockBehaviour: func(a adapter) {
				a.password.EXPECT().Hash("password").
					Return("hashed", nil)
				a.user.EXPECT().Create(mock.Anything, mock.AnythingOfType("*model.User")).
					Return(nil, errors.New("failed to create user"))
			},
			expectErr: ucerrs.ErrCreateUserDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mocks
			userRepo := mocks.NewMockUserRepository(t)
			passHasher := mocks.NewMockPasswordHasher(t)
			tt.mockBehaviour(adapter{user: userRepo, password: passHasher})

			// UC
			uc := usecase.NewRegisterUC(userRepo, passHasher)

			// Call method
			out, err := uc.Execute(context.Background(), tt.input)

			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.input.Email, out.User.Email)
				assert.Equal(t, tt.input.Role, out.User.Role)
				assert.NotEmpty(t, out.User.ID)
			}
		})
	}
}
