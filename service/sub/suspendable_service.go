package sub

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sqstest/manager/dbmanager"
	"sqstest/service/testmessage"
	"sqstest/service/testreception"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/joerdav/zapray"
	"go.uber.org/zap"
)

const sleepSeconds = 20

type SuspendableService struct {
	logger    *zapray.Logger
	dbManager dbmanager.DynamoManager
	id        string
}

func NewSuspendableService(logger *zapray.Logger, cfg aws.Config, dbManager dbmanager.DynamoManager, id string) SuspendableService {
	return SuspendableService{logger: logger, dbManager: dbManager, id: id}
}

func (m SuspendableService) Handle(ctx context.Context, record events.SQSMessage) (err error) {
	m.logger.Debug("Handle", zap.String("record body", record.Body))

	var message testmessage.TestMessage
	var reception testreception.TestReception

	suspended := os.Getenv("SUSPENDED") == "true"

	err = json.Unmarshal([]byte(record.Body), &message)
	if err != nil {
		return err
	}

	m.logger.Debug("Receive: ", zap.Any("message", message))

	if suspended && !strings.Contains(message.Path, "resume") {
		m.logger.Warn("SUSPENDED", zap.Any("Path", message.Path))
		return errors.New("Suspended")
	}

	err = m.stateChange(message.Path)
	if err != nil {
		return err
	}

	// dbManager.Put...
	reception = testreception.NewTestReception(m.id, message)
	m.logger.Info("Receive: ", zap.Any("reception", reception))

	err = m.dbManager.Put(ctx, &reception)
	if err != nil {
		m.logger.Error("Receive: ", zap.Error(err))
	}

	return nil
}

func (m SuspendableService) stateChange(path string) (err error) {
	switch {
	case strings.Contains(path, "suspend"):
		m.logger.Warn("DO SUSPEND", zap.Any("Path", path))
		os.Setenv("SUSPENDED", "true")
		return nil

	case strings.Contains(path, "resume"):
		m.logger.Warn("DO RESUME", zap.Any("Path", path))
		os.Setenv("SUSPENDED", "false")
		return nil

	case strings.Contains(path, "sleep"):
		m.logger.Warn("DO SLEEP", zap.Any("Path", path))
		time.Sleep(sleepSeconds * time.Second)
		return nil

	case strings.Contains(path, "error"):
		m.logger.Warn("DO ERROR", zap.Any("Path", path))
		return errors.New(path)

	case strings.Contains(path, "panic"):
		m.logger.Warn("DO PANIC", zap.Any("Path", path))
		panic(path)

	default:
		m.logger.Warn("DO OK", zap.Any("Path", path))
		return nil
	}
}
