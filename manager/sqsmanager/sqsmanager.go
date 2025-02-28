package sqsmanager

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/joerdav/zapray"
	"go.uber.org/zap"
)

type SQSManager struct {
	logger    *zapray.Logger
	sqsClient *sqs.Client
}

func NewSQSManager(logger *zapray.Logger, cfg aws.Config) SQSManager {
	sqsClient := sqs.NewFromConfig(cfg)

	return SQSManager{logger: logger, sqsClient: sqsClient}
}

func (m SQSManager) Pub(ctx context.Context, queueUrl string, message string) error {
	m.logger.Debug("Pub", zap.String("queueUrl", queueUrl))

	m.sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
		MessageBody: &message,
		QueueUrl:    &queueUrl,
	})

	return nil
}
