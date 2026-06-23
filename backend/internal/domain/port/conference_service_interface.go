package port

import "context"

type ConferenceService interface {
	CreateMeeting(_ context.Context) (string, error)
	SetMode(mode string)
}
