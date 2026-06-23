package usecase

import (
	"backend/internal/app/dto"
	"backend/internal/app/errs"
	"backend/internal/app/mapper"
	"backend/internal/domain/model"
	port2 "backend/internal/domain/port"
	pkgerrs "backend/pkg/errs"
	"context"
	"errors"
	"time"

	"github.com/avito-tech/go-transaction-manager/trm/v2"
)

type CreateBookingUC struct {
	trManager  trm.Manager
	slot       port2.SlotRepository
	booking    port2.BookingRepository
	conference port2.ConferenceService
}

func NewCreateBookingUC(
	trManager trm.Manager,
	slot port2.SlotRepository,
	booking port2.BookingRepository,
	conference port2.ConferenceService,
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
		slot, getErr := uc.slot.Get(ctx, in.SlotID)
		if getErr != nil {
			if errors.Is(getErr, pkgerrs.ErrObjectNotFound) {
				return errs.ErrSlotNotFound
			}
			return errs.Wrap(errs.ErrGetSlotDB, getErr)
		}

		if slot.Start().Before(time.Now().UTC()) {
			return errs.ErrCannotCreateBooking
		}

		var conferenceLink *string
		if in.CreateConferenceLink {
			link, createErr := uc.conference.CreateMeeting(ctx)
			if createErr != nil {
				return errs.Wrap(errs.ErrCreateMeeting, createErr)
			}
			conferenceLink = &link
		}

		booking, createErr := model.NewBooking(
			slot.ID(),
			in.UserID,
			conferenceLink,
		)
		if createErr != nil {
			return errs.Wrap(errs.ErrInvalidInput, createErr)
		}

		createdBooking, createErr := uc.booking.Create(ctx, booking)
		if createErr != nil {
			return errs.Wrap(errs.ErrCreateBookingDB, createErr)
		}

		out = mapper.MapDomainToCreateBookingDTO(createdBooking)

		return nil
	})

	if err != nil {
		return dto.CreateBookingOutput{}, err
	}

	return out, nil
}
