package usecase_test

import (
	"backend/internal/app/dto"
	ucerrs "backend/internal/app/errs"
	"backend/internal/app/usecase"
	"backend/internal/domain/model"
	"backend/internal/domain/port/mocks"
	pkgerrs "backend/pkg/errs"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDummyLoginUC_Execute(t *testing.T) {
	type adapter struct {
		user   *mocks.MockUserRepository
		jwtGen *mocks.MockTokenGenerator
	}

	adminID := uuid.New()
	userID := uuid.New()

	type testCase struct {
		name          string
		input         dto.DummyLoginInput
		mockBehaviour func(a adapter)
		expectErr     error
	}

	var tests = []testCase{
		{
			name:  "Success - Admin Login",
			input: dto.DummyLoginInput{Role: "admin"},
			mockBehaviour: func(a adapter) {
				u := model.RestoreUser(adminID, "admin@avito.ru", "hash", "admin", time.Now())
				a.user.EXPECT().GetByID(mock.Anything, adminID).Return(u, nil)

				a.jwtGen.EXPECT().GenerateToken(u.ID(), "admin").Return("admin_token", nil)
			},
			expectErr: nil,
		},
		{
			name:  "Success - User Login",
			input: dto.DummyLoginInput{Role: "user"},
			mockBehaviour: func(a adapter) {
				u := model.RestoreUser(userID, "user@avito.ru", "hash", "user", time.Now())
				a.user.EXPECT().GetByID(mock.Anything, userID).Return(u, nil)

				a.jwtGen.EXPECT().GenerateToken(u.ID(), "user").Return("user_token", nil)
			},
			expectErr: nil,
		},
		{
			name:          "Failure - Invalid Role",
			input:         dto.DummyLoginInput{Role: "super-admin"},
			mockBehaviour: func(a adapter) {},
			expectErr:     ucerrs.ErrInvalidInput,
		},
		{
			name:  "Failure - User Not Found",
			input: dto.DummyLoginInput{Role: "admin"},
			mockBehaviour: func(a adapter) {
				a.user.EXPECT().GetByID(mock.Anything, adminID).Return(nil, pkgerrs.ErrObjectNotFound)
			},
			expectErr: ucerrs.ErrUserNotFound,
		},
		{
			name:  "Failure - DB Error",
			input: dto.DummyLoginInput{Role: "user"},
			mockBehaviour: func(a adapter) {
				a.user.EXPECT().GetByID(mock.Anything, userID).Return(nil, errors.New("db crash"))
			},
			expectErr: ucerrs.ErrGetUserByIDDB,
		},
		{
			name:  "Failure - Token Generation Error",
			input: dto.DummyLoginInput{Role: "admin"},
			mockBehaviour: func(a adapter) {
				u := model.RestoreUser(adminID, "admin@avito.ru", "hash", "admin", time.Now())
				a.user.EXPECT().GetByID(mock.Anything, adminID).Return(u, nil)
				a.jwtGen.EXPECT().GenerateToken(u.ID(), "admin").Return("", errors.New("token fail"))
			},
			expectErr: ucerrs.ErrGenerateToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Init Mocks
			userRepo := mocks.NewMockUserRepository(t)
			jwtGen := mocks.NewMockTokenGenerator(t)

			tt.mockBehaviour(adapter{user: userRepo, jwtGen: jwtGen})

			uc := usecase.NewDummyLoginUC(userRepo, jwtGen, adminID, userID)

			out, err := uc.Execute(context.Background(), tt.input)

			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErr)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, out.Token)

				if tt.input.Role == "admin" {
					assert.Equal(t, "admin_token", out.Token)
				}
			}
		})
	}
}
