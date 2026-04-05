package usecase

import (
	"MeetingRoomsAPI/internal/app/dto"
	ucerrs "MeetingRoomsAPI/internal/app/errs"
	"MeetingRoomsAPI/internal/app/mapper"
	"MeetingRoomsAPI/internal/domain/model"
	"MeetingRoomsAPI/internal/domain/port"
	pkgerrs "MeetingRoomsAPI/pkg/errs"
	"context"
	"errors"
	"time"

	"github.com/avito-tech/go-transaction-manager/trm/v2"
)

type CreateBookingUC struct {
	trManager  trm.Manager
	slot       port.SlotRepository
	booking    port.BookingRepository
	conference port.ConferenceService
}

func NewCreateBookingUC(
	trManager trm.Manager,
	slot port.SlotRepository,
	booking port.BookingRepository,
	conference port.ConferenceService,
) *CreateBookingUC {
	return &CreateBookingUC{
		trManager:  trManager,
		slot:       slot,
		booking:    booking,
		conference: conference,
	}
}

func (uc *CreateBookingUC) Execute(ctx context.Context, in dto.CreateBookingInput) (dto.CreateBookingOutput, error) {
	var out dto.CreateBookingOutput

	err := uc.trManager.Do(ctx, func(ctx context.Context) error {
		slot, err := uc.slot.Get(ctx, in.SlotID)
		if err != nil {
			if errors.Is(err, pkgerrs.ErrObjectNotFound) {
				return ucerrs.ErrSlotNotFound
			}
			return ucerrs.Wrap(ucerrs.ErrGetSlotDB, err)
		}

		if slot.Start().Before(time.Now().UTC()) {
			return ucerrs.ErrCannotCreateBooking
		}

		var conferenceLink *string
		if in.CreateConferenceLink {
			link, err := uc.conference.CreateMeeting(ctx)
			if err != nil {
				return ucerrs.Wrap(ucerrs.ErrCreateMeeting, err)
			}
			conferenceLink = &link
		}

		booking, err := model.NewBooking(
			slot.ID(),
			in.UserID,
			conferenceLink,
		)
		if err != nil {
			return ucerrs.Wrap(ucerrs.ErrInvalidInput, err)
		}

		createdBooking, err := uc.booking.Create(ctx, booking)
		if err != nil {
			return ucerrs.Wrap(ucerrs.ErrCreateBookingDB, err)
		}

		out = mapper.MapDomainToCreateBookingDTO(createdBooking)

		return nil
	})

	if err != nil {
		return dto.CreateBookingOutput{}, nil
	}

	return out, nil
}
