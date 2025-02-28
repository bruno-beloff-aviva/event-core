package pub

import (
	"context"
	"encoding/json"
	"sqstest/service/testmessage"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/joerdav/zapray"
	"go.uber.org/zap"
)

type EventBridgePubService struct {
	logger       *zapray.Logger
	client       eventbridge.Client
	eventSource  string
	eventBusName string
}

func NewEventBridgePubService(logger *zapray.Logger, cfg aws.Config, eventSource string, eventBusName string) EventBridgePubService {
	client := eventbridge.NewFromConfig(cfg)

	return EventBridgePubService{logger: logger, client: *client, eventSource: eventSource, eventBusName: eventBusName}
}

func (s EventBridgePubService) Publish(ctx context.Context, clientId string, path string) (testmessage.TestMessage, error) {
	s.logger.Debug("Publish", zap.String("clientId", clientId))

	message := testmessage.NewTestMessage(clientId, path)
	jmsg, err := json.Marshal(message)

	if err != nil {
		panic(err)
	}

	resp, err := s.client.PutEvents(ctx, &eventbridge.PutEventsInput{
		Entries: []types.PutEventsRequestEntry{
			{
				Detail:       aws.String(string(jmsg)),
				DetailType:   aws.String("TestMessage"),
				Source:       aws.String(s.eventSource),
				EventBusName: aws.String(s.eventBusName),
			},
		},
	})

	if err != nil {
		panic(err)
	}

	s.logger.Info("Publish", zap.Any("message", message), zap.Any("response", resp))

	return message, nil
}
