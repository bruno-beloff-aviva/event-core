package snsmanager

// https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/go_sns_code_examples.html

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/joerdav/zapray"
	"go.uber.org/zap"
)

type SNSManager struct {
	logger    *zapray.Logger
	snsClient *sns.Client
}

func NewSNSManager(logger *zapray.Logger, cfg aws.Config) SNSManager {
	snsClient := sns.NewFromConfig(cfg)

	return SNSManager{logger: logger, snsClient: snsClient}
}

func (m SNSManager) Pub(ctx context.Context, topicArn string, message string) error {
	m.logger.Debug("Pub", zap.String("topicArn", topicArn))

	publishInput := sns.PublishInput{TopicArn: aws.String(topicArn), Message: aws.String(message)}

	_, err := m.snsClient.Publish(ctx, &publishInput)
	if err != nil {
		m.logger.Error("Couldn't publish message", zap.String("topicArn", topicArn), zap.Error(err))
	}

	return err
}
