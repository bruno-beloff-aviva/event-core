package sub

import (
	"context"
	"encoding/json"
	"sqstest/lambda/handler/singleshot"
	"sqstest/manager/dbmanager"
	"sqstest/service/testmessage"
	"sqstest/service/testreception"
	"sqstest/services"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/joerdav/zapray"
	"go.uber.org/zap"
)

type ContinuousService struct {
	gateway   singleshot.SingleshotGateway[testmessage.TestMessage]
	logger    *zapray.Logger
	dbManager dbmanager.DynamoManager
	id        string
}

func NewContinuousService(logger *zapray.Logger, cfg aws.Config, dbManager dbmanager.DynamoManager, id string) ContinuousService {
	handler := ContinuousService{logger: logger, dbManager: dbManager, id: id}
	handler.newGateway(services.NullEventHasBeenProcessed, services.NullMarkEventAsProcessed)

	return handler
}

func (s *ContinuousService) newGateway(eventHasBeenProcessed services.EventHasBeenProcessedFunc, markEventAsProcessed services.MarkEventAsProcessedFunc) {
	s.gateway = singleshot.NewSingleshotGateway(s.logger, s, eventHasBeenProcessed, markEventAsProcessed)
}

func (s ContinuousService) Handle(ctx context.Context, record events.SQSMessage) (err error) {
	s.logger.Info("Handle", zap.String("record body", record.Body))

	var message testmessage.TestMessage

	err = json.Unmarshal([]byte(record.Body), &message)
	if err != nil {
		return err
	}

	processPerformed, err := s.gateway.ProcessOnce(ctx, message)

	if processPerformed && err == nil {
		s.logger.Info("process done - metrics here")
	}

	return err
}

func (s ContinuousService) UniqueID(event testmessage.TestMessage) (policyOrQuoteID string, eventID string, err error) {
	return event.Client, event.Sent, nil
}

func (s ContinuousService) Process(ctx context.Context, event testmessage.TestMessage) (err error) {
	s.logger.Debug("Process: ", zap.Any("event", event))

	// dbManager.Put...
	reception := testreception.NewTestReception(s.id, event)
	s.logger.Info("Process: ", zap.Any("reception", reception))

	err = s.dbManager.Put(ctx, &reception)
	if err != nil {
		s.logger.Error("Put: ", zap.Error(err))
	}

	return err
}
