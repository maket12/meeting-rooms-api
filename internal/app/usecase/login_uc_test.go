package usecase_test

import (
	"MeetingRoomsAPI/internal/app/dto"
	ucerrs "MeetingRoomsAPI/internal/app/errs"
	"MeetingRoomsAPI/internal/app/usecase"
	"MeetingRoomsAPI/internal/domain/model"
	"MeetingRoomsAPI/internal/domain/port/mocks"
	pkgerrs "MeetingRoomsAPI/pkg/errs"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLoginUC_Execute(t *testing.T) {
	type adapter struct {
		user     *mocks.MockUserRepository
		password *mocks.MockPasswordHasher
		token    *mocks.MockTokenGenerator
	}

	type testCase struct {
		name          string
		input         dto.LoginInput
		mockBehaviour func(a adapter)
		expectErr     error
	}

	var tests = []testCase{
		{
			name: "Success",
			input: dto.LoginInput{
				Email:    "test@avito.ru",
				Password: "correct_password",
			},
			mockBehaviour: func(a adapter) {
				u, _ := model.NewUser("test@avito.ru", "hashed_pass", "user")
				a.user.EXPECT().GetByEmail(mock.Anything, "test@avito.ru").Return(u, nil)
				a.password.EXPECT().Compare("hashed_pass", "correct_password").Return(true)
				a.token.EXPECT().GenerateToken(u.ID(), u.Role().String()).Return("jwt_token", nil)
			},
			expectErr: nil,
		},
		{
			name: "Failure - user not found",
			input: dto.LoginInput{
				Email:    "notfound@avito.ru",
				Password: "any",
			},
			mockBehaviour: func(a adapter) {
				a.user.EXPECT().GetByEmail(mock.Anything, "notfound@avito.ru").
					Return(nil, pkgerrs.ErrObjectNotFound)
			},
			expectErr: ucerrs.ErrInvalidCredentials,
		},
		{
			name: "Failure - wrong password",
			input: dto.LoginInput{
				Email:    "test@avito.ru",
				Password: "wrong_password",
			},
			mockBehaviour: func(a adapter) {
				u, _ := model.NewUser("test@avito.ru", "hashed_pass", "user")
				a.user.EXPECT().GetByEmail(mock.Anything, "test@avito.ru").Return(u, nil)
				a.password.EXPECT().Compare("hashed_pass", "wrong_password").Return(false)
			},
			expectErr: ucerrs.ErrInvalidCredentials,
		},
		{
			name: "Failure - token generation error",
			input: dto.LoginInput{
				Email:    "test@avito.ru",
				Password: "password",
			},
			mockBehaviour: func(a adapter) {
				u, _ := model.NewUser("test@avito.ru", "hashed_pass", "user")
				a.user.EXPECT().GetByEmail(mock.Anything, "test@avito.ru").Return(u, nil)
				a.password.EXPECT().Compare("hashed_pass", "password").Return(true)
				a.token.EXPECT().GenerateToken(u.ID(), u.Role().String()).
					Return("", errors.New("internal crypto error"))
			},
			expectErr: ucerrs.ErrGenerateToken,
		},
		{
			name: "Failure - repository error",
			input: dto.LoginInput{
				Email:    "test@avito.ru",
				Password: "password",
			},
			mockBehaviour: func(a adapter) {
				a.user.EXPECT().GetByEmail(mock.Anything, "test@avito.ru").
					Return(nil, errors.New("db connection lost"))
			},
			expectErr: ucerrs.ErrGetUserByEmailDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := mocks.NewMockUserRepository(t)
			passHasher := mocks.NewMockPasswordHasher(t)
			tokenGen := mocks.NewMockTokenGenerator(t)

			tt.mockBehaviour(adapter{
				user:     userRepo,
				password: passHasher,
				token:    tokenGen,
			})

			uc := usecase.NewLoginUC(userRepo, passHasher, tokenGen)

			out, err := uc.Execute(context.Background(), tt.input)

			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "jwt_token", out.Token)
			}
		})
	}
}
