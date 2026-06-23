package conference

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	ErrServiceUnavailable = errors.New("conference service: connection timeout")
	ErrInternalError      = errors.New("conference service: internal error")
)

type ConferenceService struct {
	mode string
}

func NewConferenceService(mode string) *ConferenceService {
	return &ConferenceService{mode: mode}
}

func (c *ConferenceService) CreateMeeting(_ context.Context) (string, error) {
	switch c.mode {
	case "timeout":
		return "", ErrServiceUnavailable
	case "500":
		return "", ErrInternalError
	default:
		return fmt.Sprintf(
			"https://telemost.yandex.ru/%s",
			uuid.New().String(),
		), nil
	}
}

func (c *ConferenceService) SetMode(mode string) {
	c.mode = mode
}
