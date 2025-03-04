package singleshot

import (
	"context"

	"github.com/bruno-beloff-aviva/event-core/services"

	"github.com/joerdav/zapray"
	"go.uber.org/zap"
)

type SingleshotHandler[T any] interface {
	UniqueID(event T) (policyOrQuoteID string, eventID string, err error)
	Process(ctx context.Context, event T) (err error)
}

type SingleshotGateway[T any] struct {
	logger                *zapray.Logger
	handler               SingleshotHandler[T]
	eventHasBeenProcessed services.EventHasBeenProcessedFunc
	markEventAsProcessed  services.MarkEventAsProcessedFunc
}

func NewSingleshotGateway[T any](logger *zapray.Logger, handler SingleshotHandler[T], eventHasBeenProcessed services.EventHasBeenProcessedFunc, markEventAsProcessed services.MarkEventAsProcessedFunc) SingleshotGateway[T] {
	return SingleshotGateway[T]{
		logger:                logger,
		handler:               handler,
		eventHasBeenProcessed: eventHasBeenProcessed,
		markEventAsProcessed:  markEventAsProcessed,
	}
}

func (g SingleshotGateway[T]) ProcessOnce(ctx context.Context, event T) (bool, error) {
	g.logger.Debug("ProcessOnce: ", zap.Any("event", event))

	// Check...
	policyOrQuoteID, eventID, err := g.handler.UniqueID(event)
	if err != nil {
		g.logger.Error("Error getting UniqueID", zap.Error(err))
		return false, err
	}

	eventHasBeenProcessed, err := g.eventHasBeenProcessed(ctx, policyOrQuoteID, eventID)
	if err != nil {
		g.logger.Error("Error checking if event has been processed", zap.Error(err))
		return false, err
	}

	if eventHasBeenProcessed {
		g.logger.Info("Event has already been processed")
		return false, nil
	}

	err = g.handler.Process(ctx, event)
	if err != nil {
		g.logger.Error("Process error", zap.Error(err))
		return true, err
	}

	// Mark as processed...
	err = g.markEventAsProcessed(ctx, policyOrQuoteID, eventID)
	if err != nil {
		g.logger.Error("Error marking event as processed", zap.Error(err))
		return true, nil
	}

	return true, nil
}
