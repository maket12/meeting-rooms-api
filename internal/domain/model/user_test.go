package model_test

import (
	"MeetingRoomsAPI/internal/domain/model"
	pkgerrs "MeetingRoomsAPI/pkg/errs"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUser(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name         string
		email        string
		passwordHash string
		role         model.UserRole
		expect       error
	}

	var (
		testEmail    = "super-test@avito.ru"
		testPassword = "hashed-password"
		testRole     = model.UserRole("admin")
	)

	var testCases = []testCase{
		{
			name:         "Success",
			email:        testEmail,
			passwordHash: testPassword,
			role:         testRole,
			expect:       nil,
		},
		{
			name:   "Failure - email not specified",
			email:  "",
			expect: pkgerrs.ErrValueIsRequired,
		},
		{
			name:         "Failure - password not specified",
			email:        testEmail,
			passwordHash: "",
			expect:       pkgerrs.ErrValueIsRequired,
		},
		{
			name:         "Failure - role not specified",
			email:        testEmail,
			passwordHash: testPassword,
			role:         "",
			expect:       pkgerrs.ErrValueIsRequired,
		},
		{
			name:         "Failure - invalid role value",
			email:        testEmail,
			passwordHash: testPassword,
			role:         "stranger",
			expect:       pkgerrs.ErrValueIsInvalid,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			user, err := model.NewUser(
				tt.email,
				tt.passwordHash,
				tt.role,
			)
			if tt.expect == nil {
				require.NoError(t, err)
				require.NotNil(t, user)

				assert.True(t, user.ID() != uuid.Nil)
				assert.Equal(t, tt.email, user.Email())
				assert.Equal(t, tt.passwordHash, user.PasswordHash())
				assert.Equal(t, tt.role, user.Role())
				assert.True(t, user.Role().String() == string(testRole))
			} else {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expect)
				assert.Nil(t, user)
			}
		})
	}
}
