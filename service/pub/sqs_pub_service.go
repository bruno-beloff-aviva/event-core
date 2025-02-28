package pub

import (
	"context"
	"encoding/json"
	"sqstest/manager/sqsmanager"
	"sqstest/service/testmessage"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/joerdav/zapray"
	"go.uber.org/zap"
)

type SQSPubService struct {
	logger     *zapray.Logger
	sqsManager sqsmanager.SQSManager
	queueUrl   string
}

func NewSQSPubService(logger *zapray.Logger, cfg aws.Config, queueUrl string) SQSPubService {
	sqsManager := sqsmanager.NewSQSManager(logger, cfg)

	return SQSPubService{logger: logger, sqsManager: sqsManager, queueUrl: queueUrl}
}

func (s SQSPubService) Publish(ctx context.Context, clientId string, path string) (testmessage.TestMessage, error) {
	s.logger.Debug("Publish", zap.String("clientId", clientId))

	message := testmessage.NewTestMessage(clientId, path)

	jmsg, err := json.Marshal(message)
	strmsg := string(jmsg)

	if err != nil {
		panic(err)
	}

	s.logger.Info("Publish", zap.Any("message", message))

	return message, s.sqsManager.Pub(ctx, s.queueUrl, strmsg)
}
