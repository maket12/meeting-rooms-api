package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/maket12/meeting-rooms-api/internal/app/dto"
	ucerrs "github.com/maket12/meeting-rooms-api/internal/app/errs"
	"github.com/maket12/meeting-rooms-api/internal/app/mapper"
	"github.com/maket12/meeting-rooms-api/internal/domain/model"
	"github.com/maket12/meeting-rooms-api/internal/domain/port"
	pkgerrs "github.com/maket12/meeting-rooms-api/pkg/errs"

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

	err := uc.trManager.Do(ctx, func(txCtx context.Context) error {
		slot, getErr := uc.slot.Get(txCtx, in.SlotID)
		if getErr != nil {
			if errors.Is(getErr, pkgerrs.ErrObjectNotFound) {
				return ucerrs.ErrSlotNotFound
			}
			return ucerrs.Wrap(ucerrs.ErrGetSlotDB, getErr)
		}

		if slot.Start().Before(time.Now().UTC()) {
			return ucerrs.ErrCannotCreateBooking
		}

		var conferenceLink *string
		if in.CreateConferenceLink {
			link, createErr := uc.conference.CreateMeeting(txCtx)
			if createErr != nil {
				return ucerrs.Wrap(ucerrs.ErrCreateMeeting, createErr)
			}
			conferenceLink = &link
		}

		booking, createErr := model.NewBooking(
			slot.ID(),
			in.UserID,
			conferenceLink,
		)
		if createErr != nil {
			return ucerrs.Wrap(ucerrs.ErrInvalidInput, createErr)
		}

		createdBooking, createErr := uc.booking.Create(txCtx, booking)
		if createErr != nil {
			if errors.Is(createErr, pkgerrs.ErrObjectAlreadyExists) {
				return ucerrs.ErrBookingAlreadyExists
			}
			return ucerrs.Wrap(ucerrs.ErrCreateBookingDB, createErr)
		}

		out = mapper.MapDomainToCreateBookingDTO(createdBooking)

		return nil
	})

	if err != nil {
		return dto.CreateBookingOutput{}, err
	}

	return out, nil
}
