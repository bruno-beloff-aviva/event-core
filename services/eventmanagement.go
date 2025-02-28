package services

import "context"

type EventHasBeenProcessedFunc func(ctx context.Context, policyOrQuoteID string, eventID string) (bool, error)
type MarkEventAsProcessedFunc func(ctx context.Context, policyID string, eventID string) error

func NullEventHasBeenProcessed(ctx context.Context, policyOrQuoteID string, eventID string) (bool, error) {
	return false, nil
}

func NullMarkEventAsProcessed(ctx context.Context, policyID string, eventID string) error {
	return nil
}
