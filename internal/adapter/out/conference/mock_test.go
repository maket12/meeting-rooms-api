package conference_test

import (
	"MeetingRoomsAPI/internal/adapter/out/conference"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConferenceMock_CreateMeeting(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mock := conference.NewConferenceService("available")

		link, err := mock.CreateMeeting(ctx)

		require.NoError(t, err)
		assert.True(t, strings.HasPrefix(link, "https://telemost.yandex.ru/"))

		assert.Greater(t, len(link), 30) // Valid UUID
	})

	t.Run("Failure - connection timeout", func(t *testing.T) {
		mock := conference.NewConferenceService("timeout")

		link, err := mock.CreateMeeting(ctx)

		assert.Empty(t, link)
		assert.ErrorIs(t, err, conference.ErrServiceUnavailable)
	})

	t.Run("Failure - internal server error", func(t *testing.T) {
		mock := conference.NewConferenceService("500")

		link, err := mock.CreateMeeting(ctx)

		assert.Empty(t, link)
		assert.ErrorIs(t, err, conference.ErrInternalError)
	})

	t.Run("Dynamic mode switching", func(t *testing.T) {
		mock := conference.NewConferenceService("available")

		// Success at first
		_, err := mock.CreateMeeting(ctx)
		assert.NoError(t, err)

		// Then connection timeout
		mock.SetMode("timeout")
		_, err = mock.CreateMeeting(ctx)
		assert.ErrorIs(t, err, conference.ErrServiceUnavailable)
	})
}
